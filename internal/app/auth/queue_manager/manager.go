package queue_manager

import (
	"context"
	"github.com/jackc/pgx/v5"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type manager struct {
	kafkaProducer *kafka.Producer
	db            *pgx.Conn
}

func New(kafkaProducer *kafka.Producer, db *pgx.Conn) *manager {
	return &manager{
		kafkaProducer: kafkaProducer,
		db:            db,
	}
}

func (c *manager) Produce(ctx context.Context, topic string) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := c.process(ctx, topic); err != nil {
				return err
			}
		}
	}
}

func (c *manager) process(ctx context.Context, topic string) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `select id, event, payload from jobs where deleted_at is null limit 10`)
	if err != nil {
		return err
	}

	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		var event, payload string

		err := rows.Scan(&id, &event, &payload)
		if err != nil {
			return err
		}

		ids = append(ids, id)

		err = c.kafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(event),
			Value:          []byte(payload),
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
