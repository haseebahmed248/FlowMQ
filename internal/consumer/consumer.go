// Consumer/group tracking, delivery assignments
package consumer

import (
	"errors"
	"flowmq/internal/topic"
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
	for _, v := range topic.Topics {
		for k, v1 := range v.Messages {
			if v1.ID == messageID {
				v.Messages[k].ID = "ACKNOWLEDGED"
			}
		}
	}
	return nil
}

func NACK(connAddr net.Conn, messageID string) error {
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
	for _, v := range topic.Topics {
		for _, v1 := range v.Messages {
			if v1.ID == messageID {
				for k1 := range v.Subscribers {
					if v.Subscribers[k1] == connAddr {
						v.Subscribers[k1].Write(v1.Payload)
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
		messageID := ""
		var timeLeft time.Time
		var client net.Conn
		for _, v := range Consumers {
			for k, v1 := range v.PendingMessages {
				client = v.Conn
				messageID = k
				timeLeft = v1
			}
		}
		if messageID != "" {
			elapsedSeconds := int(time.Since(timeLeft).Seconds())
			if elapsedSeconds > 30 {
				for _, v := range topic.Topics {
					for _, v1 := range v.Messages {
						if v1.ID == messageID {
							client.Write(v1.Payload)
						}
					}
				}
			}
		}
	}
}
