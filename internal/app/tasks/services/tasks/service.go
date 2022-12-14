package tasks

import (
	"context"
	"fmt"
	"math/rand"

	"education.org/popug-tasks/internal/app/tasks/services"
	"github.com/google/uuid"
)

//go:generate mockery --name=Repository
type TasksRepo interface {
	Create(ctx context.Context, task *services.Task) (*services.Task, error)
	Close(ctx context.Context, user *services.Task) error
	FindByPublicID(ctx context.Context, task *services.Task) (*services.Task, error)
}

type UsersRepo interface {
	FindByToken(ctx context.Context, token string) (*services.User, error)
}

type service struct {
	tasksRepo TasksRepo
	usersRepo UsersRepo
}

func New(repo TasksRepo, usersRepo UsersRepo) *service {
	return &service{
		tasksRepo: repo,
		usersRepo: usersRepo,
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
	task.Jira, task.Title = extractJira(task.Title)
	return s.tasksRepo.Create(ctx, task)
}

func extractJira(title string) (*string, string) {
	// regexp logic, but too lazy
	return nil, title
}

// Complete
func (s *service) Complete(ctx context.Context, task *services.Task, token string) error {
	task, err := s.tasksRepo.FindByPublicID(ctx, task)
	if err != nil {
		return err
	}

	user, err := s.usersRepo.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	if user.PublicID != task.AssigneeID {
		return fmt.Errorf("403 UNfobidden")
	}

	return s.tasksRepo.Close(ctx, task)
}
