package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"education.org/popug-tasks/internal/app/analytics/services"
)

type tasksService interface {
	Create(ctx context.Context, task *services.Task) (*services.Task, error)
	Complete(ctx context.Context, task *services.Task, token string) error
}

type tasksHandler struct {
	taskServ tasksService
}

func New(service tasksService) *tasksHandler {
	return &tasksHandler{
		taskServ: service,
	}
}

func (b *tasksHandler) Create(w http.ResponseWriter, req *http.Request) {
	task := &services.Task{}

	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to read body: %v", err)))
		return
	}

	err = json.Unmarshal(bytes, task)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to read body: %v", err)))
		return
	}

	if len(task.Description) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("empty description"))
		return
	}

	task, err = b.taskServ.Create(context.Background(), task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("something failed: %v", err)))
		return
	}

	w.WriteHeader(http.StatusCreated)
	payload, _ := json.Marshal(task)
	_, _ = w.Write(payload)
}

func (b *tasksHandler) Complete(w http.ResponseWriter, req *http.Request) {
	task := services.Task{}

	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to read body: %v", err)))
		return
	}

	err = json.Unmarshal(bytes, &task)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to read body: %v", err)))
		return
	}

	token := req.Header.Get("x-user-token")

	err = b.taskServ.Complete(context.Background(), &task, token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("something failed: %v", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	payload, _ := json.Marshal(nil)
	_, _ = w.Write(payload)
}
