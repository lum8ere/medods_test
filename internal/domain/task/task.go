package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

type RuleType string

const (
	RuleTypeDaily         RuleType = "daily"
	RuleTypeMonthlyDay    RuleType = "monthly_day"
	RuleTypeParity        RuleType = "parity"
	RuleTypeSpecificDates RuleType = "specific_dates"
)

type ParityType string

const (
	ParityEven ParityType = "even"
	ParityOdd  ParityType = "odd"
)

type RecurrenceRule struct {
	ID            int64
	RuleType      RuleType
	IntervalDays  *int
	MonthlyDay    *int
	Parity        *ParityType
	SpecificDates []time.Time
	ValidUntil    *time.Time
	CreatedAt     time.Time
}

type Task struct {
	ID               int64     `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Status           Status    `json:"status"`
	RecurrenceRuleID *int64    `json:"recurrence_rule_id,omitempty"`
	ScheduledDate    time.Time `json:"scheduled_date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
