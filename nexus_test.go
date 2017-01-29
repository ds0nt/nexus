package nexus

import (
	"log"
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

const testOrigin = "http://localhost"
const testListenAddr = ":54222"
const testDialAddr = "ws://localhost:54222/"

func TestMain(t *testing.M) {
	go startTestServer()

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
	http.ListenAndServe(testListenAddr, websocket.Handler(n.Serve))
}

func TestNexus(t *testing.T) {

	ws, err := websocket.Dial(testDialAddr, "", testOrigin)
	if err != nil {
		panic(err)
	}
	p := Packet{
		Type: "test1",
		Data: "rawr",
	}
	log.Printf("Sending %v", p)
	err = websocket.JSON.Send(ws, &p)
	if err != nil {
		panic(err)
	}
	err = websocket.JSON.Receive(ws, &p)
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
	err = websocket.JSON.Send(ws, &p)
	if err != nil {
		panic(err)
	}
	err = websocket.JSON.Receive(ws, &p)
	if err != nil {
		panic(err)
	}
	log.Printf("Receive %v", p)
	if p.Data != "rawr2" {
		t.Fail()
		return
	}

	ws.Close()
}

func TestNexusTwice(t *testing.T) {

	ws, err := websocket.Dial(testDialAddr, "", testOrigin)
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
	err = websocket.JSON.Send(ws, &p)
	if err != nil {
		panic(err)
	}
	err = websocket.JSON.Receive(ws, &p)
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
	err = websocket.JSON.Send(ws, &p)
	if err != nil {
		panic(err)
	}
	err = websocket.JSON.Receive(ws, &p)
	if err != nil {
		panic(err)
	}
	log.Printf("Receive %v", p)
	if p.Data != "rawr2" {
		t.Fail()
		return
	}

	ws.Close()
}

func TestNexusBadMessage(t *testing.T) {

	ws, err := websocket.Dial(testDialAddr, "", testOrigin)
	if err != nil {
		panic(err)
	}
	err = ws.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
	if err != nil {
		panic(err)
	}
	p := "this-is-not-a-packet-at-all"
	log.Printf("Sending %v", p)
	err = websocket.JSON.Send(ws, p)
	if err != nil {
		panic(err)
	}
	err = websocket.JSON.Receive(ws, &p)
	if err == nil {
		panic("Expected error")
	}
	log.Printf("Expects error: got %v", err)

	ws.Close()
}

func TestNexusBadHandler(t *testing.T) {

	ws, err := websocket.Dial(testDialAddr, "", testOrigin)
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
	err = websocket.JSON.Send(ws, &p)
	if err != nil {
		panic(err)
	}

	// connection should stay alive
	err = websocket.JSON.Receive(ws, &p)
	if err == nil {
		panic("Expected read timeout error")
	}
	log.Printf("Receive %v", p)
	log.Printf("Expects read error: got %v", err)
	ws.Close()
}
