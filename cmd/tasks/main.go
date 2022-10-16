package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"education.org/popug-tasks/internal/app/database"
	barHandler "education.org/popug-tasks/internal/app/tasks/handlers/tasks"
	tasksRepo "education.org/popug-tasks/internal/app/tasks/repositiories/tasks"
	userRepo "education.org/popug-tasks/internal/app/tasks/repositiories/users"
	"education.org/popug-tasks/internal/app/tasks/services"
	tasksService "education.org/popug-tasks/internal/app/tasks/services/tasks"
	"education.org/popug-tasks/internal/app/transport/kafka/consumer"
	"education.org/popug-tasks/internal/app/transport/kafka/producer"
	"education.org/popug-tasks/package/event_schema_registry/spec"
	queue_manager_consumer "education.org/popug-tasks/package/queue_manager/consumer"
	queue_manager_producer "education.org/popug-tasks/package/queue_manager/producer"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const AppPrefix = "tasks"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println(AppPrefix)

	tasksRepository := tasksRepo.New(database.ConnectByName(AppPrefix))
	userRepository := userRepo.New(database.ConnectByName(AppPrefix))
	tasksServ := tasksService.New(tasksRepository, userRepository)
	tasksHandler := barHandler.New(tasksServ)

	kafkaConsumer := consumer.MustNewConsumer("tasks-group-id-5")
	manager := queue_manager_consumer.New(
		kafkaConsumer,
		database.ConnectByName(AppPrefix),
		getHandlers(userRepository),
	)

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			err := manager.Consume(cancelCtx, "usermanagmnet-streaming")
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
						return "be.tasks"
					}
					return "tasks-streaming"
				},
				"task",
				"task",
			)
			if err != nil {
				fmt.Println("error kafka producer: ", err.Error())
			}
			time.Sleep(time.Second)
		}
	}()

	router := mux.NewRouter()

	router.
		HandleFunc("/tasks", tasksHandler.Create).
		Methods(http.MethodPost)

	router.
		HandleFunc("/tasks/close", tasksHandler.Complete).
		Methods(http.MethodPut)

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

func getHandlers(repo interface {
	Create(ctx context.Context, user *services.User) (*services.User, error)
	Update(ctx context.Context, user *services.User) (*services.User, error)
	UpdateToken(ctx context.Context, user *services.User) (*services.User, error)
},
) map[string]func(string) error {
	return map[string]func(string) error{
		"user.updated": func(s string) error {
			var (
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
	}
}
