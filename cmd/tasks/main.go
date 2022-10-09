package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"education.org/popug-tasks/internal/app/database"
	barHandler "education.org/popug-tasks/internal/app/tasks/handlers/tasks"
	"education.org/popug-tasks/internal/app/tasks/queue_manager"
	tasksRepo "education.org/popug-tasks/internal/app/tasks/repositiories/tasks"
	tasksService "education.org/popug-tasks/internal/app/tasks/services/tasks"
	"education.org/popug-tasks/internal/app/transport/kafka/consumer"
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

	//dbConn := database.ConnectByName(AppPrefix)

	kafkaConsumer := consumer.MustNewConsumer("tasks-group-id-5")
	manager := queue_manager.New(kafkaConsumer, database.ConnectByName(AppPrefix))

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			err := manager.Consume(cancelCtx, "auth")
			if err != nil {
				fmt.Println("error kafka consumer: ", err.Error())
			}
			time.Sleep(time.Second)
		}
	}()

	tasksRepository := tasksRepo.New(database.ConnectByName(AppPrefix))
	tasksServ := tasksService.New(tasksRepository)
	tasksHandler := barHandler.New(tasksServ)

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
