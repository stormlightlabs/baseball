package etl

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/seed"
)

const defaultRetrosheetYears = "2023-2025"

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

func retrosheetEraNames() []string {
	eras := seed.GetAllEras()
	names := make([]string, 0, len(eras))
	for _, era := range eras {
		names = append(names, era.ShortName)
	}
	return names
}

func retrosheetEraHelp() string {
	return fmt.Sprintf(
		"Load data for a specific era (%s)",
		strings.Join(retrosheetEraNames(), ", "),
	)
}

func unknownEraError(era string) error {
	return fmt.Errorf(
		"unknown era %q. Valid eras: %s",
		era,
		strings.Join(retrosheetEraNames(), ", "),
	)
}

func normalizeEraFlag(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func parseYearFlag(flagValue string) ([]int, error) {
	if strings.TrimSpace(flagValue) == "" {
		return nil, nil
	}

	var years []int
	tokens := strings.SplitSeq(flagValue, ",")
	for token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		if token == "all" {
			currentYear := time.Now().Year()
			for year := 1910; year <= currentYear; year++ {
				years = append(years, year)
			}
			continue
		}

		if strings.Contains(token, "-") {
			parts := strings.SplitN(token, "-", 2)
			start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid year in range: %s", parts[0])
			}
			end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid year in range: %s", parts[1])
			}
			if end < start {
				return nil, fmt.Errorf("invalid range %s: end before start", token)
			}
			for year := start; year <= end; year++ {
				years = append(years, year)
			}
			continue
		}

		year, err := strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("invalid year: %s", token)
		}
		years = append(years, year)
	}

	if len(years) == 0 {
		return nil, nil
	}

	sort.Ints(years)
	years = uniqueInts(years)
	return years, nil
}

func uniqueInts(values []int) []int {
	if len(values) == 0 {
		return values
	}

	result := make([]int, 0, len(values))
	prev := values[0]
	result = append(result, prev)

	for _, v := range values[1:] {
		if v == prev {
			continue
		}
		result = append(result, v)
		prev = v
	}

	return result
}

func resolveDataRoot(cmd *cobra.Command) string {
	if cmd == nil {
		return seed.ResolveDataRoot("")
	}

	value, err := cmd.Flags().GetString("data-root")
	if err == nil {
		return seed.ResolveDataRoot(value)
	}
	return seed.ResolveDataRoot("")
}

func quoteIdentifier(id string) string {
	return `"` + strings.ReplaceAll(id, `"`, `""`) + `"`
}
