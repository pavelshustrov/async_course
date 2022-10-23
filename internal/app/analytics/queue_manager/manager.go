package queue_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"education.org/popug-tasks/internal/app/tasks/repositiories/users"
	"education.org/popug-tasks/internal/app/tasks/services"
	"github.com/jackc/pgx/v5"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type manager struct {
	kafkaConsumer *kafka.Consumer
	db            *pgx.Conn
	eventHandlers map[string]func(string) error
}

func New(kafkaConsumer *kafka.Consumer, db *pgx.Conn) *manager {
	return &manager{
		kafkaConsumer: kafkaConsumer,
		db:            db,
		eventHandlers: map[string]func(string) error{
			"user.updated": func(s string) error {
				var (
					repo = users.New(db)
					user services.User
				)

				err := json.Unmarshal([]byte(s), &user)
				if err != nil {
					return err
				}

				_, err = repo.Update(context.Background(), &user)
				if err != nil {
					return err
				}

				fmt.Println("user.updated", s)
				return nil
			},
			"user.created": func(s string) error {
				var (
					repo = users.New(db)
					user services.User
				)

				err := json.Unmarshal([]byte(s), &user)
				if err != nil {
					return err
				}

				_, err = repo.Create(context.Background(), &user)
				if err != nil {
					return err
				}

				fmt.Println("user.created", s)
				return nil
			},
			"user.auth": func(s string) error {
				var (
					repo = users.New(db)
					user services.User
				)

				err := json.Unmarshal([]byte(s), &user)
				if err != nil {
					return err
				}

				_, err = repo.UpdateToken(context.Background(), &user)
				if err != nil {
					return err
				}

				fmt.Println("user.auth", s)
				return nil
			},
		},
	}
}

func (c *manager) Consume(ctx context.Context, topic string) error {
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
