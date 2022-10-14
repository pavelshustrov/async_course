package consumer

import (
	"fmt"
	"os"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

func MustNewConsumer(groupID string) *kafka.Consumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("bootstrap.servers"),
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(fmt.Errorf("kafka producer creation err: %w", err))
	}

	return c
}
