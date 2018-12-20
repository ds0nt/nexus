package nexus

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type PacketMarshaler interface {
	Marshal(*Packet) []byte
	Unmarshal([]byte) (*Packet, error)
}

type Delimited struct {
	Delimiter string
}

var DefaultDelimited = &Delimited{":"}

func (d *Delimited) Unmarshal(bytes []byte) (*Packet, error) {
	str := string(bytes)
	fmt.Println(str)
	peices := strings.SplitN(str, d.Delimiter, 3)
	if len(peices) != 3 {
		return nil, errors.New("delimiter not found in message bytes")
	}

	lenType, err := strconv.Atoi(peices[0])
	if err != nil {
		return nil, errors.Wrap(err, "could not parse Type length")
	}
	lenStreamID, err := strconv.Atoi(peices[1])
	if err != nil {
		return nil, errors.Wrap(err, "could not parse StreamID length")
	}

	if len(peices[2]) < lenType+lenStreamID {
		return nil, errors.New("delimited message data bytes too short")
	}

	p := Packet{
		Type:     peices[2][:lenType],
		StreamID: peices[2][lenType : lenType+lenStreamID],
		Data:     peices[2][lenType+lenStreamID:],
	}
	return &p, nil
}

func (d *Delimited) Marshal(p *Packet) []byte {
	return []byte(
		strconv.Itoa(len(p.Type)) +
			d.Delimiter +
			strconv.Itoa(len(p.StreamID)) +
			d.Delimiter +
			p.Type +
			p.StreamID +
			p.Data)
}

type JSON struct{}

var DefaultJSON = &JSON{}

func (d *JSON) Unmarshal(bytes []byte) (*Packet, error) {
	p := &Packet{}
	err := json.Unmarshal(bytes, p)
	return p, err
}

func (d *JSON) Marshal(p *Packet) []byte {
	bytes, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
	}
	return bytes
}
