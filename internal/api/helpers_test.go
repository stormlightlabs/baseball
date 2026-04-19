package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
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
