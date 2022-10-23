package accounts

import (
	"context"
	"education.org/popug-tasks/internal/app/account/services"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	AddAuditRecord      = `insert into audit_log(public_id, debit,credit, event_id, task_id) values ($1, $2, $3, $4,$5) on conflict(event_id) fail`
	UpdateAccount       = `update accounts set amount = amount + $1 - $2 where account_id = $3`
	FindAccountByUserID = `select id, public_id, user_id, amount from accounts where user_id = $1`
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

func (r *repo) FindAccountByUserID(ctx context.Context, record *services.Account) (*services.Account, error) {
	err := r.db.QueryRow(ctx, FindAccountByUserID, record.UserID).
		Scan(
			&record.ID,
			&record.PublicID,
			&record.UserID,
			&record.Amount,
		)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (r *repo) UpdateAccount(ctx context.Context, record *services.AuditRecord) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, AddAuditRecord, record.PublicID, record.Debit, record.Credit, record.EventID, record.TaskID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, UpdateAccount, record.AccountID, record.Debit, record.Credit)
	if err != nil {
		return err
	}

	// send message
	return tx.Commit(ctx)
}
