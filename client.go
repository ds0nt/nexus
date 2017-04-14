package nexus

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	name        string
	closed      bool
	context     context.Context
	cancel      context.CancelFunc
	messageChan chan *Packet
	Env         map[interface{}]interface{}
	sendCloseMu sync.Mutex
}

func newClient(conn *websocket.Conn) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		sendCloseMu: sync.Mutex{},
		name:        conn.RemoteAddr().String(),
		context:     ctx,
		cancel:      cancel,
		messageChan: make(chan *Packet, 100),
		Env:         make(map[interface{}]interface{}),
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
