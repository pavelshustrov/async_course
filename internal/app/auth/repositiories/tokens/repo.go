package tokens

import (
	"context"
	"encoding/json"

	"education.org/popug-tasks/internal/app/auth/services"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	FindUser = `select public_id from users where email = $1`

	CreateToken = `insert into tokens(access_token, public_id) values ($1, $2) ON CONFLICT (public_id) DO UPDATE SET access_token = EXCLUDED.access_token`

	FindUserByToken = `select public_id, name, role, email from users where public_id in (select public_id from tokens where access_token = $1)`

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

func (r *repo) Login(ctx context.Context, credentials *services.Credentials) (*services.Token, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	var publicID string

	err = tx.QueryRow(ctx, FindUser, credentials.Email).Scan(&publicID)
	if err != nil {
		return nil, err
	}

	token := services.Token{
		AccessToken: uuid.New().String(),
	}

	_, err = tx.Exec(ctx, CreateToken, token.AccessToken, publicID)
	if err != nil {
		return nil, err
	}

	authUserEventPayload, _ := json.Marshal(map[string]interface{}{
		"public_id": publicID,
		"token":     token.AccessToken,
	})

	_, err = tx.Exec(ctx, AddNewJob, "user.auth", authUserEventPayload)
	if err != nil {
		return nil, err
	}

	return &token, tx.Commit(ctx)
}

func (r *repo) Verify(ctx context.Context, token *services.Token) (*services.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	var user services.User

	err = tx.QueryRow(ctx, FindUserByToken, token.AccessToken).
		Scan(
			&user.PublicID,
			&user.Name,
			&user.Role,
			&user.Email,
		)
	if err != nil {
		return nil, err
	}

	return &user, tx.Commit(ctx)
}
