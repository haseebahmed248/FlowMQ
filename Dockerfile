FROM golang:1.25.5-alpine

WORKDIR /app

COPY . .

RUN go build -o flowmq ./cmd/flowmq


EXPOSE 9876


CMD [ "./flowmq" ]
