package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// @name RecurrenceRule
type recurrenceDTO struct {
	Type          string      `json:"type" example:"daily" enums:"daily,monthly_day,parity,specific_dates"`
	IntervalDays  *int        `json:"interval_days" example:"2"`
	MonthlyDay    *int        `json:"monthly_day" example:"15"`
	Parity        *string     `json:"parity" example:"even" enums:"even,odd"`
	ValidUntil    *time.Time  `json:"valid_until" example:"2024-12-31T23:59:59Z"`
	SpecificDates []time.Time `json:"specific_dates"`
}

// @name TaskMutation
type taskMutationDTO struct {
	Title         string            `json:"title" example:"Обход пациентов"`
	Description   string            `json:"description" example:"Проверить состояние пациентов в 5-й палате"`
	Status        taskdomain.Status `json:"status" example:"new" enums:"new,in_progress,done" default:"new"`
	ScheduledDate time.Time         `json:"scheduled_date" example:"2024-04-06T09:00:00Z"`
	Recurrence    *recurrenceDTO    `json:"recurrence"`
}

// @name Task
type taskDTO struct {
	ID               int64             `json:"id" example:"101"`
	RecurrenceRuleID *int64            `json:"recurrence_rule_id,omitempty" example:"1"`
	Title            string            `json:"title" example:"Обход пациентов"`
	Description      string            `json:"description" example:"Проверить состояние пациентов в 5-й палате"`
	Status           taskdomain.Status `json:"status" example:"new"`
	ScheduledDate    time.Time         `json:"scheduled_date" example:"2024-04-06T09:00:00Z"`
	CreatedAt        time.Time         `json:"created_at" example:"2024-04-06T22:23:55Z"`
	UpdatedAt        time.Time         `json:"updated_at" example:"2024-04-06T22:23:55Z"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:               task.ID,
		RecurrenceRuleID: task.RecurrenceRuleID,
		Title:            task.Title,
		Description:      task.Description,
		Status:           task.Status,
		ScheduledDate:    task.ScheduledDate,
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
	}
}

// @name TaskListResponse
type taskListResponseDTO struct {
	Items      []taskDTO `json:"items"`
	TotalCount int64     `json:"total_count" example:"150"`
	Limit      int       `json:"limit" example:"20"`
	Offset     int       `json:"offset" example:"0"`
}
