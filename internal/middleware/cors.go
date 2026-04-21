package middleware

import (
	"net/http"
	"strings"
)

const defaultCORSAllowedHeaders = "Accept, Authorization, Content-Type, X-Requested-With, X-Trace-ID, X-BigFly-Client"
const defaultCORSAllowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"

// CORS returns a middleware that applies origin-based CORS headers and
// handles preflight requests.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	allowAll := false
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		if trimmed == "*" {
			allowAll = true
			continue
		}
		originSet[trimmed] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			allowedOrigin := ""

			if origin != "" {
				switch {
				case allowAll:
					allowedOrigin = "*"
				default:
					if _, ok := originSet[origin]; ok {
						allowedOrigin = origin
					}
				}
			}

			if allowedOrigin != "" {
				header := w.Header()
				header.Add("Vary", "Origin")
				header.Add("Vary", "Access-Control-Request-Method")
				header.Add("Vary", "Access-Control-Request-Headers")
				header.Set("Access-Control-Allow-Origin", allowedOrigin)
				header.Set("Access-Control-Allow-Methods", defaultCORSAllowedMethods)

				requestHeaders := strings.TrimSpace(r.Header.Get("Access-Control-Request-Headers"))
				if requestHeaders == "" {
					requestHeaders = defaultCORSAllowedHeaders
				}
				header.Set("Access-Control-Allow-Headers", requestHeaders)
				header.Set("Access-Control-Max-Age", "600")

				if allowedOrigin != "*" {
					header.Set("Access-Control-Allow-Credentials", "true")
				}
			}

			if r.Method == http.MethodOptions {
				if allowedOrigin == "" {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
