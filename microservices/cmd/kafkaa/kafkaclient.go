package kafkaa

import (
	"log"

	"github.com/segmentio/kafka-go"
)

func NewKafkaWriterClient(kconf kafka.WriterConfig) *kafka.Writer {
	kafkaWC := kafka.NewWriter(kconf)
	log.Println("kafka writer posts producers has been connected")
	return kafkaWC
}
