package api

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"stormlightlabs.org/baseball/internal/core"
)

type DerivedRoutes struct {
	repo   core.DerivedStatsRepository
	weRepo core.WinExpectancyRepository
}

func NewDerivedRoutes(repo core.DerivedStatsRepository, weRepo core.WinExpectancyRepository) *DerivedRoutes {
	return &DerivedRoutes{
		repo:   repo,
		weRepo: weRepo,
	}
}

func (dr *DerivedRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/players/{player_id}/streaks", dr.handlePlayerStreaks)
	mux.HandleFunc("GET /v1/players/{player_id}/splits", dr.handlePlayerSplits)
	mux.HandleFunc("GET /v1/teams/{team_id}/run-differential", dr.handleTeamRunDifferential)
	mux.HandleFunc("GET /v1/games/{game_id}/win-probability", dr.handleGameWinProbability)
	mux.HandleFunc("GET /v1/win-expectancy", dr.handleGetWinExpectancy)
	mux.HandleFunc("GET /v1/win-expectancy/eras", dr.handleListWinExpectancyEras)
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
		writeNotFound(w, "game/play-by-play data")
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

// handleGetWinExpectancy godoc
// @Summary Get win expectancy for a game state
// @Description Get the historical win probability for a specific game situation
// @Tags derived, win-expectancy
// @Accept json
// @Produce json
// @Param inning query integer true "Inning (1-9)"
// @Param is_bottom query boolean true "Bottom of inning (true/false)"
// @Param outs query integer true "Outs (0-2)"
// @Param runners query string true "Runners state (e.g., ___, 1__, 12_, 123)"
// @Param score_diff query integer true "Score differential from home team perspective (-11 to +11)"
// @Param start_year query integer false "Start year for historical era filter"
// @Param end_year query integer false "End year for historical era filter"
// @Success 200 {object} core.WinExpectancy
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /win-expectancy [get]
func (dr *DerivedRoutes) handleGetWinExpectancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse required parameters
	inning := getIntQuery(r, "inning", 0)
	if inning < 1 || inning > 9 {
		writeBadRequest(w, "inning must be between 1 and 9")
		return
	}

	isBottomStr := r.URL.Query().Get("is_bottom")
	if isBottomStr == "" {
		writeBadRequest(w, "is_bottom query parameter required (true or false)")
		return
	}
	isBottom := isBottomStr == "true"

	outs := getIntQuery(r, "outs", -1)
	if outs < 0 || outs > 2 {
		writeBadRequest(w, "outs must be between 0 and 2")
		return
	}

	runners := r.URL.Query().Get("runners")
	if runners == "" {
		writeBadRequest(w, "runners query parameter required (e.g., ___, 1__, 123)")
		return
	}

	scoreDiff := getIntQuery(r, "score_diff", 0)

	state := core.GameState{
		Inning:      inning,
		IsBottom:    isBottom,
		Outs:        outs,
		RunnersCode: runners,
		ScoreDiff:   scoreDiff,
	}

	startYearStr := r.URL.Query().Get("start_year")
	endYearStr := r.URL.Query().Get("end_year")

	var we *core.WinExpectancy
	var err error

	if startYearStr != "" || endYearStr != "" {
		var startYear, endYear *int
		if startYearStr != "" {
			y := getIntQuery(r, "start_year", 0)
			startYear = &y
		}
		if endYearStr != "" {
			y := getIntQuery(r, "end_year", 0)
			endYear = &y
		}
		we, err = dr.weRepo.GetWinExpectancyForEra(ctx, state, startYear, endYear)
	} else {
		we, err = dr.weRepo.GetWinExpectancy(ctx, state)
	}

	if err != nil {
		writeNotFound(w, "win expectancy data for this game state")
		return
	}

	writeJSON(w, http.StatusOK, we)
}

// handleListWinExpectancyEras godoc
// @Summary List available win expectancy eras
// @Description Get all available historical eras in the win expectancy database
// @Tags derived, win-expectancy
// @Accept json
// @Produce json
// @Success 200 {array} core.WinExpectancyEra
// @Failure 500 {object} ErrorResponse
// @Router /win-expectancy/eras [get]
func (dr *DerivedRoutes) handleListWinExpectancyEras(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	eras, err := dr.weRepo.ListAvailableEras(ctx)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, eras)
}
