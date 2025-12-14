package api

import (
	"net/http"
	"strconv"

	"stormlightlabs.org/baseball/internal/core"
)

type PitchRoutes struct {
	repo core.PitchRepository
}

func NewPitchRoutes(repo core.PitchRepository) *PitchRoutes {
	return &PitchRoutes{repo: repo}
}

func (pr *PitchRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/pitches", pr.handleListPitches)
	mux.HandleFunc("GET /v1/games/{id}/pitches", pr.handleGamePitches)
	mux.HandleFunc("GET /v1/games/{game_id}/plays/{play_num}/pitches", pr.handlePlayPitches)
}

// handleListPitches godoc
// @Summary List pitches
// @Description Query individual pitches derived from play-by-play pitch sequences
// @Tags pitches
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
// @Param top_bot query integer false "Filter by top (0) or bottom (1) of inning" Enums(0, 1)
// @Param pitch_type query string false "Filter by pitch type (B, C, F, S, X, etc.)"
// @Param ball_count query integer false "Filter by ball count (0-3)"
// @Param strike_count query integer false "Filter by strike count (0-2)"
// @Param is_in_play query boolean false "Filter to only pitches in play (X)"
// @Param is_strike query boolean false "Filter to only strikes"
// @Param is_ball query boolean false "Filter to only balls"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /pitches [get]
func (pr *PitchRoutes) handleListPitches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.PitchFilter{
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

	if topBot := r.URL.Query().Get("top_bot"); topBot != "" {
		tb := getIntQuery(r, "top_bot", -1)
		if tb == 0 || tb == 1 {
			filter.TopBot = &tb
		}
	}

	if pitchType := r.URL.Query().Get("pitch_type"); pitchType != "" {
		filter.PitchType = &pitchType
	}

	if ballCount := r.URL.Query().Get("ball_count"); ballCount != "" {
		b := getIntQuery(r, "ball_count", -1)
		if b >= 0 {
			filter.BallCount = &b
		}
	}

	if strikeCount := r.URL.Query().Get("strike_count"); strikeCount != "" {
		s := getIntQuery(r, "strike_count", -1)
		if s >= 0 {
			filter.StrikeCount = &s
		}
	}

	if isInPlay := r.URL.Query().Get("is_in_play"); isInPlay == "true" {
		t := true
		filter.IsInPlay = &t
	}

	if isStrike := r.URL.Query().Get("is_strike"); isStrike == "true" {
		t := true
		filter.IsStrike = &t
	}

	if isBall := r.URL.Query().Get("is_ball"); isBall == "true" {
		t := true
		filter.IsBall = &t
	}

	pitches, err := pr.repo.List(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := pr.repo.Count(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    pitches,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleGamePitches godoc
// @Summary Get pitches for a game
// @Description Get all pitches for a specific game in chronological order
// @Tags pitches, games
// @Accept json
// @Produce json
// @Param id path string true "Game ID (format: YYYYMMDD + game_number + home_team)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(200)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{id}/pitches [get]
func (pr *PitchRoutes) handleGamePitches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gameID := core.GameID(r.PathValue("id"))

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 200),
	}

	pitches, err := pr.repo.ListByGame(ctx, gameID, pagination)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	filter := core.PitchFilter{
		GameID:     &gameID,
		Pagination: core.Pagination{Page: 1, PerPage: 1},
	}
	total, err := pr.repo.Count(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    pitches,
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		Total:   total,
	})
}

// handlePlayPitches godoc
// @Summary Get pitches for a specific plate appearance
// @Description Get all pitches from a specific plate appearance within a game
// @Tags pitches, plays, games
// @Accept json
// @Produce json
// @Param game_id path string true "Game ID"
// @Param play_num path integer true "Play number within the game"
// @Success 200 {object} map[string]any
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{game_id}/plays/{play_num}/pitches [get]
func (pr *PitchRoutes) handlePlayPitches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gameID := core.GameID(r.PathValue("game_id"))

	playNumStr := r.PathValue("play_num")
	playNum, err := strconv.Atoi(playNumStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid play number"})
		return
	}

	pitches, err := pr.repo.ListByPlay(ctx, gameID, playNum)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": pitches})
}
