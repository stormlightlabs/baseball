package db

import (
	"fmt"
	"strconv"
	"strings"
)

func cleanNumeric(val string) string {
	val = strings.TrimSpace(val)
	if val == "" || val == "unknown" || val == "-1" || val == "0" {
		return ""
	}
	if strings.HasSuffix(val, "?") || strings.HasPrefix(val, "<") || strings.HasPrefix(val, ">") {
		return ""
	}
	return val
}

func cleanText(val string) string {
	val = strings.TrimSpace(val)
	if val == "" || val == "unknown" {
		return ""
	}
	return val
}

// parseTime converts "0:00PM" format to "00:00:00" TIME format
func parseTime(val string) string {
	val = strings.TrimSpace(val)
	if val == "" || val == "0:00PM" {
		return ""
	}

	val = strings.ToUpper(val)
	isPM := strings.HasSuffix(val, "PM")
	val = strings.TrimSuffix(strings.TrimSuffix(val, "PM"), "AM")

	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		return ""
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 12 {
		return ""
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return ""
	}

	if isPM && hour != 12 {
		hour += 12
	} else if !isPM && hour == 12 {
		hour = 0
	}

	return fmt.Sprintf("%02d:%02d:00", hour, minute)
}

// parseBoolean converts various boolean representations
func cleanBoolean(val string) string {
	val = strings.ToLower(strings.TrimSpace(val))
	if val == "true" || val == "t" || val == "1" {
		return "true"
	}
	if val == "false" || val == "f" || val == "0" {
		return "false"
	}
	return ""
}
