package nexus

import (
	"encoding/json"
	"fmt"
)

type Packet struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func (p *Packet) String() string {
	return fmt.Sprintf("packet: type=%s data=%s", p.Type, p.Data)
}

func (p *Packet) Bind(i interface{}) error {
	return json.Unmarshal([]byte(p.Data), &i)
}
