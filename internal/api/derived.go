package api

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"stormlightlabs.org/baseball/internal/core"
)

type DerivedRoutes struct {
	repo core.DerivedStatsRepository
}

func NewDerivedRoutes(repo core.DerivedStatsRepository) *DerivedRoutes {
	return &DerivedRoutes{repo: repo}
}

func (dr *DerivedRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/players/{player_id}/streaks", dr.handlePlayerStreaks)
	mux.HandleFunc("GET /v1/players/{player_id}/splits", dr.handlePlayerSplits)
	mux.HandleFunc("GET /v1/teams/{team_id}/run-differential", dr.handleTeamRunDifferential)
	mux.HandleFunc("GET /v1/games/{game_id}/win-probability", dr.handleGameWinProbability)
}

// handlePlayerStreaks godoc
// @Summary Get player streaks
// @Description Get hitting or scoreless innings streaks for a player
// @Tags derived, players
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID"
// @Param kind query string true "Streak kind: hitting or scoreless_innings"
// @Param season query integer true "Season year"
// @Param min_length query integer false "Minimum streak length" default(5)
// @Success 200 {array} core.Streak
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{player_id}/streaks [get]
func (dr *DerivedRoutes) handlePlayerStreaks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playerID := core.PlayerID(r.PathValue("player_id"))

	kindStr := r.URL.Query().Get("kind")
	if kindStr == "" {
		writeBadRequest(w, "kind query parameter required (hitting or scoreless_innings)")
		return
	}

	kind := core.StreakKind(kindStr)
	if kind != core.StreakKindHitting && kind != core.StreakKindScorelessInnings {
		writeBadRequest(w, "invalid kind: must be 'hitting' or 'scoreless_innings'")
		return
	}

	seasonStr := r.URL.Query().Get("season")
	if seasonStr == "" {
		writeBadRequest(w, "season query parameter required")
		return
	}

	season := core.SeasonYear(getIntQuery(r, "season", 2024))
	minLength := getIntQuery(r, "min_length", 5)

	streaks, err := dr.repo.PlayerStreaks(ctx, playerID, kind, season, minLength)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, streaks)
}

// handleTeamRunDifferential godoc
// @Summary Get team run differential
// @Description Get season run differential with rolling windows for a team
// @Tags derived, teams
// @Accept json
// @Produce json
// @Param team_id path string true "Team ID"
// @Param season query integer true "Season year"
// @Param windows query string false "Comma-separated rolling window sizes (e.g., 10,20,30)" default("10,20,30")
// @Success 200 {object} core.RunDifferentialSeries
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /teams/{team_id}/run-differential [get]
func (dr *DerivedRoutes) handleTeamRunDifferential(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamID := core.TeamID(r.PathValue("team_id"))

	seasonStr := r.URL.Query().Get("season")
	if seasonStr == "" {
		writeBadRequest(w, "season query parameter required")
		return
	}

	season := core.SeasonYear(getIntQuery(r, "season", 2024))
	windowsStr := r.URL.Query().Get("windows")
	if windowsStr == "" {
		windowsStr = "10,20,30"
	}

	var windows []int
	for _, wStr := range strings.Split(windowsStr, ",") {
		windowSize, err := strconv.Atoi(strings.TrimSpace(wStr))
		if err != nil {
			writeBadRequest(w, "invalid windows parameter: must be comma-separated integers")
			return
		}
		windows = append(windows, windowSize)
	}

	series, err := dr.repo.TeamRunDifferential(ctx, teamID, season, windows)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, series)
}

// handleGameWinProbability godoc
// @Summary Get game win probability curve
// @Description Get play-by-play win probability for a game
// @Tags derived, games
// @Accept json
// @Produce json
// @Param game_id path string true "Game ID"
// @Success 200 {object} core.WinProbabilityCurve
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{game_id}/win-probability [get]
func (dr *DerivedRoutes) handleGameWinProbability(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gameID := core.GameID(r.PathValue("game_id"))

	curve, err := dr.repo.GameWinProbability(ctx, gameID)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	if len(curve.Points) == 0 {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error: "game not found or no play-by-play data available",
		})
		return
	}

	writeJSON(w, http.StatusOK, curve)
}

// handlePlayerSplits godoc
// @Summary Get player batting splits
// @Description Get batting statistics split by dimension (home/away, vs handedness, month, etc.)
// @Tags derived, players
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID"
// @Param dimension query string true "Split dimension: home_away, pitcher_handed, or month"
// @Param season query integer true "Season year"
// @Success 200 {object} core.SplitResult
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{player_id}/splits [get]
func (dr *DerivedRoutes) handlePlayerSplits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playerID := core.PlayerID(r.PathValue("player_id"))

	dimensionStr := r.URL.Query().Get("dimension")
	if dimensionStr == "" {
		writeBadRequest(w, "dimension query parameter required (home_away, pitcher_handed, or month)")
		return
	}

	dimension := core.SplitDimension(dimensionStr)
	validDimensions := []core.SplitDimension{core.SplitDimHomeAway, core.SplitDimPitcherHanded, core.SplitDimMonth}

	valid := slices.Contains(validDimensions, dimension)

	if !valid {
		writeBadRequest(w, "invalid dimension: must be 'home_away', 'pitcher_handed', or 'month'")
		return
	}

	seasonStr := r.URL.Query().Get("season")
	if seasonStr == "" {
		writeBadRequest(w, "season query parameter required")
		return
	}

	season := core.SeasonYear(getIntQuery(r, "season", 2024))

	splits, err := dr.repo.PlayerSplits(ctx, playerID, dimension, season)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, splits)
}
