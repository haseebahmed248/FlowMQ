// Example consumer client
package main

import (
	"encoding/binary"
	"log"
	"net"
	"strings"
)

func main() {
	data, err := net.Dial("tcp", "localhost:9876")
	if err != nil {
		log.Print(err)
		return
	}
	payload := []byte("world")

	// Write command (1 byte)
	// data.Write([]byte{0x03})

	// lenBuf := make([]byte, 4)
	// binary.BigEndian.PutUint32(lenBuf, uint32(len(payload)))
	// data.Write(lenBuf)

	data.Write([]byte{0x05})

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(payload)))
	data.Write(lenBuf)

	data.Write(payload)

	// Read the subscription response
	buf := make([]byte, 8)
	n, err := data.Read(buf)
	if err != nil {
		log.Print("Error reading response:", err)
		return
	}
	log.Printf("Subscribed successfully. Response: %v", buf[:n])

	// Keep the connection alive and listen for messages
	log.Println("Waiting for messages on topic 'world'...")
	for {
		// Read command byte
		cmdBuf := make([]byte, 1)
		_, err := data.Read(cmdBuf)
		if err != nil {
			log.Print("Connection closed:", err)
			return
		}

		// Read length
		lenBuf := make([]byte, 4)
		_, err = data.Read(lenBuf)
		if err != nil {
			log.Print("Error reading length:", err)
			return
		}
		length := binary.BigEndian.Uint32(lenBuf)

		// Read payload
		msgBuf := make([]byte, length)
		_, err = data.Read(msgBuf)
		if err != nil {
			log.Print("Error reading message:", err)
			return
		}
		data := strings.Split(string(msgBuf), "\x00")
		log.Printf("Received message: %s", data[1])
	}
}
