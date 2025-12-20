package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type AwardRoutes struct {
	repo core.AwardRepository
}

func NewAwardRoutes(repo core.AwardRepository) *AwardRoutes {
	return &AwardRoutes{repo: repo}
}

func (ar *AwardRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/awards", ar.handleListAwards)
	mux.HandleFunc("GET /v1/awards/{award_id}", ar.handleGetAward)
	mux.HandleFunc("GET /v1/seasons/{year}/awards", ar.handleSeasonAwards)
}

// handleListAwards godoc
// @Summary List all awards
// @Description Get a list of all baseball awards
// @Tags awards
// @Accept json
// @Produce json
// @Success 200 {object} AwardsListResponse
// @Failure 500 {object} ErrorResponse
// @Router /awards [get]
func (ar *AwardRoutes) handleListAwards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	awards, err := ar.repo.ListAwards(ctx)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, AwardsListResponse{
		Awards: awards,
	})
}

// handleGetAward godoc
// @Summary Get award details
// @Description Get detailed information about a specific award including winners
// @Tags awards
// @Accept json
// @Produce json
// @Param award_id path string true "Award ID"
// @Param year query integer false "Filter by year"
// @Param player_id query string false "Filter by player ID"
// @Param league query string false "Filter by league (AL, NL)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /awards/{award_id} [get]
func (ar *AwardRoutes) handleGetAward(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	awardID := core.AwardID(r.PathValue("award_id"))

	filter := core.AwardFilter{
		AwardID: &awardID,
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if year := r.URL.Query().Get("year"); year != "" {
		y := core.SeasonYear(getIntQuery(r, "year", 0))
		filter.Year = &y
	}

	if playerID := r.URL.Query().Get("player_id"); playerID != "" {
		pid := core.PlayerID(playerID)
		filter.PlayerID = &pid
	}

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	results, err := ar.repo.ListAwardResults(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := ar.repo.CountAwardResults(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    results,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleSeasonAwards godoc
// @Summary Get awards for a season
// @Description Get all awards issued during a specific season
// @Tags awards
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param award_id query string false "Filter by award ID"
// @Param player_id query string false "Filter by player ID"
// @Param league query string false "Filter by league (AL, NL)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/awards [get]
func (ar *AwardRoutes) handleSeasonAwards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	year := getIntPathValue(r, "year")

	filter := core.AwardFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	y := core.SeasonYear(year)
	filter.Year = &y

	if awardID := r.URL.Query().Get("award_id"); awardID != "" {
		aid := core.AwardID(awardID)
		filter.AwardID = &aid
	}

	if playerID := r.URL.Query().Get("player_id"); playerID != "" {
		pid := core.PlayerID(playerID)
		filter.PlayerID = &pid
	}

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	results, err := ar.repo.ListAwardResults(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := ar.repo.CountAwardResults(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    results,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}
