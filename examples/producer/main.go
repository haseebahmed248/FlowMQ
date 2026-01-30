// Example producer client
package main

import (
	"encoding/binary"
	"log"
	"net"
)

func main() {
	// broker is the service name
	data, err := net.Dial("tcp", "broker:9876")
	if err != nil {
		log.Print(err)
		return
	}
	defer data.Close()

	topicName := []byte("world")

	// Create Topic (0x03)
	log.Println("Creating topic 'world'...")
	data.Write([]byte{0x03})

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(topicName)))
	data.Write(lenBuf)
	data.Write(topicName)

	// Read create topic response
	buf := make([]byte, 5)
	_, err = data.Read(buf)
	if err != nil {
		log.Print("Error reading create response:", err)
		return
	}
	log.Printf("Topic created. Response: %v", buf)

	// Publish message (0x04)
	message := []byte("Hello from producer!")
	log.Println("Publishing message to topic 'world'...")

	data.Write([]byte{0x04})

	// Payload format for PUBLISH: topicName\x00message
	payload := append(topicName, '\x00')
	payload = append(payload, message...)

	lenBuf2 := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf2, uint32(len(payload)))
	data.Write(lenBuf2)
	data.Write(payload)

	// Read publish response
	buf2 := make([]byte, 50)
	n, err := data.Read(buf2)
	if err != nil {
		log.Print("Error reading publish response:", err)
		return
	}
	log.Printf("Message published. Response: %s", string(buf2[:n]))
}
