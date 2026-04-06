package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

type Repository struct {
	pool *pgxpool.Pool
	db   DB
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
		db:   pool,
	}
}

func (r *Repository) WithinTransaction(ctx context.Context, fn func(repo taskusecase.Repository) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := &Repository{
		pool: r.pool,
		db:   tx,
	}

	if err := fn(txRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) CreateRule(ctx context.Context, rule *taskdomain.RecurrenceRule) (*taskdomain.RecurrenceRule, error) {
	const query = `
		INSERT INTO task_recurrence_rules 
			(rule_type_code, interval_days, monthly_day, parity_type_code, valid_until, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		rule.RuleType,
		rule.IntervalDays,
		rule.MonthlyDay,
		rule.Parity,
		rule.ValidUntil,
		rule.CreatedAt,
	).Scan(&rule.ID, &rule.CreatedAt)

	if err != nil {
		return nil, err
	}

	if rule.RuleType == taskdomain.RuleTypeSpecificDates && len(rule.SpecificDates) > 0 {
		batch := &pgx.Batch{}
		for _, d := range rule.SpecificDates {
			batch.Queue(`INSERT INTO task_recurrence_dates (rule_id, exact_date) VALUES ($1, $2)`, rule.ID, d)
		}

		br := r.db.SendBatch(ctx, batch)
		if err := br.Close(); err != nil { // Выполнение батча
			return nil, err
		}
	}

	return rule, nil
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (recurrence_rule_id, title, description, status, scheduled_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, recurrence_rule_id, title, description, status, scheduled_date, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		task.RecurrenceRuleID, task.Title, task.Description,
		task.Status, task.ScheduledDate, task.CreatedAt, task.UpdatedAt,
	)

	return scanTask(row)
}

func (r *Repository) CreateTasksBatch(ctx context.Context, tasks []taskdomain.Task) ([]taskdomain.Task, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	batch := &pgx.Batch{}
	const query = `
		INSERT INTO tasks (recurrence_rule_id, title, description, status, scheduled_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (recurrence_rule_id, scheduled_date) WHERE recurrence_rule_id IS NOT NULL 
		DO NOTHING
		RETURNING id, recurrence_rule_id, title, description, status, scheduled_date, created_at, updated_at
	`

	for _, task := range tasks {
		batch.Queue(query,
			task.RecurrenceRuleID, task.Title, task.Description,
			task.Status, task.ScheduledDate, task.CreatedAt, task.UpdatedAt,
		)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	var created []taskdomain.Task
	for i := 0; i < len(tasks); i++ {
		row := br.QueryRow() // Читаем результат каждого запроса из батча
		task, err := scanTask(row)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// ErrNoRows здесь означает, что сработал ON CONFLICT DO NOTHING (дубликат)
				continue
			}
			return nil, err
		}
		created = append(created, *task)
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, recurrence_rule_id, title, description, status, scheduled_date, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			scheduled_date = $4,
			updated_at = $5
		WHERE id = $6
		RETURNING id, recurrence_rule_id, title, description, status, scheduled_date, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		task.Title, task.Description, task.Status,
		task.ScheduledDate, task.UpdatedAt, task.ID,
	)

	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context, start, end time.Time, limit, offset int) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, recurrence_rule_id, title, description, status, scheduled_date, created_at, updated_at
		FROM tasks
		WHERE scheduled_date >= $1 AND scheduled_date <= $2
		ORDER BY scheduled_date ASC, id ASC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, start, end, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []taskdomain.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
	)

	// Порядок должен строго соответствовать SELECT/RETURNING
	if err := scanner.Scan(
		&task.ID,
		&task.RecurrenceRuleID,
		&task.Title,
		&task.Description,
		&status,
		&task.ScheduledDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	return &task, nil
}
