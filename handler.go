package nexus

type Handler func(*Client, *Packet)
type Streamer func(*Context)
