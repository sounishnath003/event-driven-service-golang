package kafkaa

import (
	"log"

	"github.com/segmentio/kafka-go"
)

func NewKafkaReaderClient(kconf kafka.ReaderConfig) *kafka.Reader {
	kafkaWC := kafka.NewReader(kconf)
	log.Println("kafka writer posts consumers has been connected")
	return kafkaWC
}
