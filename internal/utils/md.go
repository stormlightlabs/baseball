package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type MonthDay struct {
	Month int
	Day   int
}

type MonthDayWindow struct {
	Start MonthDay
	End   MonthDay
}

func ParseMonthDay(value string) (MonthDay, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return MonthDay{}, fmt.Errorf("expected MM-DD")
	}

	month, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return MonthDay{}, fmt.Errorf("month parse error: %w", err)
	}
	day, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return MonthDay{}, fmt.Errorf("day parse error: %w", err)
	}

	if month < 1 || month > 12 {
		return MonthDay{}, fmt.Errorf("month must be 1..12")
	}
	// 2001 is arbitrary; only day/month validity matters.
	if day < 1 || day > daysInMonth(month, 2001) {
		return MonthDay{}, fmt.Errorf("day %d is invalid for month %d", day, month)
	}
	return MonthDay{Month: month, Day: day}, nil
}

func ParseMonthDayWindow(value string) (MonthDayWindow, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return MonthDayWindow{}, fmt.Errorf("expected MM-DD/MM-DD")
	}

	start, err := ParseMonthDay(parts[0])
	if err != nil {
		return MonthDayWindow{}, fmt.Errorf("invalid start date: %w", err)
	}
	end, err := ParseMonthDay(parts[1])
	if err != nil {
		return MonthDayWindow{}, fmt.Errorf("invalid end date: %w", err)
	}

	return MonthDayWindow{Start: start, End: end}, nil
}

func (w MonthDayWindow) Contains(now time.Time) bool {
	current := int(now.Month())*100 + now.Day()
	start := w.Start.Month*100 + w.Start.Day
	end := w.End.Month*100 + w.End.Day

	if start <= end {
		return current >= start && current <= end
	}
	// Wraparound range, e.g. 11-01/02-15
	return current >= start || current <= end
}

func daysInMonth(month, year int) int {
	// day 0 of next month = last day of current month
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
