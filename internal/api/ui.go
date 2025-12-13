package api

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"stormlightlabs.org/baseball/internal/core"
)

const (
	templateDir string = "internal/api/templates"
	staticDir   string = "internal/api/static"
)

type UIRoutes struct {
	apiKeyRepo core.APIKeyRepository
	templates  *template.Template
	staticFS   http.Handler
}

func NewUIRoutes(apiKeyRepo core.APIKeyRepository) *UIRoutes {
	tmpl := template.Must(template.ParseGlob(filepath.Join(templateDir, "*.html")))
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
	return &UIRoutes{apiKeyRepo: apiKeyRepo, templates: tmpl, staticFS: staticHandler}
}

func (r *UIRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/static/", r.staticFS)
	mux.HandleFunc("GET /dashboard", r.handleDashboard)
	mux.HandleFunc("GET /login", r.handleLoginPage)
	mux.HandleFunc("GET /examples", r.handleExamples)
}

func (r *UIRoutes) handleDashboard(w http.ResponseWriter, req *http.Request) {
	user, _ := req.Context().Value("user").(*core.User)
	data := struct{ User *core.User }{User: user}
	r.renderTemplate(w, "dashboard.html", data)
}

func (r *UIRoutes) handleLoginPage(w http.ResponseWriter, req *http.Request) {
	r.renderTemplate(w, "login.html", nil)
}

func (r *UIRoutes) handleExamples(w http.ResponseWriter, req *http.Request) {
	user, _ := req.Context().Value("user").(*core.User)
	data := struct{ User *core.User }{User: user}
	r.renderTemplate(w, "examples.html", data)
}

func (r *UIRoutes) renderTemplate(w http.ResponseWriter, name string, data any) {
	tmpl := r.templates.Lookup(name)
	if tmpl == nil {
		writeInternalServerError(w, fmt.Errorf("template %s not found", name))
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		writeInternalServerError(w, err)
	}
}
