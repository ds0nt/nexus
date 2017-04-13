package nexus

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Nexus struct {
	Verbose    bool
	all        *Pool
	handlersMu sync.Mutex
	handlers   map[string]Handler
}

func NewNexus() *Nexus {
	n := Nexus{
		all:      NewPool(),
		handlers: map[string]Handler{},
	}
	return &n
}

func (n *Nexus) debugf(format string, args ...interface{}) {
	if n.Verbose {
		fmt.Printf(format, args...)
	}
}

func (n *Nexus) debug(args ...interface{}) {
	if n.Verbose {
		fmt.Println(args...)
	}
}

func (n *Nexus) errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (n *Nexus) All() *Pool {
	return n.all
}

func (n *Nexus) Handle(t string, handler Handler) {
	n.handlersMu.Lock()
	defer n.handlersMu.Unlock()

	n.handlers[t] = handler
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (n *Nexus) Handler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(ws)
	defer client.close()

	n.all.Add(client)
	defer n.all.Remove(client)

	// write loop
	go func() {
		for {
			select {
			case msg := <-client.messageChan:
				n.debugf("sending message %v to %v", msg, client)
				err := ws.WriteJSON(msg)
				if err != nil {
					n.errorf("error writing to websocket %v", msg)
				}
			case <-client.context.Done():
				n.debugf("stopping websocket write loop due to context closed")
				return
			}
		}
	}()

	// read loop
	for {
		select {
		case <-client.context.Done():
			n.debugf("stopping websocket read loop due to context closed")
			return
		default:
			p := Packet{}
			_, data, err := ws.ReadMessage()
			if err != nil {
				if err == io.EOF {
					// remove connection
					n.errorf("got eof error on read", err)
					return
				}
				n.errorf("got websocket error on receive", err)
				return
			}
			err = json.Unmarshal(data, &p)
			if err != nil {
				n.errorf("received malformed json", err, p)
				continue
			}
			n.debugf("received message %v from %v", p, client)

			handler, ok := n.handlers[p.Type]
			if !ok {
				n.debugf("handler %s does not exist", p.Type, client.name)
				continue
			} else {
				go handler(client, &p)
			}
		}
	}
}
