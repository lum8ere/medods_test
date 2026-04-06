package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) ([]taskdomain.Task, error) {
	normalized, err := s.validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	if normalized.RecurrenceRule == nil {
		model := &taskdomain.Task{
			Title:         normalized.Title,
			Description:   normalized.Description,
			Status:        normalized.Status,
			ScheduledDate: normalized.ScheduledDate,
			CreatedAt:     s.now(),
			UpdatedAt:     s.now(),
		}
		created, err := s.repo.Create(ctx, model)
		if err != nil {
			return nil, err
		}

		return []taskdomain.Task{*created}, nil
	}

	var createdTasks []taskdomain.Task

	err = s.repo.WithinTransaction(ctx, func(txRepo Repository) error {
		rule := normalized.RecurrenceRule
		rule.CreatedAt = s.now()

		savedRule, err := txRepo.CreateRule(ctx, rule)
		if err != nil {
			return err
		}

		start := s.now()
		end := start.AddDate(0, 1, 0)
		dates := taskdomain.GenerateDates(*savedRule, start, end)

		if len(dates) == 0 {
			return fmt.Errorf("%w: no dates generated for this rule", ErrInvalidInput)
		}

		// Массовое создание задач
		tasks := make([]taskdomain.Task, 0, len(dates))
		for _, d := range dates {
			tasks = append(tasks, taskdomain.Task{
				RecurrenceRuleID: &savedRule.ID,
				Title:            normalized.Title,
				Description:      normalized.Description,
				Status:           taskdomain.StatusNew,
				ScheduledDate:    d,
				CreatedAt:        s.now(),
				UpdatedAt:        s.now(),
			})
		}

		created, err := txRepo.CreateTasksBatch(ctx, tasks)
		if err != nil {
			return err
		}
		createdTasks = created
		return nil
	})

	return createdTasks, err
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, start, end time.Time) ([]taskdomain.Task, error) {
	if start.IsZero() {
		start = s.now()
	}

	if end.IsZero() {
		end = start.AddDate(0, 0, 7)
	}

	return s.repo.List(ctx, start, end)
}

func (s *Service) validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.RecurrenceRule == nil {
		if input.ScheduledDate.IsZero() {
			input.ScheduledDate = s.now()
		}
	} else {
		if err := validateRecurrence(input.RecurrenceRule); err != nil {
			return CreateInput{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func validateRecurrence(rule *taskdomain.RecurrenceRule) error {
	if rule == nil {
		return nil
	}

	switch rule.RuleType {
	case taskdomain.RuleTypeDaily:
		if rule.IntervalDays == nil || *rule.IntervalDays <= 0 {
			return fmt.Errorf("interval_days is required for daily rule and must be > 0")
		}
	case taskdomain.RuleTypeMonthlyDay:
		if rule.MonthlyDay == nil || *rule.MonthlyDay < 1 || *rule.MonthlyDay > 31 {
			return fmt.Errorf("monthly_day is required for monthly rule and must be between 1 and 31")
		}
	case taskdomain.RuleTypeParity:
		if rule.Parity == nil {
			return fmt.Errorf("parity (even/odd) is required for parity rule")
		}
	case taskdomain.RuleTypeSpecificDates:
		if len(rule.SpecificDates) == 0 {
			return fmt.Errorf("at least one specific date is required")
		}
	default:
		return fmt.Errorf("unknown recurrence rule type: %s", rule.RuleType)
	}

	return nil
}
