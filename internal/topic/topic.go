// Topic management, message storage per topic
package topic

import (
	"errors"
	"flowmq/internal/protocol"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
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
}

var topics = make(map[string]*Topic)
var mu sync.RWMutex

// Functions
func CreateTopic(name string) error {
	mu.Lock()
	defer mu.Unlock()

	if topics[name] != nil {
		return errors.New("Topic already exist")
	}

	topics[name] = &Topic{
		Name:        name,
		Messages:    []Message{},
		Subscribers: []net.Conn{},
	}
	return nil
}

func ListTopics() []string {
	mu.RLock()
	defer mu.RUnlock()
	var response []string
	for k, _ := range topics {
		response = append(response, k)
	}
	return response
}

func DeleteTopic(name string) error {
	mu.Lock()
	defer mu.Unlock()
	if topics[name] == nil {
		return errors.New("No topic exsist on this name")
	}
	delete(topics, name)
	return nil
}

func Publish(name string, payload []byte) (string, error) {
	mu.Lock()
	defer mu.Unlock()
	id := uuid.NewString()
	if topics[name] == nil {
		return "", errors.New("Topic doesn't exists")
	}

	topics[name].Messages = append(topics[name].Messages, Message{
		ID:        id,
		Payload:   payload,
		Timestamp: time.Now(),
		Status:    "DELIVERED",
	})
	fullPayload := id + "\x00" + string(payload)
	for _, conn := range topics[name].Subscribers {
		protocol.WriteFrame(conn, 0x07, []byte(fullPayload))
	}

	return id, nil
}

func Subscribe(name string, conn net.Conn) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := topics[name]; !ok {
		return errors.New("No topic exists of specified name")
	}
	topics[name].Subscribers = append(topics[name].Subscribers, conn)
	return nil
}
