// Topic management, message storage per topic
package topic

import (
	"errors"
	"flowmq/internal/models"
	"flowmq/internal/protocol"
	"flowmq/internal/storage"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

var mu sync.RWMutex

// Functions
func CreateTopic(name string) error {
	mu.Lock()
	defer mu.Unlock()

	if models.Topics[name] != nil {
		return errors.New("Topic already exist")
	}

	models.Topics[name] = &models.Topic{
		Name:        name,
		Messages:    []models.Message{},
		Subscribers: []net.Conn{},
		Group:       make(map[string][]net.Conn),
		GroupIndex:  map[string]int{},
	}
	return nil
}

func ListTopics() []string {
	mu.RLock()
	defer mu.RUnlock()
	var response []string
	for k, _ := range models.Topics {
		response = append(response, k)
	}
	return response
}

func DeleteTopic(name string) error {
	mu.Lock()
	defer mu.Unlock()
	if models.Topics[name] == nil {
		return errors.New("No topic exsist on this name")
	}
	delete(models.Topics, name)
	return nil
}

func Publish(name string, payload []byte) (string, error) {
	mu.Lock()
	defer mu.Unlock()
	id := uuid.NewString()
	if models.Topics[name] == nil {
		return "", errors.New("Topic doesn't exists")
	}

	models.Topics[name].Messages = append(models.Topics[name].Messages, models.Message{
		ID:        id,
		Payload:   payload,
		Timestamp: time.Now(),
		Status:    "PENDING",
	})
	storage.AppendToWAL(name, models.Message{
		ID:        id,
		Payload:   payload,
		Timestamp: time.Now(),
		Status:    "PENDING",
	})
	fullPayload := id + "\x00" + string(payload)

	for _, conn := range models.Topics[name].Subscribers {
		protocol.WriteFrame(conn, 0x07, []byte(fullPayload))
	}

	// Convert map to slice for round-robin indexing

	for k, v := range models.Topics[name].Group {
		groupIndex := models.Topics[name].GroupIndex[k]
		connection := v[groupIndex%len(v)]
		protocol.WriteFrame(connection, 0x07, []byte(fullPayload))
		models.Topics[name].GroupIndex[k]++
	}

	return id, nil
}

func Subscribe(name string, conn net.Conn, groupName string) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := models.Topics[name]; !ok {
		return errors.New("No topic exists of specified name")
	}
	if groupName != "" {
		models.Topics[name].Group[groupName] = append(models.Topics[name].Group[groupName], conn)
	} else {
		models.Topics[name].Subscribers = append(models.Topics[name].Subscribers, conn)
	}

	return nil
}

func Cleanup(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	for _, topic := range models.Topics {
		// Clean up from Groups
		for groupName, conns := range topic.Group {
			filtered := make([]net.Conn, 0, len(conns))
			for _, c := range conns {
				if c != conn {
					filtered = append(filtered, c)
				}
			}
			topic.Group[groupName] = filtered
		}

		// Clean up from Subscribers
		filtered := make([]net.Conn, 0, len(topic.Subscribers))
		for _, c := range topic.Subscribers {
			if c != conn {
				filtered = append(filtered, c)
			}
		}
		topic.Subscribers = filtered
	}
}
