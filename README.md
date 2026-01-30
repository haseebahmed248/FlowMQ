# FlowMQ

Distributed message queue built from scratch in Go — custom binary protocol, topic-based pub/sub, consumer groups, persistence (WAL), Docker-ready.

**No external dependencies** for core functionality. Just Go's standard library + UUID generation.

## Features

- **Topics** — Create, list, delete named topics
- **Pub/Sub** — Producers publish, consumers subscribe and receive messages
- **Consumer Groups** — Load-balanced message delivery (round-robin within groups)
- **Fan-out** — Regular subscribers receive all messages
- **ACK/NACK** — At-least-once delivery with acknowledgments
- **Redelivery** — Unacknowledged messages redelivered after 30s timeout
- **Dead Letter Queue** — Failed messages (3+ retries) moved to `__dlq__` topic
- **Retention** — Auto-cleanup of acknowledged messages older than 60s
- **Persistence** — Write-Ahead Log (WAL) survives broker restart
- **Docker** — Containerized broker with example producer/consumer

## Architecture

```
Producer(s)                        Consumer(s)
    |                                  ^
    | PUBLISH                          | DELIVER
    v                                  |
+------------------------------------------+
|              FlowMQ Broker               |
|                                          |
|  +----------+  +----------+  +--------+  |
|  | Protocol |->|  Server  |->| Topics |  |
|  |  Parser  |  |  Router  |  |        |  |
|  +----------+  +----------+  +--------+  |
|                                  |        |
|                          +-------+------+ |
|                          | Storage/WAL  | |
|                          +--------------+ |
+------------------------------------------+
```

## Protocol

Binary frame format:
```
[Command: 1 byte][Length: 4 bytes big-endian][Payload: N bytes]
```

| Command | Byte | Direction | Payload |
|---------|------|-----------|---------|
| PING | 0x01 | Client → Broker | (empty) |
| PONG | 0x02 | Broker → Client | (empty) |
| CREATE_TOPIC | 0x03 | Client → Broker | topic_name |
| PUBLISH | 0x04 | Client → Broker | topic_name\x00message |
| SUBSCRIBE | 0x05 | Client → Broker | topic_name[\x00group_name] |
| ACK | 0x06 | Client → Broker | message_id |
| MESSAGE | 0x07 | Broker → Client | message_id\x00payload |
| LIST_TOPICS | 0x08 | Client → Broker | (empty) |
| DELETE_TOPIC | 0x09 | Client → Broker | topic_name |
| NACK | 0x0A | Client → Broker | message_id |
| SUCCESS | 0x10 | Broker → Client | (varies) |
| ERROR | 0xFF | Broker → Client | error_message |

## Quick Start

### Run locally
```bash
go run cmd/flowmq/main.go
# Or with custom port
go run cmd/flowmq/main.go -port 9999
```

### Run with Docker
```bash
docker build -t flowmq:1.0 .
docker-compose up
```

### Example: Producer
```bash
go run examples/producer/main.go
```

### Example: Consumer
```bash
go run examples/consumer/main.go
```

## Project Structure

```
FlowMQ/
├── cmd/flowmq/main.go          # Entry point, CLI flags
├── internal/
│   ├── server/server.go        # TCP listener, command routing
│   ├── protocol/protocol.go    # Binary protocol parser/serializer
│   ├── topic/topic.go          # Topic management, pub/sub, retention
│   ├── consumer/consumer.go    # ACK/NACK, redelivery, DLQ
│   ├── models/message.go       # Data structures
│   └── storage/wal.go          # Write-ahead log persistence
├── examples/
│   ├── producer/main.go        # Example producer client
│   └── consumer/main.go        # Example consumer client
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Message Flow

1. **Producer** connects and sends `CREATE_TOPIC`
2. **Consumer** connects and sends `SUBSCRIBE`
3. **Producer** sends `PUBLISH` with message
4. **Broker** stores message, writes to WAL, delivers to:
   - All regular subscribers (fan-out)
   - One consumer per group (round-robin)
5. **Consumer** receives `MESSAGE`, processes, sends `ACK`
6. If no `ACK` within 30s, broker redelivers
7. After 3 failed deliveries, message moves to DLQ

## What I Learned

Building FlowMQ taught me:
- Binary protocol design (vs text-based like Redis RESP)
- Message delivery guarantees (at-least-once)
- Consumer groups and load balancing
- Write-ahead logging for durability
- Goroutines for concurrent connections and background tasks
- Docker multi-container networking

## Author

**Haseeb Ahmed** — [GitHub](https://github.com/haseebahmed248) · [LinkedIn](https://linkedin.com/in/haseebahmed248)

## License

MIT
