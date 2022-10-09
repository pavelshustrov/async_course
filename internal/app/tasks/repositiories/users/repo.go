package users

import (
	"context"

	"education.org/popug-tasks/internal/app/tasks/services"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	CreateNewUser = `insert into users(public_id, role) values ($1, $2)`

	UpdateUser = `update users set role = $1 where public_id = $2`

	UpdateTokenUser = `update users set token = $1 where public_id = $2`
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

func (r *repo) Create(ctx context.Context, user *services.User) (*services.User, error) {
	_, err := r.db.Exec(ctx, CreateNewUser, user.PublicID, user.Role)

	return user, err
}

func (r *repo) Update(ctx context.Context, user *services.User) (*services.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, UpdateUser, user.Role, user.PublicID)
	if err != nil {
		return nil, err
	}

	return user, tx.Commit(ctx)
}

func (r *repo) UpdateToken(ctx context.Context, user *services.User) (*services.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, UpdateTokenUser, user.Token, user.PublicID)
	if err != nil {
		return nil, err
	}

	return user, tx.Commit(ctx)
}
