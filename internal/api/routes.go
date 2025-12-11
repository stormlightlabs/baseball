package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type PlayerRoutes struct {
	repo      core.PlayerRepository
	awardRepo core.AwardRepository
}

func NewPlayerRoutes(repo core.PlayerRepository, awardRepo core.AwardRepository) *PlayerRoutes {
	return &PlayerRoutes{repo: repo, awardRepo: awardRepo}
}

func (pr *PlayerRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/players", pr.handleListPlayers)
	mux.HandleFunc("GET /v1/players/{id}", pr.handleGetPlayer)
	mux.HandleFunc("GET /v1/players/{id}/seasons", pr.handlePlayerSeasons)
	mux.HandleFunc("GET /v1/players/{id}/awards", pr.handlePlayerAwards)
	mux.HandleFunc("GET /v1/players/{id}/hall-of-fame", pr.handlePlayerHallOfFame)
	mux.HandleFunc("GET /v1/players/{id}/game-logs", pr.handlePlayerGameLogs)
}

// handleGetPlayer godoc
// @Summary Get player by ID
// @Description Get detailed biographical information for a specific player
// @Tags players
// @Accept json
// @Produce json
// @Param id path string true "Player ID"
// @Success 200 {object} core.Player
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{id} [get]
func (pr *PlayerRoutes) handleGetPlayer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.PlayerID(r.PathValue("id"))

	player, err := pr.repo.GetByID(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, player)
}

// handleListPlayers godoc
// @Summary List players
// @Description Search and browse players with optional name filter and pagination
// @Tags players
// @Accept json
// @Produce json
// @Param name query string false "Player name search query"
// @Param debut_year query integer false "Filter by debut year"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /players [get]
func (pr *PlayerRoutes) handleListPlayers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.PlayerFilter{
		NameQuery: r.URL.Query().Get("name"),
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if debutYear := r.URL.Query().Get("debut_year"); debutYear != "" {
		year := core.SeasonYear(getIntQuery(r, "debut_year", 0))
		filter.DebutYear = &year
	}

	players, err := pr.repo.List(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := pr.repo.Count(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    players,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handlePlayerSeasons godoc
// @Summary Get player season statistics
// @Description Get season-by-season batting and pitching statistics for a player
// @Tags players
// @Accept json
// @Produce json
// @Param id path string true "Player ID"
// @Success 200 {object} PlayerSeasonsResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{id}/seasons [get]
func (pr *PlayerRoutes) handlePlayerSeasons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.PlayerID(r.PathValue("id"))

	batting, err := pr.repo.BattingSeasons(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}

	pitching, err := pr.repo.PitchingSeasons(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PlayerSeasonsResponse{
		Batting:  batting,
		Pitching: pitching,
	})
}

// handlePlayerAwards godoc
// @Summary Get player awards
// @Description Get all awards won by a player
// @Tags players
// @Accept json
// @Produce json
// @Param id path string true "Player ID"
// @Param year query integer false "Filter by year"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{id}/awards [get]
func (pr *PlayerRoutes) handlePlayerAwards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.PlayerID(r.PathValue("id"))

	filter := core.AwardFilter{
		PlayerID: &id,
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if year := r.URL.Query().Get("year"); year != "" {
		y := core.SeasonYear(getIntQuery(r, "year", 0))
		filter.Year = &y
	}

	awards, err := pr.awardRepo.ListAwardResults(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := pr.awardRepo.CountAwardResults(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    awards,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handlePlayerHallOfFame godoc
// @Summary Get player Hall of Fame records
// @Description Get Hall of Fame voting records for a player
// @Tags players
// @Accept json
// @Produce json
// @Param id path string true "Player ID"
// @Success 200 {object} HallOfFameResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{id}/hall-of-fame [get]
func (pr *PlayerRoutes) handlePlayerHallOfFame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.PlayerID(r.PathValue("id"))

	records, err := pr.awardRepo.HallOfFameByPlayer(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, HallOfFameResponse{
		Records: records,
	})
}

func (pr *PlayerRoutes) handlePlayerGameLogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, ErrorResponse{
		Error: "game logs endpoint not yet implemented",
	})
}

type GameRoutes struct {
	repo core.GameRepository
}

func NewGameRoutes(repo core.GameRepository) *GameRoutes { return &GameRoutes{repo: repo} }

// TODO: retrosheet routing
func (gr *GameRoutes) RegisterRoutes(mux *http.ServeMux) {
	// mux.HandleFunc("GET /v1/games", gr.handleListGames)
	// mux.HandleFunc("GET /v1/games/{id}", gr.handleGetGame)
	// mux.HandleFunc("GET /v1/games/{id}/events", gr.handleGameEvents)
}
