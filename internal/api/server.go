package api

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "stormlightlabs.org/baseball/internal/docs"
)

type Server struct {
	mux *http.ServeMux
}

// NewServer wires all registrars into one mux.
func NewServer(registrars ...Registrar) *Server {
	mux := http.NewServeMux()

	for _, r := range registrars {
		r.RegisterRoutes(mux)
	}

	// Health check endpoint
	// @Summary Health check
	// @Description Check if the API server is running
	// @Tags health
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]string
	// @Router /health [get]
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Swagger documentation at root and /docs
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/docs/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return &Server{mux: mux}
}

// Implement http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
