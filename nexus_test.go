package nexus

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gorilla/websocket"
)

const testOrigin = "http://127.0.0.1:55511"
const testListenAddr = ":55511"
const testDialAddr = "ws://127.0.0.1:55511/"

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
	// websocket.Dial()
	http.ListenAndServe(testListenAddr, http.HandlerFunc(n.Handler))
}

func TestNexus(t *testing.T) {

	ws, _, err := websocket.DefaultDialer.Dial(testDialAddr, nil)
	defer ws.Close()
	if err != nil {
		panic(err)
	}
	p := Packet{
		Type: "test1",
		Data: "rawr",
	}
	log.Printf("Sending %v", p)
	err = ws.WriteJSON(&p)
	if err != nil {
		panic(err)
	}
	err = ws.ReadJSON(&p)
	if err != nil {
		panic(err)
	}
	log.Printf("Receive %v", p)
	if p.Data != "rawr1" {
		t.Fail()
		return
	}

	p = Packet{
		Type: "test2",
		Data: "rawr",
	}
	log.Printf("Sending %v", p)
	err = ws.WriteJSON(&p)
	if err != nil {
		panic(err)
	}
	err = ws.ReadJSON(&p)
	if err != nil {
		panic(err)
	}
	log.Printf("Receive %v", p)
	if p.Data != "rawr2" {
		t.Fail()
		return
	}

}

func TestNexusTwice(t *testing.T) {

	ws, _, err := websocket.DefaultDialer.Dial(testDialAddr, nil)
	defer ws.Close()
	if err != nil {
		panic(err)
	}

	err = ws.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
	if err != nil {
		panic(err)
	}
	p := Packet{
		Type: "test1",
		Data: "rawr",
	}
	log.Printf("Sending %v", p)
	err = ws.WriteJSON(&p)
	if err != nil {
		panic(err)
	}
	err = ws.ReadJSON(&p)
	if err != nil {
		panic(err)
	}
	log.Printf("Receive %v", p)
	if p.Data != "rawr1" {
		t.Fail()
		return
	}

	p = Packet{
		Type: "test2",
		Data: "rawr",
	}
	log.Printf("Sending %v", p)
	err = ws.WriteJSON(&p)
	if err != nil {
		panic(err)
	}
	err = ws.ReadJSON(&p)
	if err != nil {
		panic(err)
	}
	log.Printf("Receive %v", p)
	if p.Data != "rawr2" {
		t.Fail()
		return
	}
}

func TestNexusBadMessage(t *testing.T) {

	ws, _, err := websocket.DefaultDialer.Dial(testDialAddr, nil)
	if err != nil {
		panic(err)
	}
	err = ws.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
	if err != nil {
		panic(err)
	}
	p := "this-is-not-a-packet-at-all"
	log.Printf("Sending %v", p)
	err = ws.WriteJSON(p)
	if err != nil {
		panic(err)
	}
	err = ws.ReadJSON(&p)
	if err == nil {
		panic("Expected error")
	}
	log.Printf("Expects error: got %v", err)

	ws.Close()
}

func TestNexusBadHandler(t *testing.T) {

	ws, _, err := websocket.DefaultDialer.Dial(testDialAddr, nil)
	if err != nil {
		panic(err)
	}
	err = ws.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
	if err != nil {
		panic(err)
	}

	p := Packet{
		Type: "got handler???",
		Data: "nope.",
	}
	log.Printf("Sending %v", p)
	err = ws.WriteJSON(&p)
	if err != nil {
		panic(err)
	}

	// connection should stay alive
	err = ws.ReadJSON(&p)
	if err == nil {
		panic("Expected read timeout error")
	}
	log.Printf("Receive %v", p)
	log.Printf("Expects read error: got %v", err)
	ws.Close()
}

func TestNexusConnectionTime(t *testing.T) {
	fmt.Println("Test Nexus Connection Time")
	ws, _, err := websocket.DefaultDialer.Dial(testDialAddr, nil)
	if err != nil {
		panic(err)
	}

	p := Packet{
		Type: "test1",
		Data: "test",
	}

	for i := 0; i < 15; i++ {
		log.Printf("Sending %v", p)
		err = ws.WriteJSON(&p)
		assert.Nil(t, err)
		time.Sleep(time.Second / 10)

		// connection should stay alive
		err = ws.ReadJSON(&p)
		assert.Nil(t, err)
		log.Printf("Receive %v", p)

		time.Sleep(time.Second / 1000)
	}

	ws.Close()
}
