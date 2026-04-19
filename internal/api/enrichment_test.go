package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func asMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	m, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	return m
}

func TestIncludeDetailsEnrichment(t *testing.T) {
	t.Run("object root appends meta.details", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/teams?per_page=1&include=details", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var body map[string]any
		if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
			t.Fatalf("failed decoding response: %v", err)
		}

		meta := asMap(body["meta"])
		if meta == nil {
			t.Fatalf("expected response meta object")
		}

		details := asMap(meta["details"])
		if details == nil {
			t.Fatalf("expected meta.details object")
		}

		if _, ok := details["crosswalk"]; !ok {
			t.Fatalf("expected details.crosswalk to exist")
		}
	})

	t.Run("array root wraps into {data, meta.details}", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/meta/datasets?include=details", nil)
		w := httptest.NewRecorder()

		testServer.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var body map[string]any
		if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
			t.Fatalf("failed decoding wrapped array response: %v", err)
		}

		if _, ok := body["data"].([]any); !ok {
			t.Fatalf("expected wrapped data array, got %T", body["data"])
		}

		meta := asMap(body["meta"])
		if meta == nil {
			t.Fatalf("expected wrapped meta object")
		}
		details := asMap(meta["details"])
		if details == nil {
			t.Fatalf("expected wrapped meta.details object")
		}
	})
}
