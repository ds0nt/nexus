package nexus

import (
	"io"
	"sync"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type Nexus struct {
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
				// logrus.Printf("sending message %v to %v", msg, client)
				err := websocket.JSON.Send(ws, msg)
				if err != nil {
					logrus.Debugf("error writing to websocket %v", msg)
				}
			case <-client.context.Done():
				logrus.Debugf("stopping websocket write loop due to context closed")
				return
			}
		}
	}()

	// read loop
	for {
		select {
		case <-client.context.Done():
			logrus.Debugf("stopping websocket read loop due to context closed")
			return
		default:
			p := Packet{}
			err := websocket.JSON.Receive(ws, &p)
			if err != nil {
				if err == io.EOF {
					// remove connection
					logrus.Println(err)
					return
				}
				logrus.Println(err)
				continue
			}
			// logrus.Printf("received message %v from %v", p, client)

			handler, ok := n.handlers[p.Type]
			if !ok {
				logrus.Debugf("handler %s does not exist", p.Type, client.name)
				continue
			} else {
				handler(client, &p)
			}
		}
	}
}
