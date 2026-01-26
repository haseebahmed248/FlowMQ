// Topic management, message storage per topic
package topic

import (
	"errors"
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
	Name     string
	Messages []Message
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
		Name:     name,
		Messages: []Message{},
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
		Status:    "PENDING",
	})
	return id, nil
}
