package nexus

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
)

type SendHandler func(*Packet)

func (p *Client) BeforeSend(handler SendHandler) {
	p.sendHandlerMu.Lock()
	p.sendHandlers = append(p.sendHandlers, handler)
	p.sendHandlerMu.Unlock()
}

type Client struct {
	name            string
	closed          bool
	Context         context.Context
	cancel          context.CancelFunc
	messageChan     chan *Packet
	Env             map[interface{}]interface{}
	sendCloseMu     sync.Mutex
	sendHandlers    []SendHandler
	sendHandlerMu   sync.Mutex
	Conn            *websocket.Conn
	streamCancels   map[string]context.CancelFunc
	streamCancelsMu sync.Mutex
}

func newClient(conn *websocket.Conn) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		Conn:          conn,
		sendCloseMu:   sync.Mutex{},
		name:          conn.RemoteAddr().String(),
		Context:       ctx,
		cancel:        cancel,
		messageChan:   make(chan *Packet, 100),
		Env:           make(map[interface{}]interface{}),
		streamCancels: map[string]context.CancelFunc{},
	}
}

func (c *Client) close() {
	c.sendCloseMu.Lock()
	defer c.sendCloseMu.Unlock()
	c.closed = true
	c.cancel()
	close(c.messageChan)
}

func (c *Client) Send(p *Packet) {
	for _, h := range c.sendHandlers {
		h(p)
	}
	if c.closed {
		return
	}
	c.sendCloseMu.Lock()
	defer c.sendCloseMu.Unlock()
	if c.closed {
		return
	}
	c.messageChan <- p
}

func (c *Client) String() string {
	return c.name
}
