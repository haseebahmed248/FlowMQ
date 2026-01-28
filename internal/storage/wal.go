// Write-ahead log for message persistence
package storage

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"flowmq/internal/models"
	"log"
	"os"
	"time"
)

type WALEntry struct {
	Timestamp time.Time
	TopicName string
	Message   models.Message
}

func AppendToWAL(topicName string, msg models.Message) error {
	file, err := os.OpenFile("wal.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.New("Error opening the file")
	}
	defer file.Close()
	entry := WALEntry{time.Now(), topicName, msg}
	data, err := json.Marshal(entry)
	if err != nil {
		return errors.New("Error parsing data")
	}

	// Write: [4 bytes length][data]
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))

	file.Write(lenBuf)
	file.Write(data)
	file.Write([]byte("\n"))

	log.Print("WROTE: ", lenBuf, " ========= ", data)

	return nil
}

func ReplayWAL() {
	file, err := os.Open("wal.log")
	if err != nil {
		log.Print("Error opening the file for replay")
		return
	}
	defer file.Close()

	for {
		// Read 4-byte length prefix
		lenBuf := make([]byte, 4)
		n, err := file.Read(lenBuf)
		if err != nil || n == 0 {
			break // End of file
		}

		length := binary.BigEndian.Uint32(lenBuf)

		// Read the JSON data
		data := make([]byte, length)
		_, err = file.Read(data)
		if err != nil {
			log.Print("Error reading entry:", err)
			break
		}

		// Skip the newline if it exists
		file.Read(make([]byte, 1))

		// Unmarshal the JSON
		var entry WALEntry
		err = json.Unmarshal(data, &entry)
		if err != nil {
			log.Print("Error unmarshaling entry:", err)
			continue
		}
		if entry.Message.Status != "DELIVERED" {
			models.Topics[entry.TopicName] = &models.Topic{
				Name: entry.TopicName,
			}
			models.Topics[entry.TopicName].Messages = append(models.Topics[entry.TopicName].Messages, models.Message{
				ID:        entry.Message.ID,
				Payload:   entry.Message.Payload,
				Timestamp: entry.Message.Timestamp,
				Status:    entry.Message.Status,
			})
		}

		log.Printf("Restored: Topic=%s, MessageID=%s, Status=%s",
			entry.TopicName, entry.Message.ID, entry.Message.Status)
	}
}
