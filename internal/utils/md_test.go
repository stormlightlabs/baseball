package utils

import (
	"testing"
	"time"
)

func TestParseMonthDayWindowContains(t *testing.T) {
	window, err := ParseMonthDayWindow("03-20/11-15")
	if err != nil {
		t.Fatalf("ParseMonthDayWindow returned error: %v", err)
	}

	if !window.Contains(time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("expected April date to be inside window")
	}
	if window.Contains(time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("expected December date to be outside window")
	}
}

func TestParseMonthDayWindowWraparound(t *testing.T) {
	window, err := ParseMonthDayWindow("11-15/02-15")
	if err != nil {
		t.Fatalf("ParseMonthDayWindow returned error: %v", err)
	}

	if !window.Contains(time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("expected January date to be inside wraparound window")
	}
	if window.Contains(time.Date(2026, time.July, 5, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("expected July date to be outside wraparound window")
	}
}
