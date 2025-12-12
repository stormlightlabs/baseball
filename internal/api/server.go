// Package api provides HTTP handlers for the Baseball API
//
// @title Baseball API
// @version 1.0
// @description A comprehensive REST API for baseball statistics serving data from the Lahman Baseball Database and Retrosheet
// @BasePath /v1
//
// @contact.name API Support
// @contact.url https://github.com/stormlightlabs/baseball
// @contact.email info@stormlightlabs.org
//
// @license.name MPL-2.0
// @license.url https://opensource.org/license/mpl-2-0
package api

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	docs "stormlightlabs.org/baseball/internal/docs"
)

type Server struct {
	mux *http.ServeMux
}

// NewServer wires all registrars into one mux.
func NewServer(registrars ...Registrar) *Server {
	docs.SwaggerInfo.BasePath = "/v1"

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
	// @Success 200 {object} HealthResponse
	// @Router /health [get]
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
	})

	mux.HandleFunc("/docs/", httpSwagger.WrapHandler)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})

	return &Server{mux: mux}
}

// Implement http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
