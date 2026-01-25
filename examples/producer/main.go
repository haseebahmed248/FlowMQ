// Example producer client
package main

import (
	"encoding/binary"
	"log"
	"net"
)

func main() {
	data, err := net.Dial("tcp", "localhost:9876")
	if err != nil {
		log.Print(err)
		return
	}
	payload := []byte("hello")

	// Write command (1 byte)
	data.Write([]byte{0x03})

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(payload)))
	data.Write(lenBuf)

	data.Write(payload)

	buf := make([]byte, 8)
	data.Read(buf)
	log.Print(buf)
	log.Print(string(buf))
}
