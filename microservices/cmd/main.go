package main

import (
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sounishnath003/event-driven-service/microservices/cmd/kafkaa"
)

func main() {
	// Declare kafka writer
	kafkaWriter := kafkaa.NewKafkaWriterClient(kafka.WriterConfig{
		Topic:    "create-posts",
		Brokers:  []string{"localhost:9092"},
		Balancer: &kafka.Hash{},
		Dialer: &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
		},
	})
	server := NewServer(3000, kafkaWriter)
	server.Start()
}
