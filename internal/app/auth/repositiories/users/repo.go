package users

import (
	"context"
	"encoding/json"

	service "education.org/popug-tasks/internal/app/auth/services"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	CreateNewUser = `insert into users(public_id, name, role, email) values ($1, $2, $3, $4)`

	UpdateUser = `update users set name = $1, role = $2, email = $3 where public_id = $4`

	// todo extract jobs to separate repository

	AddNewJob = `insert into jobs(event, payload) values ($1, $2)`
)

type Repository interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	Begin(context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type repo struct {
	db Repository
}

func New(db Repository) *repo {
	return &repo{
		db: db,
	}
}

func (r *repo) Create(ctx context.Context, user *service.User) (*service.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, CreateNewUser, user.PublicID, user.Name, user.Role, user.Email)
	if err != nil {
		return nil, err
	}

	createdUserEventPayload, _ := json.Marshal(map[string]interface{}{
		"public_id": user.PublicID,
		"role":      user.Role,
	})

	_, err = tx.Exec(ctx, AddNewJob, "user.created", createdUserEventPayload)
	if err != nil {
		return nil, err
	}

	return user, tx.Commit(ctx)
}

func (r *repo) Update(ctx context.Context, user *service.User) (*service.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, UpdateUser, user.Name, user.Role, user.Email, user.PublicID)
	if err != nil {
		return nil, err
	}

	// todo check role updated?
	createdUserEventPayload, _ := json.Marshal(map[string]interface{}{
		"public_id": user.PublicID,
		"role":      user.Role,
		"email":     user.Email,
		"Name":      user.Name,
	})

	_, err = tx.Exec(ctx, AddNewJob, "user.updated", createdUserEventPayload)
	if err != nil {
		return nil, err
	}

	return user, tx.Commit(ctx)
}
