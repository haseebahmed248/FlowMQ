// Example producer client
package main

import (
	"log"
	"net"
)

func main() {
	data, err := net.Dial("tcp", "localhost:9876")
	if err != nil {
		log.Print(err)
		return
	}
	data.Write([]byte{0x01, 0x00, 0x00, 0x00, 0x00})
	buf := make([]byte, 1024)
	data.Read(buf)
	if err != nil {
		log.Print(err)
	}
	log.Print(buf)
	log.Print(string(buf))
}
