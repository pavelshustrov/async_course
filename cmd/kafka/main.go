package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("test kafka")
	var (
		topic     = "auth"
		produceCh = make(chan struct{}, 0)
		consumeCh = make(chan struct{}, 0)
	)

	go func() {
		defer func() {
			produceCh <- struct{}{}
		}()

		p, err := kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers": os.Getenv("bootstrap.servers"),
			//"group.id":          "go_example_group_1",
			//"sasl.mechanisms":   conf["sasl.mechanisms"],
			//"security.protocol": conf["security.protocol"],
			//"sasl.username":     conf["sasl.username"],
			//"sasl.password":     conf["sasl.password"]
		})
		if err != nil {
			panic(fmt.Errorf("kafka producer creation err: %w", err))
		}

		err = p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte("key"),
			Value:          []byte("value"),
		}, nil)
		if err != nil {
			panic(fmt.Errorf("kafka producer send message err: %w", err))
		}

		CreateTopic(p, "auth")
		CreateTopic(p, "tasks")

	}()

	go func() {
		defer func() {
			consumeCh <- struct{}{}
		}()
		c, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": os.Getenv("bootstrap.servers"),
			"group.id":          "go_example_group_1",
			"auto.offset.reset": "earliest",
			//"sasl.mechanisms":   conf["sasl.mechanisms"],
			//"security.protocol": conf["security.protocol"],
			//"sasl.username":     conf["sasl.username"],
			//"sasl.password":     conf["sasl.password"]
		})
		if err != nil {
			panic(fmt.Errorf("kafka producer creation err: %w", err))
		}

		defer c.Close()

		err = c.Subscribe(topic, nil)
		for {
			message, err := c.ReadMessage(100 * time.Millisecond)
			if err != nil {
				continue
			}

			fmt.Printf("message received: topic=%s, value=%s\n", string(message.Key), string(message.Value))
			return
		}

	}()

	<-produceCh
	<-consumeCh
}

// CreateTopic creates a topic using the Admin Client API
func CreateTopic(p *kafka.Producer, topic string) {
	a, err := kafka.NewAdminClientFromProducer(p)
	if err != nil {
		fmt.Printf("Failed to create new admin client from producer: %s", err)
		os.Exit(1)
	}
	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Create topics on cluster.
	// Set Admin options to wait up to 60s for the operation to finish on the remote cluster
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		fmt.Printf("ParseDuration(60s): %s", err)
		os.Exit(1)
	}
	//a.DeleteTopics(ctx, []string{topic},
	//	kafka.SetAdminOperationTimeout(maxDur))
	results, err := a.CreateTopics(
		ctx,
		// Multiple topics can be created simultaneously
		// by providing more TopicSpecification structs here.
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1}},
		// Admin options
		kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Admin Client request error: %v\n", err)
		os.Exit(1)
	}
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			fmt.Printf("Failed to create topic: %v\n", result.Error)
			os.Exit(1)
		}
		fmt.Printf("%v\n", result)
	}
	a.Close()

}
