package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type recurrenceDTO struct {
	Type          string      `json:"type"`
	IntervalDays  *int        `json:"interval_days"`
	MonthlyDay    *int        `json:"monthly_day"`
	Parity        *string     `json:"parity"`
	ValidUntil    *time.Time  `json:"valid_until"`
	SpecificDates []time.Time `json:"specific_dates"`
}

type taskMutationDTO struct {
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Status        taskdomain.Status `json:"status"`
	ScheduledDate time.Time         `json:"scheduled_date"`
	Recurrence    *recurrenceDTO    `json:"recurrence"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
