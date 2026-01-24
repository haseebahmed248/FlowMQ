# FlowMQ

Distributed message queue built from scratch in Go — custom binary protocol, topic-based pub/sub, consumer groups, persistence (WAL), Docker-ready. No external libraries.

## Architecture

```
Producer(s)                        Consumer(s)
    |                                  ^
    | PUBLISH                          | DELIVER
    v                                  |
+------------------------------------------+
|              FlowMQ Broker               |
|                                          |
|  +----------+  +----------+  +--------+ |
|  | Protocol |->|  Router  |->| Topics | |
|  |  Parser  |  |          |  |        | |
|  +----------+  +----------+  +--------+ |
|                                  |       |
|                          +-------+-----+ |
|                          | Storage/WAL | |
|                          +-------------+ |
+------------------------------------------+
```

## Protocol

Binary frame format:
```
[Command Type: 1 byte][Payload Length: 4 bytes (big-endian)][Payload: N bytes]
```

## Status

Work in Progress

## Run

```bash
go run cmd/flowmq/main.go
```

## Author

Haseeb Ahmed - [GitHub](https://github.com/haseebahmed248)
