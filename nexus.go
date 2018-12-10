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

	"context"

	"github.com/gorilla/websocket"
)

type PacketFormat int

const (
	PacketFormatJSON = iota
	PacketFormatDelimited
)

type Nexus struct {
	Verbose         bool
	all             *Pool
	handlersMu      sync.Mutex
	handlers        map[string]Handler
	streamers       map[string]Streamer
	PacketFormat    PacketFormat
	streamCancels   map[string]context.CancelFunc
	streamCancelsMu sync.Mutex
}

func NewNexus() *Nexus {
	n := Nexus{
		all:           NewPool(),
		handlers:      map[string]Handler{},
		streamers:     map[string]Streamer{},
		PacketFormat:  PacketFormatJSON,
		streamCancels: map[string]context.CancelFunc{},
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

func (n *Nexus) StreamHandle(t string, streamer Streamer) {
	n.handlersMu.Lock()
	defer n.handlersMu.Unlock()

	n.streamers[t] = streamer
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (n *Nexus) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.Handler(w, r)
}

func (n *Nexus) Handler(w http.ResponseWriter, r *http.Request) {
	ws, err := Upgrader.Upgrade(w, r, nil)
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

			p := &Packet{}
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
			n.debugf("received message %v from %v", p, client)

			// kill handler by stream id
			if strings.HasPrefix(p.Type, "-") {
				if len(p.StreamID) == 0 {
					n.errorf("cannot kill stream with no stream id %s", p.StreamID)
					continue
				}
				cancel, ok := n.streamCancels[p.StreamID]
				if ok {
					cancel()
				}
				continue
			}

			// handle normal
			handler, ok := n.handlers[p.Type]
			if ok {
				go handler(client, p)
				continue
			}

			// handle stream (cancelable by kill packet)
			streamer, ok := n.streamers[p.Type]
			if ok {
				if len(p.StreamID) > 0 {
					go n.handleWithCancel(streamer, client, p)
				} else {
					n.errorf("cannot start stream with no stream id %s", p.StreamID)
				}
				continue
			}

			n.debugf("handler %s does not exist", p.Type)
			continue
		}
	}
}

type Context struct {
	Client        *Client
	Packet        *Packet
	StreamContext context.Context
}

func (n *Nexus) handleWithCancel(handler Streamer, client *Client, p *Packet) {
	ctx, cancel := context.WithCancel(client.Context)
	c := &Context{
		client,
		p,
		ctx,
	}
	n.streamCancelsMu.Lock()
	n.streamCancels[p.StreamID] = cancel
	n.streamCancelsMu.Unlock()
	n.debugf("registered stream cancel %s %s", p.Type, p.StreamID)
	handler(c)

	n.streamCancelsMu.Lock()
	delete(n.streamCancels, p.StreamID)
	n.streamCancelsMu.Unlock()
	n.debugf("deleted stream cancel %s %s", p.Type, p.StreamID)
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
