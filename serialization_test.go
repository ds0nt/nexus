package nexus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelimitedBlank(t *testing.T) {
	p := &Packet{
		Type:     "",
		StreamID: "",
		Data:     "",
	}
	assert.Equal(t, DefaultDelimited.Marshal(p), []byte("0:0:"))

	_p, err := DefaultDelimited.Unmarshal(DefaultDelimited.Marshal(p))
	assert.Nil(t, err)
	assert.Equal(t, p, _p)
}

func TestDelimitedNoStream(t *testing.T) {
	p := &Packet{
		Type: "test",
		Data: "hooray",
	}
	assert.Equal(t, DefaultDelimited.Marshal(p), []byte("4:0:testhooray"))

	_p, err := DefaultDelimited.Unmarshal(DefaultDelimited.Marshal(p))
	assert.Nil(t, err)
	assert.Equal(t, p, _p)
}

func TestDelimitedWithDelimiter(t *testing.T) {
	p := &Packet{
		Type:     "BB::B",
		StreamID: "AA:A",
		Data:     "CCC",
	}
	assert.Equal(t, DefaultDelimited.Marshal(p), []byte("5:4:BB::BAA:ACCC"))
	_p, err := DefaultDelimited.Unmarshal(DefaultDelimited.Marshal(p))
	assert.Nil(t, err)
	assert.Equal(t, p, _p)

}

func TestJSONBlank(t *testing.T) {
	p := &Packet{
		Type:     "",
		StreamID: "",
		Data:     "",
	}

	_p, err := DefaultJSON.Unmarshal(DefaultJSON.Marshal(p))
	assert.Nil(t, err)
	assert.Equal(t, p, _p)
}

func TestJSONNoStream(t *testing.T) {
	p := &Packet{
		Type: "test",
		Data: "hooray",
	}

	_p, err := DefaultJSON.Unmarshal(DefaultJSON.Marshal(p))
	assert.Nil(t, err)
	assert.Equal(t, p, _p)
}

func TestJSONWithDelimiter(t *testing.T) {
	p := &Packet{
		Type:     "BB::B",
		StreamID: "AA:A",
		Data:     "CCC",
	}
	_p, err := DefaultJSON.Unmarshal(DefaultJSON.Marshal(p))
	assert.Nil(t, err)
	assert.Equal(t, p, _p)

}
