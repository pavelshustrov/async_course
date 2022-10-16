package tasks

import (
	"context"
	"encoding/json"

	"education.org/popug-tasks/internal/app/tasks/services"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	RandomUser = `select public_id from users where role = 'worker' order by random() limit 1`

	CreateNewTask = `insert into tasks(public_id, title, description, status, assignee_id) values ($1, $2, $3, $4, $5)`
	CloseTask     = `update tasks set status = 'CLOSED' where public_id = $1`
	FindOpenTask  = `select public_id from tasks where status = 'open' for update`
	UpdateTask    = `update tasks set assignee_id = $1 where public_id = $2`

	AddNewJob = `insert into jobs(event_name, event_version, payload) values ($1, $2, $3)`
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

func (r *repo) Create(ctx context.Context, task *services.Task) (*services.Task, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx, RandomUser).Scan(&task.AssigneeID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, CreateNewTask, task.PublicID, task.Title, task.Description, task.Status, task.AssigneeID)
	if err != nil {
		return nil, err
	}

	// send message

	createdNewTaskEventPayload, _ := json.Marshal(map[string]interface{}{
		"public_id":   task.PublicID,
		"assignee_id": task.AssigneeID,
		"price_fee":   task.Price.Fee,
		"price_award": task.Price.Award,
	})

	_, err = tx.Exec(ctx, AddNewJob, "task.created", "v1", createdNewTaskEventPayload)
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

	// send message

	createdNewTaskEventPayload, _ := json.Marshal(map[string]interface{}{
		"public_id":   task.PublicID,
		"assignee_id": task.AssigneeID,
	})

	_, err = tx.Exec(ctx, AddNewJob, "task.closed", "v1", createdNewTaskEventPayload)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *repo) Reassign(ctx context.Context) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, FindOpenTask)
	if err != nil {
		return err
	}

	defer rows.Close()

	var openTaskPublicIDs []string
	for rows.Next() {
		var publicID string
		if err := rows.Scan(&publicID); err != nil {
			return err
		}

		openTaskPublicIDs = append(openTaskPublicIDs, publicID)
	}

	for _, openTaskPublicID := range openTaskPublicIDs {
		var userID string
		err = tx.QueryRow(ctx, RandomUser).Scan(&userID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, UpdateTask, userID, openTaskPublicID)
		if err != nil {
			return err
		}

		// send message

		createdNewTaskEventPayload, _ := json.Marshal(map[string]interface{}{
			"public_id":   openTaskPublicID,
			"assignee_id": userID,
		})

		_, err = tx.Exec(ctx, AddNewJob, "task.reassigned", "v1", createdNewTaskEventPayload)
		if err != nil {
			return err
		}

	}

	return tx.Commit(ctx)
}
