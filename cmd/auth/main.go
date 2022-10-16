package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	authHandler "education.org/popug-tasks/internal/app/auth/handlers/auth"
	usersHandler "education.org/popug-tasks/internal/app/auth/handlers/users"
	authRepo "education.org/popug-tasks/internal/app/auth/repositiories/tokens"
	usersRepo "education.org/popug-tasks/internal/app/auth/repositiories/users"
	authService "education.org/popug-tasks/internal/app/auth/services/auth"
	usersService "education.org/popug-tasks/internal/app/auth/services/users"
	"education.org/popug-tasks/internal/app/database"
	"education.org/popug-tasks/internal/app/transport/kafka/producer"
	"education.org/popug-tasks/package/event_schema_registry/spec"
	queue_manager_producer "education.org/popug-tasks/package/queue_manager/producer"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const AppPrefix = "auth"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println(AppPrefix)

	//dbConn := database.ConnectByName(AppPrefix)

	validator := spec.NewValidator()
	kafkaProducer := producer.MustNewProducer()
	manager := queue_manager_producer.New(kafkaProducer, database.ConnectByName(AppPrefix), validator)

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			err := manager.Produce(
				cancelCtx,
				func(event string) string {
					if strings.HasPrefix(event, "be.") {
						return "be.auth"
					}
					return "usermanagmnet-streaming"
				},
				"auth",
				"auth",
			)
			if err != nil {
				fmt.Println("error kafka producer: ", err.Error())
			}
			time.Sleep(time.Second)
		}
	}()

	bRepo := usersRepo.New(database.ConnectByName(AppPrefix))
	authRepository := authRepo.New(database.ConnectByName(AppPrefix))

	usersServ := usersService.New(bRepo)
	authServ := authService.New(authRepository)

	userController := usersHandler.New(usersServ)
	auth := authHandler.New(authServ)

	router := mux.NewRouter()

	// create new user
	router.
		HandleFunc("/users", userController.Create).
		Methods(http.MethodPost)

	// update existing user
	router.
		HandleFunc("/users", userController.Update).
		Methods(http.MethodPatch)

	// login
	router.
		HandleFunc("/login", auth.Login).
		Methods(http.MethodPut)

	// verify
	router.
		HandleFunc("/verify", auth.Verify).
		Methods(http.MethodPut)

	http.Handle("/", router)

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router, // Pass our instance of gorilla/mux in.
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
	}

}
