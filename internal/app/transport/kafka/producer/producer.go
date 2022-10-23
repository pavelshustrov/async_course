package producer

import (
	"fmt"
	"os"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

func MustNewProducer() *kafka.Producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("bootstrap.servers"),
		"go.batch.producer": true,
	})
	if err != nil {
		panic(fmt.Errorf("kafka producer creation err: %w", err))
	}
	return p
}
