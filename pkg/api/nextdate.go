package api

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	dateLayout = "20060102"
	maxDays    = 400
)

var (
	ErrEmptyRule       = errors.New("empty repeat rule")
	ErrInvalidFormat   = errors.New("invalid repeat format")
	ErrInvalidDate     = errors.New("invalid date")
	ErrInvalidDay      = errors.New("invalid day")
	ErrInvalidMonth    = errors.New("invalid month")
	ErrInvalidWeekday  = errors.New("invalid weekday")
	ErrUnsupportedRule = errors.New("unsupported repeat rule")
	ErrDaysOutOfRange  = errors.New("days value out of range")
)

// NextDate вычисляет следующую дату выполнения задачи
func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	if repeat == "" {
		return "", ErrEmptyRule
	}

	date, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		return "", ErrInvalidDate
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", ErrInvalidFormat
	}

	switch parts[0] {
	case "d":
		return handleDailyRule(now, date, parts)
	case "y":
		return handleYearlyRule(now, date)
	case "w":
		return handleWeeklyRule(now, date, parts)
	case "m":
		return handleMonthlyRule(now, date, parts)
	default:
		return "", ErrUnsupportedRule
	}
}

// handleDailyRule обрабатывает правило с интервалом в днях
func handleDailyRule(now, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", ErrInvalidFormat
	}

	days, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", ErrInvalidFormat
	}

	if days <= 0 || days > maxDays {
		return "", ErrDaysOutOfRange
	}

	for {
		date = date.AddDate(0, 0, days)
		if afterNow(date, now) {
			break
		}
	}

	return date.Format(dateLayout), nil
}

// handleYearlyRule обрабатывает ежегодное правило
func handleYearlyRule(now, date time.Time) (string, error) {
	for {
		date = date.AddDate(1, 0, 0)
		if afterNow(date, now) {
			break
		}
	}

	// Обработка 29 февраля
	if date.Month() == time.February && date.Day() == 29 {
		if !isLeap(date.Year()) {
			date = time.Date(date.Year(), time.March, 1, 0, 0, 0, 0, time.UTC)
		}
	}

	return date.Format(dateLayout), nil
}

// handleWeeklyRule обрабатывает правило для дней недели
func handleWeeklyRule(now, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", ErrInvalidFormat
	}

	weekdays, err := parseWeekdays(parts[1])
	if err != nil {
		return "", err
	}

	for {
		date = date.AddDate(0, 0, 1)
		if afterNow(date, now) {
			weekday := int(date.Weekday())
			if weekday == 0 {
				weekday = 7 // Воскресенье = 7
			}
			if weekdays[weekday] {
				break
			}
		}
	}

	return date.Format(dateLayout), nil
}

// handleMonthlyRule обрабатывает правило для дней месяца
func handleMonthlyRule(now, date time.Time, parts []string) (string, error) {
	if len(parts) < 2 || len(parts) > 3 {
		return "", ErrInvalidFormat
	}

	days, err := parseDays(parts[1])
	if err != nil {
		return "", err
	}

	var months [13]bool
	if len(parts) == 3 {
		months, err = parseMonths(parts[2])
		if err != nil {
			return "", err
		}
	} else {
		for i := 1; i <= 12; i++ {
			months[i] = true
		}
	}

	for {
		date = date.AddDate(0, 0, 1)
		if afterNow(date, now) {
			day := date.Day()
			month := int(date.Month())
			lastDay := lastDayOfMonth(date.Year(), month)

			// Проверяем специальные дни (-1, -2)
			if days[32] && day == lastDay { // -1
				if months[month] {
					break
				}
			} else if days[33] && day == lastDay-1 { // -2
				if months[month] {
					break
				}
			} else if day <= 31 && days[day] && months[month] {
				break
			}
		}
	}

	return date.Format(dateLayout), nil
}

// parseWeekdays разбирает строку с днями недели
func parseWeekdays(s string) ([8]bool, error) {
	var weekdays [8]bool // 1-7, 0 не используется

	days := strings.Split(s, ",")
	for _, dayStr := range days {
		day, err := strconv.Atoi(strings.TrimSpace(dayStr))
		if err != nil || day < 1 || day > 7 {
			return weekdays, ErrInvalidWeekday
		}
		weekdays[day] = true
	}

	return weekdays, nil
}

// parseDays разбирает строку с днями месяца
func parseDays(s string) ([34]bool, error) {
	var days [34]bool // 1-31, 32=-1, 33=-2

	dayStrs := strings.Split(s, ",")
	for _, dayStr := range dayStrs {
		dayStr = strings.TrimSpace(dayStr)
		if dayStr == "-1" {
			days[32] = true // -1 (последний день)
		} else if dayStr == "-2" {
			days[33] = true // -2 (предпоследний день)
		} else {
			day, err := strconv.Atoi(dayStr)
			if err != nil || day < 1 || day > 31 {
				return days, ErrInvalidDay
			}
			days[day] = true
		}
	}

	return days, nil
}

// parseMonths разбирает строку с месяцами
func parseMonths(s string) ([13]bool, error) {
	var months [13]bool // 1-12, 0 не используется

	monthStrs := strings.Split(s, ",")
	for _, monthStr := range monthStrs {
		month, err := strconv.Atoi(strings.TrimSpace(monthStr))
		if err != nil || month < 1 || month > 12 {
			return months, ErrInvalidMonth
		}
		months[month] = true
	}

	return months, nil
}

// afterNow проверяет, что date после now (без учета времени)
func afterNow(date, now time.Time) bool {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return date.After(now)
}

// lastDayOfMonth возвращает последний день месяца
func lastDayOfMonth(year, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// isLeap проверяет, является ли год високосным
func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
