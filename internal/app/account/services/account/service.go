package account

import (
	"context"
	"education.org/popug-tasks/internal/app/account/services"
	"github.com/google/uuid"
)

type TaskRepo interface {
	FindByPublicID(ctx context.Context, task *services.Task) (*services.Task, error)
}

type AccountRepo interface {
	FindAccountByUserID(ctx context.Context, record *services.Account) (*services.Account, error)
	UpdateAccount(ctx context.Context, record *services.AuditRecord) error
}

type service struct {
	taskRepo TaskRepo
	accRepo  AccountRepo
}

func New(taskRepo TaskRepo) *service {
	return &service{
		taskRepo: taskRepo,
	}
}

func (s *service) Close(ctx context.Context, taskClosed *services.TaskClosed) error {
	task := &services.Task{
		PublicID: taskClosed.PublicID,
	}
	task, err := s.taskRepo.FindByPublicID(ctx, task)
	if err != nil {
		return err
	}

	account := &services.Account{
		UserID: taskClosed.AssigneeID,
	}

	account, err = s.accRepo.FindAccountByUserID(ctx, account)
	if err != nil {
		return err
	}

	record := services.AuditRecord{
		PublicID:  uuid.New().String(),
		AccountID: account.ID,
		EventID:   taskClosed.EventID,
		Debit:     task.PriceAward,
		Credit:    0,
		TaskID:    task.PublicID,
	}

	return s.accRepo.UpdateAccount(ctx, &record)
}
