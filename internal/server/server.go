// TCP listener, connection handling
package server

import (
	"flowmq/internal/protocol"
	"log"
	"net"
	"sync"
)

var clients = make(map[string]net.Conn)
var mu sync.Mutex

func handleConnection(conn net.Conn) {
	clientAddr := conn.RemoteAddr().String()
	mu.Lock()
	clients[clientAddr] = conn
	mu.Unlock()

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
		if command[0] == 0x01 {
			protocol.WriteFrame(conn, 0x02, payload)
		}
	}
}

func StartServer() {
	server, err := net.Listen("tcp", "localhost:9876")
	if err != nil {
		log.Print(err)
		return
	}
	log.Print("Server listening to port 9876")
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
