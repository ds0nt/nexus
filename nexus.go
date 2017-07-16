package nexus

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/pkg/errors"

	"strings"

	"strconv"

	"github.com/gorilla/websocket"
)

type PacketFormat int

const (
	PacketFormatJSON = iota
	PacketFormatDelimited
)

type Nexus struct {
	Verbose      bool
	all          *Pool
	handlersMu   sync.Mutex
	handlers     map[string]Handler
	PacketFormat PacketFormat
}

func NewNexus() *Nexus {
	n := Nexus{
		all:          NewPool(),
		handlers:     map[string]Handler{},
		PacketFormat: PacketFormatJSON,
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

func (n *Nexus) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.Handler(w, r)
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
				if msg == nil {
					continue
				}
				n.debugf("sending message %v to %v", msg, client)

				switch n.PacketFormat {
				case PacketFormatJSON:
					err := ws.WriteJSON(msg)
					if err != nil {
						n.errorf("error writing to websocket %v", msg)
					}
				case PacketFormatDelimited:
					err := ws.WriteMessage(websocket.TextMessage, marshalDelimitedPacket(*msg))
					if err != nil {
						n.errorf("error writing to websocket %v", msg)
					}
				}
			case <-client.Context.Done():
				n.debugf("stopping websocket write loop due to context closed")
				return
			}
		}
	}()

	// read loop
	for {
		select {
		case <-client.Context.Done():
			n.debugf("stopping websocket read loop due to context closed")
			return
		default:
			p := &Packet{}
			_, data, err := ws.ReadMessage()
			if err != nil {
				if err == io.EOF {
					// remove connection
					n.errorf("got eof error on read %s", err.Error())
					return
				}
				n.errorf("got websocket error on receive %s", err.Error())
				return
			}
			switch n.PacketFormat {
			case PacketFormatJSON:
				err = json.Unmarshal(data, p)
				if err != nil {
					n.errorf("received malformed json %s %v", err.Error(), p)
					continue
				}
			case PacketFormatDelimited:
				p, err = unmarshalDelimitedPacket(data)
				if err != nil {
					n.errorf("received malformed delimited message %s %v", err.Error(), p)
					continue
				}
			}
			// n.debugf()

			fmt.Printf("received message %v from %v", p, client)
			handler, ok := n.handlers[p.Type]
			if !ok {
				n.debugf("handler %s does not exist", p.Type, client.name)
				continue
			} else {
				go handler(client, p)
			}
		}
	}
}

func unmarshalDelimitedPacket(bytes []byte) (*Packet, error) {
	str := string(bytes)
	fmt.Println(str)
	peices := strings.SplitN(str, ":", 2)
	if len(peices) != 2 {
		return nil, errors.New("delimiter not found in message bytes")
	}

	n, err := strconv.Atoi(peices[0])
	if err != nil {
		return nil, errors.Errorf("first field should be an integer, but was %s", peices[0])
	}
	if len(peices[1]) < n {
		return nil, errors.Errorf("message ended unexpectedly, message length was not the expected size")
	}

	p := Packet{
		Type: peices[1][:n],
		Data: peices[1][n:],
	}
	return &p, nil
}

func marshalDelimitedPacket(p Packet) []byte {
	return []byte(fmt.Sprintf("%d:%s%s", len(p.Type), p.Type, p.Data))
}
