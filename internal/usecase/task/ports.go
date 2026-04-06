package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	CreateRule(ctx context.Context, rule *taskdomain.RecurrenceRule) (*taskdomain.RecurrenceRule, error)
	CreateTasksBatch(ctx context.Context, tasks []taskdomain.Task) ([]taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, start, end time.Time, limit, offset int) ([]taskdomain.Task, int64, error)
	WithinTransaction(ctx context.Context, fn func(repo Repository) error) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) ([]taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, start, end time.Time, limit, offset int) ([]taskdomain.Task, int64, error)
}

type CreateInput struct {
	Title          string
	Description    string
	Status         taskdomain.Status
	ScheduledDate  time.Time
	RecurrenceRule *taskdomain.RecurrenceRule
}

type UpdateInput struct {
	Title         string
	Description   string
	Status        taskdomain.Status
	ScheduledDate time.Time
}
