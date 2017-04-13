package nexus

import (
	"fmt"
	"io"
	"sync"

	"golang.org/x/net/websocket"
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

// Echo the data received on the WebSocket.
func (n *Nexus) Serve(ws *websocket.Conn) {

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
				err := websocket.JSON.Send(ws, msg)
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
			err := websocket.JSON.Receive(ws, &p)
			if err != nil {
				if err == io.EOF {
					// remove connection
					n.errorf("got eof error on read", err)
					return
				}
				n.errorf("got websocket error on receive", err)
				return
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
