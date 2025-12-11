package api

import (
	_ "embed"
	"html/template"
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

//go:embed templates/dashboard.html
var dashboardTemplate string

//go:embed templates/login.html
var loginTemplate string

type UIRoutes struct {
	apiKeyRepo core.APIKeyRepository
}

func NewUIRoutes(apiKeyRepo core.APIKeyRepository) *UIRoutes {
	return &UIRoutes{apiKeyRepo: apiKeyRepo}
}

func (r *UIRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /dashboard", r.handleDashboard)
	mux.HandleFunc("GET /login", r.handleLoginPage)
}

func (r *UIRoutes) handleDashboard(w http.ResponseWriter, req *http.Request) {
	tmpl := template.Must(template.New("dashboard").Parse(dashboardTemplate))
	user, _ := req.Context().Value("user").(*core.User)
	data := struct{ User *core.User }{User: user}

	if err := tmpl.Execute(w, data); err != nil {
		writeError(w, err)
	}
}

func (r *UIRoutes) handleLoginPage(w http.ResponseWriter, req *http.Request) {
	tmpl := template.Must(template.New("login").Parse(loginTemplate))

	if err := tmpl.Execute(w, nil); err != nil {
		writeError(w, err)
	}
}
