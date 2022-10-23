package tasks

import (
	"context"
	"education.org/popug-tasks/internal/app/account/services"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	CreateNewTask      = `insert into tasks(public_id, price_fee, price_award, status, assignee_id) values ($1, $2, $3, $4, $5)`
	CloseTask          = `update tasks set status = 'CLOSED' where public_id = $1`
	UpdateTask         = `update tasks set assignee_id = $1 where public_id = $2`
	FindTaskByPublicID = `select public_id, price_fee, price_award, status, assignee_id from tasks where public_id = $1`
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

func (r *repo) FindByPublicID(ctx context.Context, task *services.Task) (*services.Task, error) {
	err := r.db.QueryRow(ctx, FindTaskByPublicID, task.PublicID).
		Scan(
			&task.PublicID,
			&task.PriceFee,
			&task.PriceAward,
			&task.Status,
			&task.AssigneeID,
		)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *repo) Create(ctx context.Context, task *services.Task) (*services.Task, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, CreateNewTask, task.PublicID, task.PriceFee, task.PriceAward, task.Status, task.AssigneeID)
	if err != nil {
		return nil, err
	}

	return task, tx.Commit(ctx)
}

func (r *repo) Close(ctx context.Context, task *services.Task) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, CloseTask, task.PublicID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *repo) Update(ctx context.Context, task *services.Task) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, UpdateTask, task.AssigneeID, task.PublicID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
