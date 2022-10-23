package producer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type Validator interface {
	Validate(bytes []byte, domain, eventName string, version string) error
}

type producer struct {
	validator     Validator
	kafkaProducer *kafka.Producer
	db            *pgx.Conn
}

func New(kafkaProducer *kafka.Producer, db *pgx.Conn, validator Validator) *producer {
	return &producer{
		validator:     validator,
		kafkaProducer: kafkaProducer,
		db:            db,
	}
}

func (c *producer) Produce(ctx context.Context, router func(string) string, producerName, domain string) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := c.process(ctx, router, producerName, domain); err != nil {
				return err
			}
		}
	}
}

func (c *producer) process(ctx context.Context, router func(string) string, producerName, domain string) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `select id, event_name, event_version, payload from jobs where deleted_at is null limit 10`)
	if err != nil {
		return err
	}

	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		var event, version, payload string

		err := rows.Scan(&id, &event, &version, &payload)
		if err != nil {
			return err
		}

		err = c.validator.Validate([]byte(payload), domain, event, version)
		if err != nil {
			return err
		}

		ids = append(ids, id)

		topic := router(event)
		err = c.kafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Headers: []kafka.Header{
				{
					Key:   "event_name",
					Value: []byte(event),
				},
				{
					Key:   "event_version",
					Value: []byte(version),
				},
				{
					Key:   "event_id",
					Value: []byte(uuid.New().String()),
				},
				{
					Key:   "producer",
					Value: []byte(producerName),
				},
				{
					Key:   "created_at",
					Value: []byte(time.Now().Format(time.Layout)),
				},
			},
			Key:   []byte("payload"),
			Value: []byte(payload),
		}, nil)
		if err != nil {
			return err
		}

	}

	rows.Close()

	for _, id := range ids {
		_, err = tx.Exec(ctx, `update jobs set deleted_at = now() where id = $1`, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
