package tasks

import (
	"context"
	"education.org/popug-tasks/internal/app/account/services"
	"encoding/json"
)

type tasksRepo interface {
	Create(ctx context.Context, task *services.Task) (*services.Task, error)
	Close(ctx context.Context, task *services.Task) error
	Update(ctx context.Context, task *services.Task) error
}

type tasksHandler struct {
	tasksRepo tasksRepo
}

func New(service tasksRepo) *tasksHandler {
	return &tasksHandler{
		tasksRepo: service,
	}
}

func (b *tasksHandler) Create() func(payload string) error {
	return func(payload string) error {
		task := &services.Task{}
		err := json.Unmarshal([]byte(payload), task)
		if err != nil {
			return err
		}

		_, err = b.tasksRepo.Create(context.Background(), task)
		return err
	}
}

func (b *tasksHandler) Complete() func(payload string) error {
	return func(payload string) error {
		task := &services.Task{}
		err := json.Unmarshal([]byte(payload), task)
		if err != nil {
			return err
		}

		_, err = b.tasksRepo.Create(context.Background(), task)
		return err
	}
}
