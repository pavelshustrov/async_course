package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type Validator interface {
	Validate(bytes []byte, domain, eventName string, version string) error
}

type consumer struct {
	kafkaConsumer *kafka.Consumer
	db            *pgx.Conn
	eventHandlers map[string]func(string) error
}

func New(kafkaConsumer *kafka.Consumer, db *pgx.Conn, eventHandlers map[string]func(string) error) *consumer {
	return &consumer{
		kafkaConsumer: kafkaConsumer,
		db:            db,
		eventHandlers: eventHandlers,
	}
}

func (c *consumer) Consume(ctx context.Context, topic string) error {
	err := c.kafkaConsumer.Subscribe(topic, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			message, err := c.kafkaConsumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				switch e := err.(type) {
				case kafka.Error:
					if e.Code() == kafka.ErrTimedOut {
						continue
					}
					return err
				default:
					return err
				}
			}
			headers := make(map[string]string)
			for _, v := range message.Headers {
				headers[v.Key] = string(v.Value)
			}

			event := headers["event_name"]
			payload := string(message.Value)

			f, ok := c.eventHandlers[event]
			if !ok {
				fmt.Println("failed", event, payload)
				continue
			}
			if err := f(payload); err != nil {
				return err
			}

		}
	}
}
