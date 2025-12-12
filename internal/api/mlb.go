package api

import (
	"io"
	"log"
	"net/http"
	"net/url"
	pathpkg "path"
	"strings"
	"time"
)

const mlbStatsAPIBase = "https://statsapi.mlb.com/api"

var mlbProxyCatalog = []struct {
	Route       string `json:"route"`
	Target      string `json:"target"`
	Description string `json:"description"`
}{
	{Route: "/v1/mlb/people", Target: "/v1/people", Description: "Player search and metadata from MLBAM"},
	{Route: "/v1/mlb/people/{id}", Target: "/v1/people/{personId}", Description: "Single player lookup"},
	{Route: "/v1/mlb/teams", Target: "/v1/teams", Description: "MLB team reference and roster metadata"},
	{Route: "/v1/mlb/teams/{id}", Target: "/v1/teams/{teamId}", Description: "Single team details"},
	{Route: "/v1/mlb/schedule", Target: "/v1/schedule", Description: "Daily/season schedule with game metadata"},
	{Route: "/v1/mlb/seasons", Target: "/v1/seasons", Description: "Season directory with league/division data"},
	{Route: "/v1/mlb/stats", Target: "/v1/stats", Description: "MLB-wide stats queries"},
	{Route: "/v1/mlb/standings", Target: "/v1/standings", Description: "League/division standings"},
	{Route: "/v1/mlb/awards", Target: "/v1/awards", Description: "Awards directory and recipients"},
	{Route: "/v1/mlb/awards/{id}", Target: "/v1/awards/{awardId}", Description: "Single MLB awards endpoint"},
	{Route: "/v1/mlb/venues", Target: "/v1/venues", Description: "Ballpark directory"},
}

// MLBRoutes proxies select statsapi.mlb.com endpoints through /v1/mlb so they
// show up in our documentation and share the same auth/caching stack.
type MLBRoutes struct {
	client  *http.Client
	baseURL string
}

func NewMLBRoutes(client *http.Client) *MLBRoutes {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &MLBRoutes{
		client:  client,
		baseURL: mlbStatsAPIBase,
	}
}

func (mr *MLBRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/mlb", mr.handleMLBOverview)
	mux.HandleFunc("GET /v1/mlb/people", mr.handleMLBPeople)
	mux.HandleFunc("GET /v1/mlb/people/{id}", mr.handleMLBPerson)
	mux.HandleFunc("GET /v1/mlb/teams", mr.handleMLBTeams)
	mux.HandleFunc("GET /v1/mlb/teams/{id}", mr.handleMLBTeam)
	mux.HandleFunc("GET /v1/mlb/schedule", mr.handleMLBSchedule)
	mux.HandleFunc("GET /v1/mlb/seasons", mr.handleMLBSeasons)
	mux.HandleFunc("GET /v1/mlb/stats", mr.handleMLBStats)
	mux.HandleFunc("GET /v1/mlb/standings", mr.handleMLBStandings)
	mux.HandleFunc("GET /v1/mlb/awards", mr.handleMLBAwards)
	mux.HandleFunc("GET /v1/mlb/awards/{id}", mr.handleMLBAward)
	mux.HandleFunc("GET /v1/mlb/venues", mr.handleMLBVenues)
}

// handleMLBOverview godoc
// @Summary MLB Stats proxy catalog
// @Description Lists available MLB Stats API proxy routes surfaced under /v1/mlb
// @Tags mlb
// @Produce json
// @Success 200 {array} map[string]any
// @Router /mlb [get]
func (mr *MLBRoutes) handleMLBOverview(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"base_url": "/v1/mlb",
		"target":   mlbStatsAPIBase,
		"routes":   mlbProxyCatalog,
	})
}

// handleMLBPeople godoc
// @Summary MLB people search
// @Description Proxy to MLB Stats API /v1/people for live roster metadata
// @Tags mlb
// @Accept json
// @Produce json
// @Param personIds query string false "Comma-separated MLBAM personIds"
// @Param sportId query string false "Filter by sportId"
// @Param hydrate query string false "Hydrate relationships"
// @Success 200 {object} map[string]any
// @Failure 500 {object} ErrorResponse
// @Router /mlb/people [get]
func (mr *MLBRoutes) handleMLBPeople(w http.ResponseWriter, r *http.Request) {
	mr.proxyGet(w, r, "v1", "people")
}

// handleMLBPerson godoc
// @Summary MLB person by ID
// @Description Proxy to MLB Stats API /v1/people/{personId}
// @Tags mlb
// @Accept json
// @Produce json
// @Param id path string true "MLBAM personId"
// @Param hydrate query string false "Hydrate relationships"
// @Success 200 {object} map[string]any
// @Failure 500 {object} ErrorResponse
// @Router /mlb/people/{id} [get]
func (mr *MLBRoutes) handleMLBPerson(w http.ResponseWriter, r *http.Request) {
	mr.proxyGet(w, r, "v1", "people", r.PathValue("id"))
}

// handleMLBTeams godoc
// @Summary MLB teams
// @Description Proxy to MLB Stats API /v1/teams
// @Tags mlb
// @Accept json
// @Produce json
// @Param sportId query string false "Sport filter"
// @Param season query string false "Season year"
// @Success 200 {object} map[string]any
// @Failure 500 {object} ErrorResponse
// @Router /mlb/teams [get]
func (mr *MLBRoutes) handleMLBTeams(w http.ResponseWriter, r *http.Request) {
	mr.proxyGet(w, r, "v1", "teams")
}

// handleMLBTeam godoc
// @Summary MLB team by ID
// @Description Proxy to MLB Stats API /v1/teams/{teamId}
// @Tags mlb
// @Accept json
// @Produce json
// @Param id path string true "MLB teamId"
// @Param season query string false "Season year"
// @Success 200 {object} map[string]any
// @Failure 500 {object} ErrorResponse
// @Router /mlb/teams/{id} [get]
func (mr *MLBRoutes) handleMLBTeam(w http.ResponseWriter, r *http.Request) {
	mr.proxyGet(w, r, "v1", "teams", r.PathValue("id"))
}

// handleMLBSchedule godoc
// @Summary MLB schedule
// @Description Proxy to MLB Stats API /v1/schedule
// @Tags mlb
// @Accept json
// @Produce json
// @Param sportId query string false "Sport filter"
// @Param teamId query string false "Team filter"
// @Param season query string false "Season year"
// @Param date query string false "Specific date (YYYY-MM-DD)"
// @Success 200 {object} map[string]any
// @Failure 500 {object} ErrorResponse
// @Router /mlb/schedule [get]
func (mr *MLBRoutes) handleMLBSchedule(w http.ResponseWriter, r *http.Request) {
	mr.proxyGet(w, r, "v1", "schedule")
}

// handleMLBSeasons godoc
// @Summary MLB seasons
// @Description Proxy to MLB Stats API /v1/seasons
// @Tags mlb
// @Accept json
// @Produce json
// @Param sportId query string false "Sport filter"
// @Param season query string false "Season year"
// @Success 200 {object} map[string]any
// @Failure 500 {object} ErrorResponse
// @Router /mlb/seasons [get]
func (mr *MLBRoutes) handleMLBSeasons(w http.ResponseWriter, r *http.Request) {
	mr.proxyGet(w, r, "v1", "seasons")
}

// handleMLBStats godoc
// @Summary MLB stats queries
// @Description Proxy to MLB Stats API /v1/stats for ad-hoc stats lookups
// @Tags mlb
// @Accept json
// @Produce json
// @Param stats query string true "Stat group(s) to query"
// @Param group query string true "Grouping (e.g., hitting, pitching)"
// @Param season query string false "Season year"
// @Param gameType query string false "Game type (R, S, etc.)"
// @Success 200 {object} map[string]any
// @Failure 500 {object} ErrorResponse
// @Router /mlb/stats [get]
func (mr *MLBRoutes) handleMLBStats(w http.ResponseWriter, r *http.Request) {
	mr.proxyGet(w, r, "v1", "stats")
}

// handleMLBStandings godoc
// @Summary MLB standings
// @Description Proxy to MLB Stats API /v1/standings
// @Tags mlb
// @Accept json
// @Produce json
// @Param leagueId query string false "League filter"
// @Param season query string false "Season year"
// @Param standingsTypes query string false "Standings type (byLeague, etc.)"
// @Success 200 {object} map[string]any
// @Failure 500 {object} ErrorResponse
// @Router /mlb/standings [get]
func (mr *MLBRoutes) handleMLBStandings(w http.ResponseWriter, r *http.Request) {
	mr.proxyGet(w, r, "v1", "standings")
}

// handleMLBAwards godoc
// @Summary MLB awards catalog
// @Description Proxy to MLB Stats API /v1/awards
// @Tags mlb
// @Accept json
// @Produce json
// @Param sportId query string false "Sport filter"
// @Param season query string false "Season year"
// @Success 200 {object} map[string]any
// @Failure 500 {object} ErrorResponse
// @Router /mlb/awards [get]
func (mr *MLBRoutes) handleMLBAwards(w http.ResponseWriter, r *http.Request) {
	mr.proxyGet(w, r, "v1", "awards")
}

// handleMLBAward godoc
// @Summary MLB award by ID
// @Description Proxy to MLB Stats API /v1/awards/{awardId}
// @Tags mlb
// @Accept json
// @Produce json
// @Param id path string true "MLB awardId"
// @Param season query string false "Season year"
// @Param hydrate query string false "Hydrate relationships"
// @Success 200 {object} map[string]any
// @Failure 500 {object} ErrorResponse
// @Router /mlb/awards/{id} [get]
func (mr *MLBRoutes) handleMLBAward(w http.ResponseWriter, r *http.Request) {
	mr.proxyGet(w, r, "v1", "awards", r.PathValue("id"))
}

// handleMLBVenues godoc
// @Summary MLB venues directory
// @Description Proxy to MLB Stats API /v1/venues
// @Tags mlb
// @Accept json
// @Produce json
// @Param venueIds query string false "Comma-separated venue IDs"
// @Param season query string false "Season year"
// @Success 200 {object} map[string]any
// @Failure 500 {object} ErrorResponse
// @Router /mlb/venues [get]
func (mr *MLBRoutes) handleMLBVenues(w http.ResponseWriter, r *http.Request) {
	mr.proxyGet(w, r, "v1", "venues")
}

// TODO: response types
func (mr *MLBRoutes) proxyGet(w http.ResponseWriter, r *http.Request, segments ...string) {
	trimmed := make([]string, 0, len(segments))
	for _, seg := range segments {
		seg = strings.Trim(seg, "/")
		if seg != "" {
			trimmed = append(trimmed, seg)
		}
	}
	if len(trimmed) == 0 {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "invalid MLB proxy path"})
		return
	}

	targetPath := pathpkg.Join(trimmed...)
	target, err := url.JoinPath(mr.baseURL, targetPath)
	if err != nil {
		writeError(w, err)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		writeError(w, err)
		return
	}
	req.URL.RawQuery = r.URL.RawQuery

	if accept := r.Header.Get("Accept"); accept != "" {
		req.Header.Set("Accept", accept)
	}
	if ua := r.Header.Get("User-Agent"); ua != "" {
		req.Header.Set("User-Agent", ua)
	} else {
		req.Header.Set("User-Agent", "Stormlight-Baseball-MLBProxy/1.0")
	}

	resp, err := mr.client.Do(req)
	if err != nil {
		writeError(w, err)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		if strings.EqualFold(key, "Content-Length") {
			continue
		}
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}
	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("mlb proxy response copy error: %v", err)
	}
}
