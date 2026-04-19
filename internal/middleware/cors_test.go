package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_AllowsConfiguredOrigin(t *testing.T) {
	hitNext := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitNext = true
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS([]string{"https://bigfly.tech"})(next)
	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	req.Header.Set("Origin", "https://bigfly.tech")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if !hitNext {
		t.Fatal("expected next handler to be called")
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "https://bigfly.tech" {
		t.Fatalf("unexpected allow origin: got %q", got)
	}
	if got := res.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("unexpected allow credentials header: got %q", got)
	}
}

func TestCORS_DeniesPreflightForDisallowedOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called for denied preflight")
	})

	handler := CORS([]string{"https://bigfly.tech"})(next)
	req := httptest.NewRequest(http.MethodOptions, "/v1/meta", nil)
	req.Header.Set("Origin", "https://evil.example")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("unexpected status for denied preflight: got %d", res.Code)
	}
}

func TestCORS_HandlesPreflightForAllowedOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called for preflight")
	})

	handler := CORS([]string{"https://bigfly.tech"})(next)
	req := httptest.NewRequest(http.MethodOptions, "/v1/meta", nil)
	req.Header.Set("Origin", "https://bigfly.tech")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("unexpected status for preflight: got %d", res.Code)
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "https://bigfly.tech" {
		t.Fatalf("unexpected allow origin: got %q", got)
	}
}
