package models

import (
	"net"
	"time"
)

type Message struct {
	ID        string
	Payload   []byte
	Timestamp time.Time
	Status    string // "PENDING", "DELIVERED", "ACKNOWLEDGED"
}
type Topic struct {
	Name        string
	Messages    []Message
	Subscribers []net.Conn
	Group       map[string][]net.Conn
	GroupIndex  map[string]int
}

var Topics = make(map[string]*Topic)
