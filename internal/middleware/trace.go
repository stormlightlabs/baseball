package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

type ctxKey string

const traceIDKey ctxKey = "trace_id"

// TraceIDFromContext extracts the trace ID from the request context.
// Returns empty string if no trace ID is present.
func TraceIDFromContext(ctx context.Context) string {
	if v := ctx.Value(traceIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// newTraceID generates a random 16-byte hex trace ID.
// Falls back to timestamp-based ID if crypto randomness fails.
func newTraceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().Format("20060102T150405.000000000")
	}
	return hex.EncodeToString(b[:])
}

// TraceMiddleware ensures every HTTP request has a trace ID for correlation.
// The trace ID is:
//   - Extracted from the X-Trace-ID header if present
//   - Generated as a random 16-byte hex string if not present
//   - Stored in the request context
//   - Added to the charmbracelet/log logger context
//   - Echoed back in the X-Trace-ID response header
//
// This allows end-to-end tracing of requests across logs and services.
func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = newTraceID()
		}

		ctx := context.WithValue(r.Context(), traceIDKey, traceID)
		logger := log.FromContext(ctx)
		if logger == nil {
			logger = log.Default()
		}

		logger = logger.With("trace_id", traceID)
		ctx = log.WithContext(ctx, logger)

		w.Header().Set("X-Trace-ID", traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
