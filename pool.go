package nexus

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

type Pool struct {
	Name        string
	clientMapMu sync.Mutex
	clientMap   map[*Client]bool
}

func NewPool() *Pool {
	return &Pool{
		clientMapMu: sync.Mutex{},
		clientMap:   map[*Client]bool{},
	}
}

func (p *Pool) Add(c *Client) {
	p.clientMapMu.Lock()
	defer p.clientMapMu.Unlock()

	p.clientMap[c] = true
}

func (p *Pool) Remove(c *Client) {
	p.clientMapMu.Lock()
	defer p.clientMapMu.Unlock()

	delete(p.clientMap, c)
}

func (p *Pool) Broadcast(packet *Packet) {
	logrus.Printf("broadcasting packet %v", packet)
	for c, _ := range p.clientMap {
		c.messageChan <- packet
	}
}

func (p *Pool) String() {

}
