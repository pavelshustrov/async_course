package users

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"education.org/popug-tasks/internal/app/auth/services"
)

type userService interface {
	Register(ctx context.Context, user *services.User) (*services.User, error)
	Update(ctx context.Context, user *services.User) (*services.User, error)
}

type userHandler struct {
	barService userService
}

func New(service userService) *userHandler {
	return &userHandler{
		barService: service,
	}
}

func (b *userHandler) Create(w http.ResponseWriter, req *http.Request) {
	// read vars from request
	user := &services.User{}

	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to read body: %v", err)))
		return
	}

	err = json.Unmarshal(bytes, user)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to read body: %v", err)))
		return
	}

	user, err = b.barService.Register(context.Background(), user)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("something failed: %v", err)))
		return
	}

	w.WriteHeader(http.StatusCreated)
	payload, _ := json.Marshal(user)
	_, _ = w.Write(payload)
}

func (b *userHandler) Update(w http.ResponseWriter, req *http.Request) {
	user := &services.User{}

	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to read body: %v", err)))
		return
	}

	err = json.Unmarshal(bytes, user)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to read body: %v", err)))
		return
	}

	user, err = b.barService.Update(context.Background(), user)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("something failed: %v", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	payload, _ := json.Marshal(user)
	_, _ = w.Write(payload)
}
