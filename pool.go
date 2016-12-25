package nexus

import "sync"

type (
	JoinHandler  func(p *Pool, c *Client)
	LeaveHandler func(p *Pool, c *Client)
	Pool         struct {
		Name           string
		clientMapMu    sync.Mutex
		clientMap      map[*Client]bool
		joinHandlers   []JoinHandler
		joinMu         sync.Mutex
		removeHandlers []LeaveHandler
		removeMu       sync.Mutex
	}
)

func NewPool() *Pool {
	return &Pool{
		clientMapMu: sync.Mutex{},
		clientMap:   map[*Client]bool{},
	}
}

func (p *Pool) Add(c *Client) {
	p.clientMapMu.Lock()
	p.clientMap[c] = true
	p.clientMapMu.Unlock()
	for _, h := range p.joinHandlers {
		h(p, c)
	}
}

func (p *Pool) Remove(c *Client) {
	p.clientMapMu.Lock()
	delete(p.clientMap, c)
	p.clientMapMu.Unlock()
	for _, h := range p.removeHandlers {
		h(p, c)
	}
}

func (p *Pool) Broadcast(packet *Packet) {
	for c, _ := range p.clientMap {
		c.messageChan <- packet
	}
}

func (p *Pool) String() {

}

func (p *Pool) AfterAdd(handler JoinHandler) {
	p.joinMu.Lock()
	p.joinHandlers = append(p.joinHandlers, handler)
	p.joinMu.Unlock()
}

func (p *Pool) AfterRemove(handler LeaveHandler) {
	p.joinMu.Lock()
	p.removeHandlers = append(p.removeHandlers, handler)
	p.joinMu.Unlock()
}
