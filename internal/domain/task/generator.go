package task

import "time"

func GenerateDates(rule RecurrenceRule, windowStart, windowEnd time.Time) []time.Time {
	var dates []time.Time

	current := truncateToDay(windowStart)
	end := truncateToDay(windowEnd)

	if rule.ValidUntil != nil {
		validUntilDay := truncateToDay(*rule.ValidUntil)
		if end.After(validUntilDay) {
			end = validUntilDay
		}
	}

	for !current.After(end) {
		if isMatch(rule, current) {
			dates = append(dates, current)
		}
		current = current.AddDate(0, 0, 1)
	}

	return dates
}

func isMatch(rule RecurrenceRule, date time.Time) bool {
	switch rule.RuleType {
	case RuleTypeDaily:
		if rule.IntervalDays == nil || *rule.IntervalDays <= 0 {
			return false
		}
		baseDate := truncateToDay(rule.CreatedAt)
		if date.Before(baseDate) {
			return false
		}

		// Считаем разницу в днях. Если остаток от деления на интервал = 0, значит это "наш" день.
		// Пример: интервал 3. Разница 0 (сегодня) - да, разница 3 (через 3 дня) - да.
		daysDiff := int(date.Sub(baseDate).Hours() / 24)
		return daysDiff%(*rule.IntervalDays) == 0

	case RuleTypeMonthlyDay:
		if rule.MonthlyDay == nil {
			return false
		}

		targetDay := *rule.MonthlyDay
		lastDay := lastDayOfMonth(date)

		// краевой случай (например, ждем 31-го, а в феврале 28 дней)
		if targetDay > lastDay {
			targetDay = lastDay
		}
		return date.Day() == targetDay

	case RuleTypeParity:
		if rule.Parity == nil {
			return false
		}
		isEven := date.Day()%2 == 0
		if *rule.Parity == ParityEven {
			return isEven
		}
		return !isEven

	case RuleTypeSpecificDates:
		for _, specific := range rule.SpecificDates {
			if truncateToDay(specific).Equal(date) {
				return true
			}
		}
		return false
	}

	return false
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func lastDayOfMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
}
