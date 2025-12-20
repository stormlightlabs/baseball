package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type AchievementRoutes struct {
	repo core.AchievementRepository
}

func NewAchievementRoutes(repo core.AchievementRepository) *AchievementRoutes {
	return &AchievementRoutes{repo: repo}
}

func (ar *AchievementRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/achievements/no-hitters", ar.handleListNoHitters)
	mux.HandleFunc("GET /v1/achievements/cycles", ar.handleListCycles)
	mux.HandleFunc("GET /v1/achievements/multi-hr-games", ar.handleListMultiHRGames)
	mux.HandleFunc("GET /v1/achievements/triple-plays", ar.handleListTriplePlays)
	mux.HandleFunc("GET /v1/achievements/extra-inning-games", ar.handleListExtraInningGames)
}

// handleListNoHitters godoc
// @Summary List no-hitters
// @Description Get all games where a team allowed zero hits to the opposing team
// @Tags achievements
// @Accept json
// @Produce json
// @Param season query integer false "Filter by season year"
// @Param season_from query integer false "Filter from season (inclusive)"
// @Param season_to query integer false "Filter to season (inclusive)"
// @Param team_id query string false "Filter by team ID"
// @Param park_id query string false "Filter by park ID"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /achievements/no-hitters [get]
func (ar *AchievementRoutes) handleListNoHitters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := buildAchievementFilter(r)

	noHitters, err := ar.repo.ListNoHitters(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := ar.repo.CountNoHitters(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    noHitters,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleListCycles godoc
// @Summary List hitting for the cycle achievements
// @Description Get all games where a player hit a single, double, triple, and home run
// @Tags achievements
// @Accept json
// @Produce json
// @Param season query integer false "Filter by season year"
// @Param season_from query integer false "Filter from season (inclusive)"
// @Param season_to query integer false "Filter to season (inclusive)"
// @Param team_id query string false "Filter by team ID"
// @Param player_id query string false "Filter by player ID (Retrosheet format)"
// @Param park_id query string false "Filter by park ID"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /achievements/cycles [get]
func (ar *AchievementRoutes) handleListCycles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := buildAchievementFilter(r)

	cycles, err := ar.repo.ListCycles(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := ar.repo.CountCycles(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    cycles,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleListMultiHRGames godoc
// @Summary List multiple home run games
// @Description Get all games where a player hit 3 or more home runs
// @Tags achievements
// @Accept json
// @Produce json
// @Param season query integer false "Filter by season year"
// @Param season_from query integer false "Filter from season (inclusive)"
// @Param season_to query integer false "Filter to season (inclusive)"
// @Param team_id query string false "Filter by team ID"
// @Param player_id query string false "Filter by player ID (Retrosheet format)"
// @Param park_id query string false "Filter by park ID"
// @Param min_hr query integer false "Minimum home runs (default: 3)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /achievements/multi-hr-games [get]
func (ar *AchievementRoutes) handleListMultiHRGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := buildAchievementFilter(r)

	// Add min_hr filter
	if minHR := r.URL.Query().Get("min_hr"); minHR != "" {
		hr := getIntQuery(r, "min_hr", 3)
		filter.MinHR = &hr
	}

	multiHRGames, err := ar.repo.ListMultiHRGames(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := ar.repo.CountMultiHRGames(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    multiHRGames,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleListTriplePlays godoc
// @Summary List triple plays
// @Description Get all games where a team recorded one or more triple plays
// @Tags achievements
// @Accept json
// @Produce json
// @Param season query integer false "Filter by season year"
// @Param season_from query integer false "Filter from season (inclusive)"
// @Param season_to query integer false "Filter to season (inclusive)"
// @Param team_id query string false "Filter by team ID"
// @Param park_id query string false "Filter by park ID"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /achievements/triple-plays [get]
func (ar *AchievementRoutes) handleListTriplePlays(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := buildAchievementFilter(r)

	triplePlays, err := ar.repo.ListTriplePlays(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := ar.repo.CountTriplePlays(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    triplePlays,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleListExtraInningGames godoc
// @Summary List extra inning games
// @Description Get all games that lasted 20 or more innings
// @Tags achievements
// @Accept json
// @Produce json
// @Param season query integer false "Filter by season year"
// @Param season_from query integer false "Filter from season (inclusive)"
// @Param season_to query integer false "Filter to season (inclusive)"
// @Param team_id query string false "Filter by team ID (home or away)"
// @Param park_id query string false "Filter by park ID"
// @Param min_innings query integer false "Minimum innings (default: 20)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /achievements/extra-inning-games [get]
func (ar *AchievementRoutes) handleListExtraInningGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := buildAchievementFilter(r)

	// Add min_innings filter
	if minInnings := r.URL.Query().Get("min_innings"); minInnings != "" {
		innings := getIntQuery(r, "min_innings", 20)
		filter.MinInnings = &innings
	}

	extraInningGames, err := ar.repo.ListExtraInningGames(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := ar.repo.CountExtraInningGames(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    extraInningGames,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// buildAchievementFilter builds an AchievementFilter from HTTP query parameters
func buildAchievementFilter(r *http.Request) core.AchievementFilter {
	filter := core.AchievementFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if season := r.URL.Query().Get("season"); season != "" {
		y := core.SeasonYear(getIntQuery(r, "season", 0))
		filter.Season = &y
	}

	if seasonFrom := r.URL.Query().Get("season_from"); seasonFrom != "" {
		y := core.SeasonYear(getIntQuery(r, "season_from", 0))
		filter.SeasonFrom = &y
	}

	if seasonTo := r.URL.Query().Get("season_to"); seasonTo != "" {
		y := core.SeasonYear(getIntQuery(r, "season_to", 0))
		filter.SeasonTo = &y
	}

	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		t := core.TeamID(teamID)
		filter.TeamID = &t
	}

	if playerID := r.URL.Query().Get("player_id"); playerID != "" {
		filter.PlayerID = &playerID
	}

	if parkID := r.URL.Query().Get("park_id"); parkID != "" {
		p := core.ParkID(parkID)
		filter.ParkID = &p
	}
	return filter
}
