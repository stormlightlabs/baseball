package middleware

import (
	"expvar"
	"net/http"
	"sync"
	"time"
)

// Metrics holds global HTTP metrics exposed via expvar at /debug/vars.
// Tracks total requests, errors (5xx), per-route request counts, and latency sums.
type Metrics struct {
	// total HTTP requests across all routes
	RequestsTotal *expvar.Int
	// total 5xx responses
	ErrorsTotal *expvar.Int
	// RequestsByRoute maps route -> request count
	RequestsByRoute *expvar.Map
	// LatencyMsByRoute maps route -> cumulative latency in milliseconds
	LatencyMsByRoute *expvar.Map
}

var (
	metrics     *Metrics
	metricsOnce sync.Once
)

// GlobalMetrics returns the singleton metrics instance.
// Registers expvar metrics on first call.
func GlobalMetrics() *Metrics {
	metricsOnce.Do(func() {
		metrics = &Metrics{
			RequestsTotal:    expvar.NewInt("http_requests_total"),
			ErrorsTotal:      expvar.NewInt("http_errors_total"),
			RequestsByRoute:  expvar.NewMap("http_requests_by_route"),
			LatencyMsByRoute: expvar.NewMap("http_latency_ms_by_route"),
		}
	})
	return metrics
}

// RouteNamer extracts a route label from an HTTP request.
// Used to group requests by route in metrics (e.g., "GET /v1/players/{id}").
type RouteNamer func(r *http.Request) string

// DefaultRouteNamer returns "METHOD path" as the route label.
func DefaultRouteNamer(r *http.Request) string {
	return r.Method + " " + r.URL.Path
}

// MetricsMiddleware instruments HTTP requests with counters and latency tracking.
// Metrics are exposed via expvar at /debug/vars endpoint.
//
// Tracks:
//   - Total requests across all routes
//   - Requests per route
//   - Cumulative latency per route (derive average offline)
//   - Total 5xx errors
//
// If routeNamer is nil, uses DefaultRouteNamer.
func MetricsMiddleware(routeNamer RouteNamer) func(http.Handler) http.Handler {
	if routeNamer == nil {
		routeNamer = DefaultRouteNamer
	}
	m := GlobalMetrics()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			route := routeNamer(r)
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			latencyMs := time.Since(start).Milliseconds()

			m.RequestsTotal.Add(1)
			m.RequestsByRoute.Add(route, 1)
			m.LatencyMsByRoute.Add(route, latencyMs)

			if wrapped.statusCode >= 500 {
				m.ErrorsTotal.Add(1)
			}
		})
	}
}
