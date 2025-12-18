package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/db"
)

// formatYearRange formats a slice of years into a compact string representation.
// Examples: [2020, 2021, 2022] -> "2020-2022"
//
//	[2020, 2022, 2023, 2025] -> "2020, 2022-2023, 2025"
func formatYearRange(years []int) string {
	if len(years) == 0 {
		return ""
	}

	sort.Ints(years)
	var ranges []string
	start := years[0]
	end := years[0]

	for i := 1; i < len(years); i++ {
		if years[i] == end+1 {
			end = years[i]
		} else {
			if start == end {
				ranges = append(ranges, fmt.Sprintf("%d", start))
			} else if end == start+1 {
				ranges = append(ranges, fmt.Sprintf("%d, %d", start, end))
			} else {
				ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
			}
			start = years[i]
			end = years[i]
		}
	}

	if start == end {
		ranges = append(ranges, fmt.Sprintf("%d", start))
	} else if end == start+1 {
		ranges = append(ranges, fmt.Sprintf("%d, %d", start, end))
	} else {
		ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
	}

	return strings.Join(ranges, ", ")
}

// formatYearRangeWithGaps formats a slice of years showing ranges and gaps clearly.
// Examples: [1903, 1904, 1912, 1913, 1914, 1920, 1921] -> "7 years: 1903-1904, 1912-1914, 1920-1921"
//
//	[2020, 2021, 2022, 2023, 2024, 2025] -> "6 years: 2020-2025"
//	[2020, 2023, 2025] -> "3 years: 2020, 2023, 2025"
func formatYearRangeWithGaps(years []int) string {
	if len(years) == 0 {
		return "0 years"
	}

	sort.Ints(years)
	var ranges []string
	start := years[0]
	end := years[0]

	for i := 1; i < len(years); i++ {
		if years[i] == end+1 {
			end = years[i]
		} else {
			if start == end {
				ranges = append(ranges, fmt.Sprintf("%d", start))
			} else {
				ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
			}
			start = years[i]
			end = years[i]
		}
	}

	if start == end {
		ranges = append(ranges, fmt.Sprintf("%d", start))
	} else {
		ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
	}

	rangeStr := strings.Join(ranges, ", ")
	return fmt.Sprintf("%d years: %s", len(years), rangeStr)
}

// formatLargeNumber formats a number with comma separators.
// Example: 1234567 -> "1,234,567"
func formatLargeNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result []byte
	commaIdx := len(s) % 3
	if commaIdx == 0 {
		commaIdx = 3
	}

	for i, c := range s {
		if i == commaIdx && i != 0 {
			result = append(result, ',')
			commaIdx += 3
		}
		result = append(result, byte(c))
	}

	return string(result)
}

func formatTTL(ttl time.Duration) string {
	if ttl < 0 {
		return "No expiry"
	}
	if ttl < time.Minute {
		return fmt.Sprintf("%ds", int(ttl.Seconds()))
	}
	if ttl < time.Hour {
		return fmt.Sprintf("%dm", int(ttl.Minutes()))
	}
	return fmt.Sprintf("%.1fh", ttl.Hours())
}

func formatRefresh(entry *db.DatasetRefresh) string {
	if entry == nil || entry.LastLoadedAt.IsZero() {
		return "not yet recorded"
	}

	return fmt.Sprintf("%s (%s ago, %d rows)",
		entry.LastLoadedAt.Format(time.RFC1123),
		time.Since(entry.LastLoadedAt).Round(time.Minute),
		entry.RowCount,
	)
}

func humanizeModTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	ago := time.Since(t)
	return fmt.Sprintf("%s (%s ago)", t.Format("2006-01-02 15:04"), ago.Round(time.Minute))
}
