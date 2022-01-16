package nexus

import (
	"log"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

const testOrigin = "http://127.0.0.1:55511"
const testListenAddr = ":55511"
const testDialAddr = "ws://127.0.0.1:55511/"

var streamDone = map[string]chan struct{}{}
var doneMu = sync.Mutex{}

func TestMain(t *testing.M) {
	go startTestServer()

	time.Sleep(1 * time.Second)
	t.Run()
}

func startTestServer() {
	n := NewNexus()
	n.Handle("test1", func(c *Client, p *Packet) {
		c.Send(&Packet{
			Type: "test1",
			Data: p.Data + "1",
		})
	})

	n.Handle("test2", func(c *Client, p *Packet) {
		c.Send(&Packet{
			Type: "test2",
			Data: p.Data + "2",
		})
	})

	n.StreamHandle("test3", func(c *Context, p *Packet) {
		doneMu.Lock()
		streamDone[p.StreamID] = make(chan struct{})
		defer close(streamDone[p.StreamID])
		doneMu.Unlock()

		for {
			select {
			case <-c.StreamContext.Done():
				log.Println("stream", p.StreamID, "closed")
				return
			default:
				c.Client.Send(&Packet{
					Type: "test3",
					Data: p.Data + "3",
				})
				time.Sleep(time.Millisecond)
			}
		}
	})

	// websocket.Dial()
	http.ListenAndServe(testListenAddr, http.HandlerFunc(n.Handler))
}

func dial(t *testing.T) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(testDialAddr, nil)
	assert.Nil(t, err)
	assert.Nil(t, ws.SetReadDeadline(time.Now().Add(time.Millisecond*10)))
	return ws
}

func TestNexus(t *testing.T) {
	ws := dial(t)
	defer ws.Close()

	p := Packet{
		Type: "test1",
		Data: "rawr",
	}
	log.Printf("Sending %v", p)
	assert.Nil(t, ws.WriteJSON(&p))
	assert.Nil(t, ws.ReadJSON(&p))
	log.Printf("Receive %v", p)
	assert.Equal(t, p.Data, "rawr1")

	p = Packet{
		Type: "test2",
		Data: "rawr",
	}
	log.Printf("Sending %v", p)
	assert.Nil(t, ws.WriteJSON(&p))
	assert.Nil(t, ws.ReadJSON(&p))
	log.Printf("Receive %v", p)
	assert.Equal(t, p.Data, "rawr2")

}
func TestNexusStreamer(t *testing.T) {
	ws := dial(t)
	defer ws.Close()

	p := Packet{
		Type:     "test3",
		Data:     "rawr",
		StreamID: "001",
	}
	log.Printf("Sending %v", p)
	assert.Nil(t, ws.WriteJSON(&p))

	p2 := Packet{
		Type:     "test3",
		Data:     "wewt",
		StreamID: "002",
	}
	log.Printf("Sending %v", p2)
	assert.Nil(t, ws.WriteJSON(&p2))

	go func() {
		time.Sleep(time.Millisecond * 5)
		killp := Packet{
			Type:     "-test3",
			StreamID: "001",
		}
		log.Printf("Sending %v", killp)
		assert.Nil(t, ws.WriteJSON(&killp))

		time.Sleep(time.Millisecond)
		killp = Packet{
			Type:     "-test3",
			StreamID: "002",
		}
		log.Printf("Sending %v", killp)
		assert.Nil(t, ws.WriteJSON(&killp))
	}()

	x := 0
	for {
		select {
		case <-streamDone["001"]:
			select {
			case <-streamDone["002"]:
				return
			default:
			}
		default:
			x++
		}
		assert.True(t, x < 50, "streamer 001 did not close before 50 messages")
		err := ws.ReadJSON(&p)
		if err != nil {
			assert.True(t, x >= 6, "stream 002 killed before receiving 4 messages with err %s", err.Error())
		}
		log.Printf("Receive %v", p)
		assert.True(t, p.Data == "rawr3" || p.Data == "wewt3")

	}
}

func TestNexusTwice(t *testing.T) {
	ws := dial(t)
	defer ws.Close()

	p := Packet{
		Type: "test1",
		Data: "rawr",
	}
	log.Printf("Sending %v", p)
	assert.Nil(t, ws.WriteJSON(&p))
	assert.Nil(t, ws.ReadJSON(&p))
	log.Printf("Receive %v", p)
	assert.Equal(t, p.Data, "rawr1")

	p = Packet{
		Type: "test2",
		Data: "rawr",
	}
	log.Printf("Sending %v", p)
	assert.Nil(t, ws.WriteJSON(&p))
	assert.Nil(t, ws.ReadJSON(&p))
	log.Printf("Receive %v", p)
	assert.Equal(t, p.Data, "rawr2")
}

func TestNexusBadMessage(t *testing.T) {
	ws := dial(t)
	defer ws.Close()

	p := "this-is-not-a-packet-at-all"
	log.Printf("Sending %v", p)
	assert.Nil(t, ws.WriteJSON(p))
	assert.NotNil(t, ws.ReadJSON(&p))

}

func TestNexusBadHandler(t *testing.T) {
	ws := dial(t)
	defer ws.Close()

	p := Packet{
		Type: "got handler???",
		Data: "nope.",
	}
	log.Printf("Sending %v", p)
	assert.Nil(t, ws.WriteJSON(&p))

	// connection should stay alive
	assert.NotNil(t, ws.ReadJSON(&p))
}

func TestNexusConnectionTime(t *testing.T) {
	ws := dial(t)
	assert.Nil(t, ws.SetReadDeadline(time.Now().Add(time.Second*10)))
	defer ws.Close()

	p := Packet{
		Type: "test1",
		Data: "test",
	}

	for i := 0; i < 15; i++ {
		log.Printf("Sending %v", p)
		assert.Nil(t, ws.WriteJSON(&p))
		time.Sleep(time.Second / 10)

		// connection should stay alive
		assert.Nil(t, ws.ReadJSON(&p))
		log.Printf("Receive %v", p)

		time.Sleep(time.Second / 1000)
	}

}
