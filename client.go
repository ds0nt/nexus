package nexus

import (
	"context"

	"github.com/gorilla/websocket"
)

type Client struct {
	name        string
	context     context.Context
	cancel      context.CancelFunc
	messageChan chan *Packet
	Env         map[interface{}]interface{}
}

func newClient(conn *websocket.Conn) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		name:        conn.RemoteAddr().String(),
		context:     ctx,
		cancel:      cancel,
		messageChan: make(chan *Packet, 100),
		Env:         make(map[interface{}]interface{}),
	}
}

func (c *Client) close() {
	c.cancel()
	close(c.messageChan)
}

func (c *Client) Send(p *Packet) {
	select {
	case <-c.context.Done():
	case c.messageChan <- p:
	}
}

func (c *Client) String() string {
	return c.name
}
