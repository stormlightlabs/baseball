package api

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"stormlightlabs.org/baseball/internal/core"
)

const (
	templateDir         string = "internal/api/templates"
	staticDir           string = "internal/api/static"
	defaultContactEmail string = "info@stormlightlabs.org"
)

type UIRoutes struct {
	apiKeyRepo core.APIKeyRepository
	templates  map[string]*template.Template
	staticFS   http.Handler
}

type TemplateData struct {
	User         *core.User
	ActivePage   string
	ShowLogout   bool
	ContactEmail string
}

func NewUIRoutes(apiKeyRepo core.APIKeyRepository) *UIRoutes {
	pages := mustParseTemplates()
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
	return &UIRoutes{apiKeyRepo: apiKeyRepo, templates: pages, staticFS: staticHandler}
}

func (r *UIRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /static/", r.staticFS)
	mux.Handle("HEAD /static/", r.staticFS)
	mux.HandleFunc("GET /", r.handleHome)
	mux.HandleFunc("GET /dashboard", r.handleDashboard)
	mux.HandleFunc("GET /login", r.handleLoginPage)
	mux.HandleFunc("GET /examples", r.handleExamples)
	mux.HandleFunc("GET /examples/advanced", r.handleAdvancedExamples)
}

func (r *UIRoutes) handleHome(w http.ResponseWriter, req *http.Request) {
	user, _ := req.Context().Value("user").(*core.User)
	data := newTemplateData(user, "home")
	r.renderTemplate(w, "home.html", data)
}

func (r *UIRoutes) handleDashboard(w http.ResponseWriter, req *http.Request) {
	user, _ := req.Context().Value("user").(*core.User)
	data := newTemplateData(user, "dashboard")
	if user != nil {
		data.ShowLogout = true
	}
	r.renderTemplate(w, "dashboard.html", data)
}

func (r *UIRoutes) handleLoginPage(w http.ResponseWriter, req *http.Request) {
	data := newTemplateData(nil, "login")
	r.renderTemplate(w, "login.html", data)
}

func (r *UIRoutes) handleExamples(w http.ResponseWriter, req *http.Request) {
	user, _ := req.Context().Value("user").(*core.User)
	data := newTemplateData(user, "examples-basic")
	r.renderTemplate(w, "examples.html", data)
}

func (r *UIRoutes) handleAdvancedExamples(w http.ResponseWriter, req *http.Request) {
	user, _ := req.Context().Value("user").(*core.User)
	data := newTemplateData(user, "examples-advanced")
	r.renderTemplate(w, "advanced.html", data)
}

func (r *UIRoutes) renderTemplate(w http.ResponseWriter, name string, data any) {
	tmpl, ok := r.templates[name]
	if !ok || tmpl == nil {
		writeInternalServerError(w, fmt.Errorf("template %s not found", name))
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		writeInternalServerError(w, err)
	}
}

func newTemplateData(user *core.User, activePage string) TemplateData {
	contact := defaultContactEmail
	if user != nil && user.Email != "" {
		contact = user.Email
	}
	return TemplateData{
		User:         user,
		ActivePage:   activePage,
		ContactEmail: contact,
	}
}

func mustParseTemplates() map[string]*template.Template {
	basePath := filepath.Join(templateDir, "base.html")
	baseTmpl := template.Must(template.ParseFiles(basePath))

	files, err := filepath.Glob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		panic(fmt.Errorf("failed to list templates: %w", err))
	}

	parsed := make(map[string]*template.Template)
	for _, file := range files {
		if filepath.Base(file) == "base.html" {
			continue
		}

		clone, err := baseTmpl.Clone()
		if err != nil {
			panic(fmt.Errorf("failed to clone base template: %w", err))
		}

		if _, err := clone.ParseFiles(file); err != nil {
			panic(fmt.Errorf("failed to parse template %s: %w", file, err))
		}

		name := filepath.Base(file)
		parsed[name] = clone.Lookup(name)
	}

	return parsed
}
