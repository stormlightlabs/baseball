package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type TeamRoutes struct {
	repo core.TeamRepository
}

func NewTeamRoutes(repo core.TeamRepository) *TeamRoutes {
	return &TeamRoutes{repo: repo}
}

func (tr *TeamRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/teams", tr.handleListTeams)
	mux.HandleFunc("GET /v1/teams/{id}", tr.handleGetTeam)
	mux.HandleFunc("GET /v1/seasons", tr.handleListSeasons)
	mux.HandleFunc("GET /v1/seasons/{year}/teams", tr.handleSeasonTeams)
	mux.HandleFunc("GET /v1/seasons/{year}/teams/{team_id}/roster", tr.handleTeamRoster)
	mux.HandleFunc("GET /v1/franchises", tr.handleListFranchises)
	mux.HandleFunc("GET /v1/franchises/{id}", tr.handleGetFranchise)
}

// handleListTeams godoc
// @Summary List team seasons
// @Description List team seasons with optional year and league filters
// @Tags teams
// @Accept json
// @Produce json
// @Param year query integer false "Filter by year"
// @Param league query string false "Filter by league (AL, NL)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /teams [get]
func (tr *TeamRoutes) handleListTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.TeamFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if year := r.URL.Query().Get("year"); year != "" {
		y := core.SeasonYear(getIntQuery(r, "year", 0))
		filter.Year = &y
	}

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	teams, err := tr.repo.ListTeamSeasons(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := tr.repo.CountTeamSeasons(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    teams,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleGetTeam godoc
// @Summary Get team season
// @Description Get a single team-season record
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param year query integer false "Year" default(2024)
// @Success 200 {object} core.TeamSeason
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /teams/{id} [get]
func (tr *TeamRoutes) handleGetTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.TeamID(r.PathValue("id"))

	year := core.SeasonYear(getIntQuery(r, "year", 2024))

	team, err := tr.repo.GetTeamSeason(ctx, id, year)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, team)
}

// handleSeasonTeams godoc
// @Summary Get all teams for a season
// @Description List all teams that played in a specific season
// @Tags teams
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param league query string false "Filter by league (AL, NL)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/teams [get]
func (tr *TeamRoutes) handleSeasonTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntQuery(r, "year", 2024))

	filter := core.TeamFilter{
		Year: &year,
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	teams, err := tr.repo.ListTeamSeasons(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := tr.repo.CountTeamSeasons(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    teams,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleListFranchises godoc
// @Summary List franchises
// @Description List all baseball franchises with optional active filter
// @Tags franchises
// @Accept json
// @Produce json
// @Param active query boolean false "Filter to only active franchises"
// @Success 200 {object} FranchisesResponse
// @Failure 500 {object} ErrorResponse
// @Router /franchises [get]
func (tr *TeamRoutes) handleListFranchises(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	onlyActive := r.URL.Query().Get("active") == "true"

	franchises, err := tr.repo.ListFranchises(ctx, onlyActive)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, FranchisesResponse{
		Franchises: franchises,
		Total:      len(franchises),
	})
}

// handleListSeasons godoc
// @Summary List all seasons
// @Description Get summary of all available seasons with league and team counts
// @Tags seasons
// @Accept json
// @Produce json
// @Success 200 {array} core.Season
// @Failure 500 {object} ErrorResponse
// @Router /seasons [get]
func (tr *TeamRoutes) handleListSeasons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	seasons, err := tr.repo.ListSeasons(ctx)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, seasons)
}

// handleTeamRoster godoc
// @Summary Get team roster
// @Description Get roster for a specific team and season with basic stats
// @Tags teams
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param team_id path string true "Team ID"
// @Success 200 {array} core.RosterPlayer
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/teams/{team_id}/roster [get]
func (tr *TeamRoutes) handleTeamRoster(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	teamID := core.TeamID(r.PathValue("team_id"))

	roster, err := tr.repo.Roster(ctx, year, teamID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, roster)
}

// handleGetFranchise godoc
// @Summary Get franchise
// @Description Get details for a specific franchise
// @Tags franchises
// @Accept json
// @Produce json
// @Param id path string true "Franchise ID"
// @Success 200 {object} core.Franchise
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /franchises/{id} [get]
func (tr *TeamRoutes) handleGetFranchise(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.FranchiseID(r.PathValue("id"))

	franchise, err := tr.repo.GetFranchise(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, franchise)
}
