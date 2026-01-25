// Binary protocol parser/serializer
package protocol

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

func ReadFrame(conn net.Conn) ([]byte, uint32, []byte, error) {
	// command (1st byte)
	command := make([]byte, 1)
	_, err := io.ReadFull(conn, command)
	if err != nil {
		return nil, 0, nil, errors.New("user disconnected")
	}

	// Length (4 bytes)
	lenBuf := make([]byte, 4)
	_, err = io.ReadFull(conn, lenBuf)
	if err != nil {
		return nil, 0, nil, errors.New("user disconnected")
	}
	length := binary.BigEndian.Uint32(lenBuf)

	// Payload (Rest of length bytes)
	payload := make([]byte, length)

	_, err = io.ReadFull(conn, payload)
	if err != nil {
		return nil, 0, nil, errors.New("user disconnected")
	}

	return command, length, payload, nil
}

func WriteFrame(conn net.Conn, command byte, payload []byte) {
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(payload)))
	conn.Write([]byte{command})
	conn.Write(lenBuf)
	conn.Write(payload)
}
