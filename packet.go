package nexus

import "fmt"

type Packet struct {
	// Thread is a unique reference to a discussion.
	// Thread allows for back and forth communication between client and server
	// in a limited/temporary context
	// i.e. responding with an error would use the same thread id as the request
	// and responding with data would carry the same thread id as the request
	Thread int    `json:"thread"`
	Type   string `json:"type"`
	Data   string `json:"data"`
}

func (p *Packet) String() string {
	return fmt.Sprintf("packet: thread=%d type=%s data=%s", p.Thread, p.Type, p.Data)
}
