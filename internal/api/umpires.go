package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type UmpireRoutes struct {
	repo core.UmpireRepository
}

func NewUmpireRoutes(repo core.UmpireRepository) *UmpireRoutes {
	return &UmpireRoutes{repo: repo}
}

func (ur *UmpireRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/umpires", ur.handleListUmpires)
	mux.HandleFunc("GET /v1/umpires/{umpire_id}", ur.handleGetUmpire)
	mux.HandleFunc("GET /v1/umpires/{umpire_id}/games", ur.handleUmpireGames)
}

// handleListUmpires godoc
// @Summary List umpires
// @Description Get a paginated list of all umpires
// @Tags umpires
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /umpires [get]
func (ur *UmpireRoutes) handleListUmpires(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 50),
	}

	umpires, err := ur.repo.List(ctx, pagination)
	if err != nil {
		writeError(w, err)
		return
	}

	total := len(umpires)

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    umpires,
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		Total:   total,
	})
}

// handleGetUmpire godoc
// @Summary Get umpire by ID
// @Description Get detailed information about a specific umpire
// @Tags umpires
// @Accept json
// @Produce json
// @Param umpire_id path string true "Umpire ID"
// @Success 200 {object} core.Umpire
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /umpires/{umpire_id} [get]
func (ur *UmpireRoutes) handleGetUmpire(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	umpireID := core.UmpireID(r.PathValue("umpire_id"))

	umpire, err := ur.repo.GetByID(ctx, umpireID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, umpire)
}

// handleUmpireGames godoc
// @Summary Get games officiated by an umpire
// @Description Get all games where the umpire officiated in any position
// @Tags umpires, games
// @Accept json
// @Produce json
// @Param umpire_id path string true "Umpire ID"
// @Param season query integer false "Filter by season year"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /umpires/{umpire_id}/games [get]
func (ur *UmpireRoutes) handleUmpireGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	umpireID := core.UmpireID(r.PathValue("umpire_id"))

	filter := core.GameFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		season := core.SeasonYear(getIntQuery(r, "season", 0))
		filter.Season = &season
	}

	games, err := ur.repo.GamesForUmpire(ctx, umpireID, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    games,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   len(games),
	})
}
