package task

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateDates(t *testing.T) {
	date := func(y, m, d int) time.Time {
		return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	}

	intPtr := func(i int) *int { return &i }
	parityPtr := func(p ParityType) *ParityType { return &p }

	tests := []struct {
		name        string
		rule        RecurrenceRule
		windowStart time.Time
		windowEnd   time.Time
		want        []time.Time
	}{
		{
			name: "Daily every 2 days",
			rule: RecurrenceRule{
				RuleType:     RuleTypeDaily,
				IntervalDays: intPtr(2),
				CreatedAt:    date(2023, 10, 1),
			},
			windowStart: date(2023, 10, 1),
			windowEnd:   date(2023, 10, 5),
			want: []time.Time{
				date(2023, 10, 1),
				date(2023, 10, 3),
				date(2023, 10, 5),
			},
		},
		{
			name: "Monthly on 31st (short month handle)",
			rule: RecurrenceRule{
				RuleType:   RuleTypeMonthlyDay,
				MonthlyDay: intPtr(31),
			},
			windowStart: date(2023, 2, 1),
			windowEnd:   date(2023, 4, 1),
			want: []time.Time{
				date(2023, 2, 28), // Февраль схлопнулся до 28
				date(2023, 3, 31), // Март есть 31
			},
		},
		{
			name: "Even days of month",
			rule: RecurrenceRule{
				RuleType: RuleTypeParity,
				Parity:   parityPtr(ParityEven),
			},
			windowStart: date(2023, 10, 1),
			windowEnd:   date(2023, 10, 5),
			want: []time.Time{
				date(2023, 10, 2),
				date(2023, 10, 4),
			},
		},
		{
			name: "Specific dates",
			rule: RecurrenceRule{
				RuleType: RuleTypeSpecificDates,
				SpecificDates: []time.Time{
					date(2023, 12, 31),
					date(2024, 1, 1),
				},
			},
			windowStart: date(2023, 12, 1),
			windowEnd:   date(2024, 1, 31),
			want: []time.Time{
				date(2023, 12, 31),
				date(2024, 1, 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateDates(tt.rule, tt.windowStart, tt.windowEnd)
			assert.Equal(t, tt.want, got)
		})
	}
}
