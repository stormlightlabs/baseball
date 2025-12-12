package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

const mlbStatsAPIBase string = "https://statsapi.mlb.com/api"

var mlbProxyCatalog = []core.MLBProxyCatalogItem{
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

// MLBRoutes proxies select statsapi.mlb.com endpoints through /v1/mlb
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
// @Description Lists available MLB Stats API proxy routes surfaced under /v1/mlb. All endpoints default to sportId=1 (Major League Baseball) unless specified.
// @Tags mlb
// @Produce json
// @Success 200 {object} core.MLBOverviewResponse
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
// @Description Proxy to MLB Stats API /v1/people for live roster metadata. Defaults to sportId=1 (Major League Baseball) if not provided.
// @Tags mlb
// @Accept json
// @Produce json
// @Param personIds query string false "Comma-separated MLBAM personIds"
// @Param sportId query string false "Filter by sportId (defaults to 1 for MLB)"
// @Param hydrate query string false "Hydrate relationships"
// @Success 200 {object} core.MLBPeopleResponse
// @Failure 500 {object} ErrorResponse
// @Router /mlb/people [get]
func (mr *MLBRoutes) handleMLBPeople(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "people")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBPeopleResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBPerson godoc
// @Summary MLB person by ID
// @Description Proxy to MLB Stats API /v1/people/{personId}
// @Tags mlb
// @Accept json
// @Produce json
// @Param id path string true "MLBAM personId"
// @Param hydrate query string false "Hydrate relationships"
// @Success 200 {object} core.MLBPeopleResponse
// @Failure 500 {object} ErrorResponse
// @Router /mlb/people/{id} [get]
func (mr *MLBRoutes) handleMLBPerson(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "people", r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBPeopleResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBTeams godoc
// @Summary MLB teams
// @Description Proxy to MLB Stats API /v1/teams. Defaults to sportId=1 (Major League Baseball) if not provided.
// @Tags mlb
// @Accept json
// @Produce json
// @Param sportId query string false "Sport filter (defaults to 1 for MLB)"
// @Param season query string false "Season year"
// @Success 200 {object} core.MLBTeamsResponse
// @Failure 500 {object} ErrorResponse
// @Router /mlb/teams [get]
func (mr *MLBRoutes) handleMLBTeams(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "teams")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBTeamsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBTeam godoc
// @Summary MLB team by ID
// @Description Proxy to MLB Stats API /v1/teams/{teamId}
// @Tags mlb
// @Accept json
// @Produce json
// @Param id path string true "MLB teamId"
// @Param season query string false "Season year"
// @Success 200 {object} core.MLBTeamsResponse
// @Failure 500 {object} ErrorResponse
// @Router /mlb/teams/{id} [get]
func (mr *MLBRoutes) handleMLBTeam(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "teams", r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBTeamsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBSchedule godoc
// @Summary MLB schedule
// @Description Proxy to MLB Stats API /v1/schedule. Defaults to sportId=1 (Major League Baseball) if not provided.
// @Tags mlb
// @Accept json
// @Produce json
// @Param sportId query string false "Sport filter (defaults to 1 for MLB)"
// @Param teamId query string false "Team filter"
// @Param season query string false "Season year"
// @Param date query string false "Specific date (YYYY-MM-DD)"
// @Success 200 {object} core.MLBScheduleResponse
// @Failure 500 {object} ErrorResponse
// @Router /mlb/schedule [get]
func (mr *MLBRoutes) handleMLBSchedule(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "schedule")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBScheduleResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBSeasons godoc
// @Summary MLB seasons
// @Description Proxy to MLB Stats API /v1/seasons. Defaults to sportId=1 (Major League Baseball) if not provided.
// @Tags mlb
// @Accept json
// @Produce json
// @Param sportId query string false "Sport filter (defaults to 1 for MLB)"
// @Param season query string false "Season year"
// @Success 200 {object} core.MLBSeasonsResponse
// @Failure 500 {object} ErrorResponse
// @Router /mlb/seasons [get]
func (mr *MLBRoutes) handleMLBSeasons(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "seasons")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBSeasonsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBStats godoc
// @Summary MLB stats queries
// @Description Proxy to MLB Stats API /v1/stats for ad-hoc stats lookups. Note: sportId defaults to 1 (Major League Baseball).
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
	target, err := url.JoinPath(mr.baseURL, "v1", "stats")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result map[string]any // Stats endpoint has variable structure
	if err := json.Unmarshal(body, &result); err != nil {
		writeError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBStandings godoc
// @Summary MLB standings
// @Description Proxy to MLB Stats API /v1/standings. Note: sportId defaults to 1 (Major League Baseball).
// @Tags mlb
// @Accept json
// @Produce json
// @Param leagueId query string false "League filter"
// @Param season query string false "Season year"
// @Param standingsTypes query string false "Standings type (byLeague, etc.)"
// @Success 200 {object} core.MLBStandingsResponse
// @Failure 500 {object} ErrorResponse
// @Router /mlb/standings [get]
func (mr *MLBRoutes) handleMLBStandings(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "standings")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBStandingsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBAwards godoc
// @Summary MLB awards catalog
// @Description Proxy to MLB Stats API /v1/awards. Defaults to sportId=1 (Major League Baseball) if not provided.
// @Tags mlb
// @Accept json
// @Produce json
// @Param sportId query string false "Sport filter (defaults to 1 for MLB)"
// @Param season query string false "Season year"
// @Success 200 {object} core.MLBAwardsResponse
// @Failure 500 {object} ErrorResponse
// @Router /mlb/awards [get]
func (mr *MLBRoutes) handleMLBAwards(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "awards")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBAwardsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
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
// @Success 200 {object} core.MLBAwardsResponse
// @Failure 500 {object} ErrorResponse
// @Router /mlb/awards/{id} [get]
func (mr *MLBRoutes) handleMLBAward(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "awards", r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBAwardsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBVenues godoc
// @Summary MLB venues directory
// @Description Proxy to MLB Stats API /v1/venues. Defaults to sportId=1 (Major League Baseball) if not provided.
// @Tags mlb
// @Accept json
// @Produce json
// @Param venueIds query string false "Comma-separated venue IDs"
// @Param season query string false "Season year"
// @Param sportId query string false "Sport filter (defaults to 1 for MLB)"
// @Success 200 {object} core.MLBVenuesResponse
// @Failure 500 {object} ErrorResponse
// @Router /mlb/venues [get]
func (mr *MLBRoutes) handleMLBVenues(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "venues")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBVenuesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// fetchFromMLB handles the common HTTP request logic for all MLB proxy endpoints.
// It adds sportId=1 as default if not provided, sets User-Agent, and returns the response body and status code.
func (mr *MLBRoutes) fetchFromMLB(r *http.Request, target string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		return nil, 0, err
	}

	query := r.URL.Query()
	if query.Get("sportId") == "" {
		query.Set("sportId", "1")
	}
	req.URL.RawQuery = query.Encode()
	req.Header.Set("User-Agent", "Stormlight-Baseball-MLBProxy/1.0")

	resp, err := mr.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}
