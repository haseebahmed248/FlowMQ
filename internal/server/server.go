// TCP listener, connection handling
package server

import (
	"flowmq/internal/consumer"
	"flowmq/internal/protocol"
	"flowmq/internal/storage"
	"flowmq/internal/topic"
	"log"
	"net"
	"strings"
	"sync"
)

var clients = make(map[string]net.Conn)
var mu sync.Mutex

func handleConnection(conn net.Conn) {
	clientAddr := conn.RemoteAddr().String()
	mu.Lock()
	clients[clientAddr] = conn
	mu.Unlock()

	defer func() {
		topic.Cleanup(conn)
	}()
	defer conn.Close()

	defer func() {
		mu.Lock()
		log.Print("Client disconnected with address: ", clientAddr)
		delete(clients, clientAddr)
		mu.Unlock()
	}()

	for {
		command, _, payload, err := protocol.ReadFrame(conn)
		if err != nil {
			return
		}
		switch command[0] {
		case 0x01:
			{
				protocol.WriteFrame(conn, 0x02, payload)
				break
			}
		case 0x03: //CREATE_TOPIC
			{
				if len(payload) > 0 {
					err := topic.CreateTopic(string(payload))
					if err != nil {
						protocol.WriteFrame(conn, 0xFF, payload)
						break
					}
					protocol.WriteFrame(conn, 0x10, payload)
					break
				}
				protocol.WriteFrame(conn, 0xFF, payload)
			}
		case 0x08:
			{
				topics := topic.ListTopics()
				result := strings.Join(topics, ",")
				protocol.WriteFrame(conn, 0x10, []byte(result))
			}
		case 0x09:
			{
				if len(payload) > 0 {
					err := topic.DeleteTopic(string(payload))
					if err != nil {
						protocol.WriteFrame(conn, 0xFF, payload)
						break
					}
					protocol.WriteFrame(conn, 0x10, payload)
					break
				}
				protocol.WriteFrame(conn, 0xFF, payload)
			}
		case 0x04:
			{
				data := strings.Split(string(payload), "\x00")
				if len(data) >= 2 && data[1] != "" && data[0] != "" {
					response, err := topic.Publish(data[0], []byte(data[1]))
					if err != nil {
						protocol.WriteFrame(conn, 0xFF, payload)
						break
					}
					protocol.WriteFrame(conn, 0x10, []byte(response))
					break
				}
			}
		case 0x05:
			{
				topicName := strings.Split(string(payload), "\x00")
				if len(topicName) >= 2 {
					err := topic.Subscribe(topicName[0], conn, topicName[1])
					if err != nil {
						protocol.WriteFrame(conn, 0xFF, []byte(err.Error()))
						break
					}
				} else {
					err := topic.Subscribe(topicName[0], conn, "")
					if err != nil {
						protocol.WriteFrame(conn, 0xFF, []byte(err.Error()))
						break
					}
				}

				protocol.WriteFrame(conn, 0x10, nil)
			}
		case 0x06:
			{
				messageID := string(payload)
				consumer.Acknowledged(conn, messageID)
				protocol.WriteFrame(conn, 0x10, nil)
			}
		case 0x0A:
			{
				messageID := string(payload)
				consumer.NACK(conn, messageID)
				protocol.WriteFrame(conn, 0x10, nil)
			}
		}
	}
}

func StartServer() {
	server, err := net.Listen("tcp", "0.0.0.0:9876")
	if err != nil {
		log.Print(err)
		return
	}
	log.Print("Server listening to port 9876")
	storage.ReplayWAL()
	go consumer.StartRedeliveryChecker()
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Print(err)
			return
		}
		log.Print("Client connected!")

		go handleConnection(conn)

	}
}
