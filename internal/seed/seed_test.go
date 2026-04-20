package seed

import "testing"

func TestRetrosheetYearDateBounds(t *testing.T) {
	start, end := retrosheetYearDateBounds(1980)
	if start != "19800101" {
		t.Fatalf("expected start 19800101, got %s", start)
	}
	if end != "19810101" {
		t.Fatalf("expected end 19810101, got %s", end)
	}
}
