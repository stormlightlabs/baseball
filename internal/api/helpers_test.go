package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stormlightlabs.org/baseball/internal/core"
)

func TestWriteError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	err := core.NewNotFoundError("player", "foo")

	writeError(w, err)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("expected Content-Type application/json; charset=utf-8, got %q", ct)
	}
}

func TestWriteError_InternalServerError(t *testing.T) {
	w := httptest.NewRecorder()
	err := errors.New("boom")

	writeError(w, err)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("expected Content-Type application/json; charset=utf-8, got %q", ct)
	}
}

func TestParseStatProviderAlias_AcceptsBigFlyAliases(t *testing.T) {
	t.Parallel()

	aliases := []string{"bigfly", "big-fly", "big_fly", "big fly", "bf", "BF", " Big Fly "}
	for _, alias := range aliases {
		alias := alias
		t.Run(alias, func(t *testing.T) {
			t.Parallel()

			provider, err := parseStatProviderAlias(alias)
			if err != nil {
				t.Fatalf("expected alias %q to be accepted, got error: %v", alias, err)
			}
			if provider == nil {
				t.Fatalf("expected provider for alias %q", alias)
			}
			if *provider != core.StatProviderBigFly {
				t.Fatalf("expected provider %q, got %q", core.StatProviderBigFly, *provider)
			}
		})
	}
}

func TestParseStatProviderAlias_RejectsUnsupportedAliases(t *testing.T) {
	t.Parallel()

	aliases := []string{"internal", "fangraphs", "unknown", "xyz"}
	for _, alias := range aliases {
		alias := alias
		t.Run(alias, func(t *testing.T) {
			t.Parallel()

			provider, err := parseStatProviderAlias(alias)
			if err == nil {
				t.Fatalf("expected alias %q to be rejected", alias)
			}
			if provider != nil {
				t.Fatalf("expected nil provider for rejected alias %q", alias)
			}
			if !strings.Contains(err.Error(), "supported aliases") {
				t.Fatalf("expected supported aliases message, got: %v", err)
			}
		})
	}
}
