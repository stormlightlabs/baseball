package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type EjectionRoutes struct {
	repo core.EjectionRepository
}

func NewEjectionRoutes(repo core.EjectionRepository) *EjectionRoutes {
	return &EjectionRoutes{repo: repo}
}

func (er *EjectionRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/ejections", er.handleListEjections)
	mux.HandleFunc("GET /v1/seasons/{year}/ejections", er.handleSeasonEjections)
}

// handleListEjections godoc
// @Summary List ejections
// @Description Get ejections with optional filters for player, umpire, team, and role
// @Tags ejections
// @Accept json
// @Produce json
// @Param player_id query string false "Filter by ejected player ID"
// @Param umpire_id query string false "Filter by umpire ID"
// @Param team query string false "Filter by team"
// @Param role query string false "Filter by role" Enums(P, M, C) "P=Player, M=Manager, C=Coach"
// @Param season query integer false "Filter by season year"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /ejections [get]
func (er *EjectionRoutes) handleListEjections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.EjectionFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if playerID := r.URL.Query().Get("player_id"); playerID != "" {
		pid := core.RetroPlayerID(playerID)
		filter.PlayerID = &pid
	}

	if umpireID := r.URL.Query().Get("umpire_id"); umpireID != "" {
		uid := core.UmpireID(umpireID)
		filter.UmpireID = &uid
	}

	if team := r.URL.Query().Get("team"); team != "" {
		tid := core.TeamID(team)
		filter.TeamID = &tid
	}

	if role := r.URL.Query().Get("role"); role != "" {
		filter.Role = &role
	}

	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		season := core.SeasonYear(getIntQuery(r, "season", 0))
		filter.Season = &season
	}

	ejections, err := er.repo.List(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := er.repo.Count(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    ejections,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleSeasonEjections godoc
// @Summary Get ejections for a specific season
// @Description Get all ejections that occurred during a specific season
// @Tags ejections, seasons
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/ejections [get]
func (er *EjectionRoutes) handleSeasonEjections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	year := getIntPathValue(r, "year")
	if year == 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid year"})
		return
	}

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 50),
	}

	ejections, err := er.repo.ListBySeason(ctx, core.SeasonYear(year), pagination)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := er.repo.CountBySeason(ctx, core.SeasonYear(year))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    ejections,
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		Total:   total,
	})
}
