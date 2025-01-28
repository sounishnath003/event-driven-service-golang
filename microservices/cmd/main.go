package main

import (
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sounishnath003/event-driven-service/microservices/cmd/kafkaa"
	"github.com/sounishnath003/event-driven-service/microservices/cmd/mongodatabase"
)

func main() {
	// Database client
	mongoClient, err := mongodatabase.NewMongoDatabaseClient()
	if err != nil {
		panic(err)
	}

	// Server Conf
	serverConf := ServerConf{
		Port: 3000,
		KafkaWriter: kafkaa.NewKafkaWriterClient(kafka.WriterConfig{
			Topic:    "create-posts",
			Brokers:  []string{"localhost:9092"},
			Balancer: &kafka.Hash{},
			Dialer: &kafka.Dialer{
				Timeout:   10 * time.Second,
				DualStack: true,
			},
		}),
		MongoClient: mongoClient,
	}

	// Create the server
	server := NewServer(serverConf)

	// Start the server
	server.Start()
}
