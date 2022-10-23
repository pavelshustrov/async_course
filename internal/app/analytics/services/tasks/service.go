package tasks

import (
	"context"
	"education.org/popug-tasks/internal/app/tasks/services"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
)

//go:generate mockery --name=Repository
type TasksRepo interface {
	Create(ctx context.Context, task *services.Task) (*services.Task, error)
	Close(ctx context.Context, user *services.Task) error
}

type UsersRepo interface {
	FindByToken(ctx context.Context, token string) (*services.User, error)
}

type service struct {
	tasksRepo TasksRepo
	usersRepo UsersRepo
}

func New(repo TasksRepo) *service {
	return &service{
		tasksRepo: repo,
	}
}

// Create
func (s *service) Create(ctx context.Context, task *services.Task) (*services.Task, error) {

	task.PublicID = uuid.New().String()
	task.Status = services.StatusOpen
	task.Price = services.Price{
		Fee:   -rand.Intn(11) - 10, // [0; 11) (-11;0] (-21;-10]
		Award: 20 + rand.Intn(21),
	}

	return s.tasksRepo.Create(ctx, task)
}

// Complete
func (s *service) Complete(ctx context.Context, task *services.Task, token string) error {
	user, err := s.usersRepo.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	if user.PublicID != task.AssigneeID {
		return fmt.Errorf("403 UNfobidden")
	}

	return s.tasksRepo.Close(ctx, task)
}
