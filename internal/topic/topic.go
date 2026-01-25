// Topic management, message storage per topic
package topic

import (
	"errors"
	"sync"
)

type Message struct{}

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
