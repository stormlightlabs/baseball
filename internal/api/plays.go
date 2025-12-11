package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

type PlayRoutes struct {
	repo       core.PlayRepository
	playerRepo core.PlayerRepository
}

func NewPlayRoutes(repo core.PlayRepository, playerRepo core.PlayerRepository) *PlayRoutes {
	return &PlayRoutes{repo: repo, playerRepo: playerRepo}
}

func (pr *PlayRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/plays", pr.handleListPlays)
	mux.HandleFunc("GET /v1/games/{id}/plays", pr.handleGamePlays)
	mux.HandleFunc("GET /v1/players/{id}/plays", pr.handlePlayerPlays)
	mux.HandleFunc("GET /v1/players/{id}/plate-appearances", pr.handlePlayerPlateAppearances)
}

// handleListPlays godoc
// @Summary List plays
// @Description Query plays with various filters
// @Tags plays
// @Accept json
// @Produce json
// @Param batter query string false "Batter Retrosheet ID"
// @Param pitcher query string false "Pitcher Retrosheet ID"
// @Param bat_team query string false "Batting team ID"
// @Param pit_team query string false "Pitching team ID"
// @Param date query string false "Game date (YYYYMMDD)"
// @Param date_from query string false "Start date (YYYYMMDD)"
// @Param date_to query string false "End date (YYYYMMDD)"
// @Param inning query integer false "Filter by inning"
// @Param home_runs query boolean false "Filter to only home runs"
// @Param walks query boolean false "Filter to only walks"
// @Param strikeouts query boolean false "Filter to only strikeouts"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /plays [get]
func (pr *PlayRoutes) handleListPlays(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.PlayFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if batter := r.URL.Query().Get("batter"); batter != "" {
		b := core.RetroPlayerID(batter)
		filter.Batter = &b
	}

	if pitcher := r.URL.Query().Get("pitcher"); pitcher != "" {
		p := core.RetroPlayerID(pitcher)
		filter.Pitcher = &p
	}

	if batTeam := r.URL.Query().Get("bat_team"); batTeam != "" {
		t := core.TeamID(batTeam)
		filter.BatTeam = &t
	}

	if pitTeam := r.URL.Query().Get("pit_team"); pitTeam != "" {
		t := core.TeamID(pitTeam)
		filter.PitTeam = &t
	}

	if date := r.URL.Query().Get("date"); date != "" {
		filter.Date = &date
	}

	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		filter.DateFrom = &dateFrom
	}

	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		filter.DateTo = &dateTo
	}

	if inning := r.URL.Query().Get("inning"); inning != "" {
		i := getIntQuery(r, "inning", 0)
		filter.Inning = &i
	}

	if hr := r.URL.Query().Get("home_runs"); hr == "true" {
		t := true
		filter.HomeRuns = &t
	}

	if walk := r.URL.Query().Get("walks"); walk == "true" {
		t := true
		filter.Walks = &t
	}

	if k := r.URL.Query().Get("strikeouts"); k == "true" {
		t := true
		filter.K = &t
	}

	plays, err := pr.repo.List(ctx, filter)
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
		Data:    plays,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleGamePlays godoc
// @Summary Get plays for a game
// @Description Get all plays for a specific game in chronological order
// @Tags plays, games
// @Accept json
// @Produce json
// @Param id path string true "Game ID (format: YYYYMMDD + game_number + home_team)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(200)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{id}/plays [get]
func (pr *PlayRoutes) handleGamePlays(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gameID := core.GameID(r.PathValue("id"))

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 200),
	}

	plays, err := pr.repo.ListByGame(ctx, gameID, pagination)
	if err != nil {
		writeError(w, err)
		return
	}

	filter := core.PlayFilter{
		GameID: &gameID,
		Pagination: core.Pagination{
			Page:    1,
			PerPage: 1,
		},
	}
	total, err := pr.repo.Count(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    plays,
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		Total:   total,
	})
}

// handlePlayerPlays godoc
// @Summary Get plays for a player
// @Description Get all plays involving a specific player (as batter or pitcher)
// @Tags plays, players
// @Accept json
// @Produce json
// @Param id path string true "Player ID"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{id}/plays [get]
func (pr *PlayRoutes) handlePlayerPlays(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playerID := core.PlayerID(r.PathValue("id"))

	retroID, err := pr.lookupRetroID(ctx, playerID)
	if err != nil {
		if errors.Is(err, errMissingRetroID) {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		} else {
			writeError(w, err)
		}
		return
	}

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 50),
	}

	plays, err := pr.repo.ListByPlayer(ctx, retroID, pagination)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := pr.repo.CountByPlayer(ctx, retroID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    plays,
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		Total:   total,
	})
}

// handlePlayerPlateAppearances godoc
// @Summary Get plate appearances for a player
// @Description Return plate appearances where the player is the batter with optional filters
// @Tags plays, players
// @Accept json
// @Produce json
// @Param id path string true "Player ID"
// @Param season query integer false "Filter by season year"
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Param game_id query string false "Retrosheet game ID"
// @Param pitcher query string false "Filter by pitcher Retrosheet ID"
// @Param vs_pitcher query string false "Filter by Lahman pitcher ID"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{id}/plate-appearances [get]
func (pr *PlayRoutes) handlePlayerPlateAppearances(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playerID := core.PlayerID(r.PathValue("id"))

	retroID, err := pr.lookupRetroID(ctx, playerID)
	if err != nil {
		if errors.Is(err, errMissingRetroID) {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		} else {
			writeError(w, err)
		}
		return
	}

	filter := core.PlayFilter{
		Batter: &retroID,
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
		SortBy:    "date",
		SortOrder: core.SortDesc,
	}

	if season := getIntQuery(r, "season", 0); season > 0 {
		start := fmt.Sprintf("%04d0101", season)
		end := fmt.Sprintf("%04d1231", season)
		filter.DateFrom = &start
		filter.DateTo = &end
	}

	if from := r.URL.Query().Get("date_from"); from != "" {
		if parsed, err := time.Parse("2006-01-02", from); err == nil {
			formatted := parsed.Format("20060102")
			filter.DateFrom = &formatted
		}
	}

	if to := r.URL.Query().Get("date_to"); to != "" {
		if parsed, err := time.Parse("2006-01-02", to); err == nil {
			formatted := parsed.Format("20060102")
			filter.DateTo = &formatted
		}
	}

	if gameID := r.URL.Query().Get("game_id"); gameID != "" {
		gid := core.GameID(gameID)
		filter.GameID = &gid
	}

	if vsPitcher := r.URL.Query().Get("vs_pitcher"); vsPitcher != "" {
		if retroPitcher, err := pr.lookupRetroID(ctx, core.PlayerID(vsPitcher)); err == nil {
			pid := retroPitcher
			filter.Pitcher = &pid
		}
	} else if pitcher := r.URL.Query().Get("pitcher"); pitcher != "" {
		p := core.RetroPlayerID(pitcher)
		filter.Pitcher = &p
	}

	plays, err := pr.repo.List(ctx, filter)
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
		Data:    plays,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

var errMissingRetroID = errors.New("player does not have a Retrosheet retroID")

func (pr *PlayRoutes) lookupRetroID(ctx context.Context, playerID core.PlayerID) (core.RetroPlayerID, error) {
	player, err := pr.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return "", err
	}
	if player.RetroID == nil || *player.RetroID == "" {
		return "", errMissingRetroID
	}
	return *player.RetroID, nil
}
