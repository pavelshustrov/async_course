package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	tasksHandler "education.org/popug-tasks/internal/app/account/handlers/tasks"
	usersHandler "education.org/popug-tasks/internal/app/account/handlers/users"
	tasksRepo "education.org/popug-tasks/internal/app/account/repositiories/tasks"
	userRepo "education.org/popug-tasks/internal/app/account/repositiories/users"
	"education.org/popug-tasks/internal/app/database"
	"education.org/popug-tasks/internal/app/transport/kafka/consumer"
	"education.org/popug-tasks/internal/app/transport/kafka/producer"
	"education.org/popug-tasks/package/event_schema_registry/spec"
	queue_manager_consumer "education.org/popug-tasks/package/queue_manager/consumer"
	queue_manager_producer "education.org/popug-tasks/package/queue_manager/producer"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const AppPrefix = "analytic"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println(AppPrefix)

	tasksRepository := tasksRepo.New(database.ConnectByName(AppPrefix))
	userRepository := userRepo.New(database.ConnectByName(AppPrefix))

	tasksHandler := tasksHandler.New(tasksRepository)
	usersHandler := usersHandler.New(userRepository)

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		kafkaConsumer := consumer.MustNewConsumer("account-group-id-5")
		manager := queue_manager_consumer.New(
			kafkaConsumer,
			database.ConnectByName(AppPrefix),
			map[string]func(string) error{
				"user.created":   usersHandler.Created(),
				"user.updated":   usersHandler.Updated(),
				"user.auth":      usersHandler.Auth(),
				"task.created":   usersHandler.Auth(),
				"task.completed": usersHandler.Auth(),
			},
		)

		for {
			err := manager.Consume(cancelCtx, "usermanagmnet-streaming")
			if err != nil {
				fmt.Println("error kafka consumer: ", err.Error())
			}
			time.Sleep(time.Second)
		}
	}()

	go func() {
		kafkaConsumer := consumer.MustNewConsumer("account-group-id-5")
		manager := queue_manager_consumer.New(
			kafkaConsumer,
			database.ConnectByName(AppPrefix),
			map[string]func(string) error{
				"task.created": tasksHandler.Create(),
				"task.closed":  tasksHandler.Complete(),
			},
		)

		for {
			err := manager.Consume(cancelCtx, "tasks-streaming")
			if err != nil {
				fmt.Println("error kafka consumer: ", err.Error())
			}
			time.Sleep(time.Second)
		}
	}()

	validator := spec.NewValidator()
	kafkaProducer := producer.MustNewProducer()
	managerProducer := queue_manager_producer.New(kafkaProducer, database.ConnectByName(AppPrefix), validator)

	go func() {
		for {
			err := managerProducer.Produce(
				cancelCtx,
				func(event string) string {
					if strings.HasPrefix(event, "be.") {
						return "be.accounts"
					}
					return "accounts-streaming"
				},
				"account",
				"account",
			)
			if err != nil {
				fmt.Println("error kafka producer: ", err.Error())
			}
			time.Sleep(time.Second)
		}
	}()

	router := mux.NewRouter()

	http.Handle("/", router)

	srv := &http.Server{
		Addr: "0.0.0.0:8089",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router, // Pass our instance of gorilla/mux in.
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
	}
	//router.HandleFunc("/tasks/reassign", tasksHandler.Handle)
	//router.HandleFunc("/tasks/all", tasksHandler.Handle)
	//router.HandleFunc("/my/tasks", tasksHandler.Handle)
}
