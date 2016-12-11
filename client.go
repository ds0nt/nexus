package nexus

import (
	"context"

	ws "golang.org/x/net/websocket"
)

type Client struct {
	name        string
	context     context.Context
	cancel      context.CancelFunc
	messageChan chan *Packet
}

func newClient(conn *ws.Conn) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		name:        conn.RemoteAddr().String(),
		context:     ctx,
		cancel:      cancel,
		messageChan: make(chan *Packet, 100),
	}
}

func (c *Client) close() {
	c.cancel()
	close(c.messageChan)
}

func (c *Client) String() string {
	return c.name
}
