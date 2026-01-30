// Consumer/group tracking, delivery assignments
package consumer

import (
	"errors"
	"flowmq/internal/models"
	"flowmq/internal/protocol"
	"net"
	"sync"
	"time"
)

type Consumer struct {
	Conn            net.Conn
	PendingMessages map[string]time.Time
}

var Consumers = make(map[string]*Consumer)
var mu sync.Mutex

func Acknowledged(connAddr net.Conn, messageID string) error {
	mu.Lock()
	defer mu.Unlock()
	deleted := false
	for _, v := range Consumers {
		if _, ok := v.PendingMessages[messageID]; ok {
			delete(v.PendingMessages, messageID)
			deleted = true
		}
	}
	if !deleted {
		return errors.New("Message doesn't exists")
	}
	for _, v := range models.Topics {
		for k, v1 := range v.Messages {
			if v1.ID == messageID {
				v.Messages[k].Status = "ACKNOWLEDGED"
			}
		}
	}
	return nil
}

func NACK(connAddr net.Conn, messageID string) error {
	mu.Lock()
	defer mu.Unlock()
	deleted := false
	for _, v := range Consumers {
		if _, ok := v.PendingMessages[messageID]; ok {
			delete(v.PendingMessages, messageID)
			deleted = true
		}
	}
	if !deleted {
		return errors.New("Message doesn't exists")
	}
	for _, v := range models.Topics {
		for _, v1 := range v.Messages {
			if v1.ID == messageID {
				for k1 := range v.Subscribers {
					if v.Subscribers[k1] == connAddr {
						protocol.WriteFrame(connAddr, 0x07, v1.Payload)
					}
				}
			}
		}
	}
	return nil
}

func StartRedeliveryChecker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var client net.Conn
		for _, v := range Consumers {
			for k, v1 := range v.PendingMessages {
				client = v.Conn
				if k != "" {
					elapsedSeconds := int(time.Since(v1).Seconds())
					if elapsedSeconds > 30 {
						for k1, v := range models.Topics {
							for i, v1 := range v.Messages {
								if v1.ID == k {
									if v1.Retry >= 3 {
										data := models.Topics[k1].Messages[i]
										models.Topics[k1].Messages = append(models.Topics[k1].Messages[:i], models.Topics[k1].Messages[i+1:]...)
										if models.Topics["___dlq__"+k1] == nil {
											models.Topics["___dlq__"+k1] = &models.Topic{
												Name:        "___dlq__" + k1,
												Messages:    []models.Message{},
												Subscribers: []net.Conn{},
												Group:       make(map[string][]net.Conn),
												GroupIndex:  map[string]int{},
											}
										}
										models.Topics["___dlq__"+k1].Messages = append(models.Topics["___dlq__"+k1].Messages, data)
										continue
									}
									v.Messages[i].Retry++
									protocol.WriteFrame(client, 0x07, v1.Payload)
								}
							}
						}
					}
				}
			}
		}

	}
}
