package nexus

import "fmt"

type Packet struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func (p *Packet) String() string {
	return fmt.Sprintf("packet: type=%s data=%s", p.Type, p.Data)
}
