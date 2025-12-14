package api

import (
	"net/http"
	"strconv"

	"stormlightlabs.org/baseball/internal/core"
)

type ComputedRoutes struct {
	advancedRepo   core.AdvancedStatsRepository
	leverageRepo   core.LeverageRepository
	parkFactorRepo core.ParkFactorRepository
}

func NewComputedRoutes(a core.AdvancedStatsRepository, l core.LeverageRepository, p core.ParkFactorRepository) *ComputedRoutes {
	return &ComputedRoutes{advancedRepo: a, leverageRepo: l, parkFactorRepo: p}
}

func (cr *ComputedRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/players/{player_id}/stats/batting/advanced", cr.handlePlayerAdvancedBatting)
	mux.HandleFunc("GET /v1/players/{player_id}/stats/pitching/advanced", cr.handlePlayerAdvancedPitching)
	mux.HandleFunc("GET /v1/players/{player_id}/stats/baserunning", cr.handlePlayerBaserunning)
	mux.HandleFunc("GET /v1/players/{player_id}/stats/fielding", cr.handlePlayerFielding)
	mux.HandleFunc("GET /v1/players/{player_id}/stats/war", cr.handlePlayerWAR)
	mux.HandleFunc("GET /v1/players/{player_id}/leverage/summary", cr.handlePlayerLeverageSummary)
	mux.HandleFunc("GET /v1/players/{player_id}/leverage/high", cr.handlePlayerHighLeveragePAs)
	mux.HandleFunc("GET /v1/games/{game_id}/plate-appearances/leverage", cr.handleGamePlateLeverages)
	mux.HandleFunc("GET /v1/games/{game_id}/win-probability/summary", cr.handleGameWinProbabilitySummary)
	mux.HandleFunc("GET /v1/parks/{park_id}/factors", cr.handleParkFactor)
	mux.HandleFunc("GET /v1/parks/{park_id}/factors/series", cr.handleParkFactorSeries)
	mux.HandleFunc("GET /v1/seasons/{season}/park-factors", cr.handleSeasonParkFactors)
	mux.HandleFunc("GET /v1/seasons/{season}/leaders/batting/advanced", cr.handleSeasonBattingLeaders)
	mux.HandleFunc("GET /v1/seasons/{season}/leaders/pitching/advanced", cr.handleSeasonPitchingLeaders)
	mux.HandleFunc("GET /v1/seasons/{season}/leaders/war", cr.handleSeasonWARLeaders)
}

// handlePlayerAdvancedBatting godoc
// @Summary Get advanced batting stats
// @Description Get wOBA, wRC+, ISO, BABIP, and other advanced batting stats for a player
// @Tags computed, players, batting
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID"
// @Param season query integer false "Season year" default(2024)
// @Param team_id query string false "Team ID"
// @Success 200 {object} core.AdvancedBattingStats
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{player_id}/stats/batting/advanced [get]
func (cr *ComputedRoutes) handlePlayerAdvancedBatting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playerID := core.PlayerID(r.PathValue("player_id"))

	filter := core.AdvancedBattingFilter{}

	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		season := core.SeasonYear(getIntQuery(r, "season", 2024))
		filter.Season = &season
	}

	if teamIDStr := r.URL.Query().Get("team_id"); teamIDStr != "" {
		teamID := core.TeamID(teamIDStr)
		filter.TeamID = &teamID
	}

	stats, err := cr.advancedRepo.PlayerAdvancedBatting(ctx, playerID, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// handlePlayerAdvancedPitching godoc
// @Summary Get advanced pitching stats
// @Description Get FIP, xFIP, ERA+, and other advanced pitching stats for a player
// @Tags computed, players, pitching
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID"
// @Param season query integer false "Season year" default(2024)
// @Param team_id query string false "Team ID"
// @Success 200 {object} core.AdvancedPitchingStats
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{player_id}/stats/pitching/advanced [get]
func (cr *ComputedRoutes) handlePlayerAdvancedPitching(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playerID := core.PlayerID(r.PathValue("player_id"))

	filter := core.AdvancedPitchingFilter{}

	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		season := core.SeasonYear(getIntQuery(r, "season", 2024))
		filter.Season = &season
	}

	if teamIDStr := r.URL.Query().Get("team_id"); teamIDStr != "" {
		teamID := core.TeamID(teamIDStr)
		filter.TeamID = &teamID
	}

	stats, err := cr.advancedRepo.PlayerAdvancedPitching(ctx, playerID, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// handleGamePlateLeverages godoc
// @Summary Get plate appearance leverages
// @Description Get leverage index for each plate appearance in a game
// @Tags computed, games, leverage
// @Accept json
// @Produce json
// @Param game_id path string true "Game ID"
// @Param min_li query number false "Minimum leverage index" default(0.0)
// @Success 200 {array} core.PlateAppearanceLeverage
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{game_id}/plate-appearances/leverage [get]
func (cr *ComputedRoutes) handleGamePlateLeverages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gameID := core.GameID(r.PathValue("game_id"))

	var minLI *float64
	if minLIStr := r.URL.Query().Get("min_li"); minLIStr != "" {
		minLIFloat, err := strconv.ParseFloat(minLIStr, 64)
		if err != nil {
			writeBadRequest(w, "invalid min_li: must be a number")
			return
		}
		minLI = &minLIFloat
	}

	leverages, err := cr.leverageRepo.GamePlateLeverages(ctx, gameID, minLI)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, leverages)
}

// handleParkFactor godoc
// @Summary Get park factor
// @Description Get park factor for a specific park and season
// @Tags computed, parks
// @Accept json
// @Produce json
// @Param park_id path string true "Park ID"
// @Param season query integer true "Season year"
// @Success 200 {object} core.ParkFactor
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /parks/{park_id}/factors [get]
func (cr *ComputedRoutes) handleParkFactor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parkID := core.ParkID(r.PathValue("park_id"))
	seasonStr := r.URL.Query().Get("season")
	if seasonStr == "" {
		writeBadRequest(w, "season query parameter required")
		return
	}
	season := core.SeasonYear(getIntQuery(r, "season", 2024))
	factor, err := cr.parkFactorRepo.ParkFactor(ctx, parkID, season)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, factor)
}

// handleParkFactorSeries godoc
// @Summary Get park factor series
// @Description Get park factors over a range of seasons
// @Tags computed, parks
// @Accept json
// @Produce json
// @Param park_id path string true "Park ID"
// @Param from_season query integer true "Starting season"
// @Param to_season query integer true "Ending season"
// @Success 200 {array} core.ParkFactor
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /parks/{park_id}/factors/series [get]
func (cr *ComputedRoutes) handleParkFactorSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parkID := core.ParkID(r.PathValue("park_id"))

	fromSeasonStr := r.URL.Query().Get("from_season")
	toSeasonStr := r.URL.Query().Get("to_season")

	if fromSeasonStr == "" || toSeasonStr == "" {
		writeBadRequest(w, "from_season and to_season query parameters required")
		return
	}

	fromSeason := core.SeasonYear(getIntQuery(r, "from_season", 2020))
	toSeason := core.SeasonYear(getIntQuery(r, "to_season", 2024))

	factors, err := cr.parkFactorRepo.ParkFactorSeries(ctx, parkID, fromSeason, toSeason)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, factors)
}

// handleSeasonParkFactors godoc
// @Summary Get all park factors for a season
// @Description Get park factors for all parks in a given season
// @Tags computed, parks
// @Accept json
// @Produce json
// @Param season path integer true "Season year"
// @Param type query string false "Factor type (runs, hr)"
// @Success 200 {array} core.ParkFactor
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{season}/park-factors [get]
func (cr *ComputedRoutes) handleSeasonParkFactors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	season := getIntPathValue(r, "season")
	if season == 0 {
		writeBadRequest(w, "invalid season")
		return
	}

	var factorType *string
	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		factorType = &typeStr
	}

	factors, err := cr.parkFactorRepo.SeasonParkFactors(ctx, core.SeasonYear(season), factorType)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, factors)
}

// handlePlayerBaserunning godoc
// @Summary Get baserunning stats
// @Description Get baserunning value (wSB) for a player
// @Tags computed, players
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID"
// @Param season query integer false "Season year" default(2024)
// @Param team_id query string false "Team ID"
// @Success 200 {object} core.BaserunningStats
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{player_id}/stats/baserunning [get]
func (cr *ComputedRoutes) handlePlayerBaserunning(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playerID := core.PlayerID(r.PathValue("player_id"))
	season := core.SeasonYear(getIntQuery(r, "season", 2024))

	var teamID *core.TeamID
	if teamIDStr := r.URL.Query().Get("team_id"); teamIDStr != "" {
		tid := core.TeamID(teamIDStr)
		teamID = &tid
	}

	stats, err := cr.advancedRepo.PlayerBaserunning(ctx, playerID, season, teamID)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// handlePlayerFielding godoc
// @Summary Get fielding stats
// @Description Get fielding runs above average for a player
// @Tags computed, players
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID"
// @Param season query integer false "Season year" default(2024)
// @Param team_id query string false "Team ID"
// @Success 200 {object} core.FieldingStats
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{player_id}/stats/fielding [get]
func (cr *ComputedRoutes) handlePlayerFielding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playerID := core.PlayerID(r.PathValue("player_id"))
	season := core.SeasonYear(getIntQuery(r, "season", 2024))

	var teamID *core.TeamID
	if teamIDStr := r.URL.Query().Get("team_id"); teamIDStr != "" {
		tid := core.TeamID(teamIDStr)
		teamID = &tid
	}

	stats, err := cr.advancedRepo.PlayerFielding(ctx, playerID, season, teamID)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// handlePlayerWAR godoc
// @Summary Get WAR
// @Description Get Wins Above Replacement and component breakdown for a player
// @Tags computed, players
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID"
// @Param season query integer false "Season year" default(2024)
// @Param team_id query string false "Team ID"
// @Success 200 {object} core.PlayerWARSummary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{player_id}/stats/war [get]
func (cr *ComputedRoutes) handlePlayerWAR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playerID := core.PlayerID(r.PathValue("player_id"))
	filter := core.WARFilter{}
	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		season := core.SeasonYear(getIntQuery(r, "season", 2024))
		filter.Season = &season
	}

	if teamIDStr := r.URL.Query().Get("team_id"); teamIDStr != "" {
		teamID := core.TeamID(teamIDStr)
		filter.TeamID = &teamID
	}

	war, err := cr.advancedRepo.PlayerWAR(ctx, playerID, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, war)
}

// handleSeasonBattingLeaders godoc
// @Summary Get advanced batting leaders
// @Description Get top players by advanced batting stats (wOBA, wRC+, ISO, BABIP, etc.)
// @Tags computed, leaders, batting
// @Accept json
// @Produce json
// @Param season path integer true "Season year"
// @Param stat query string false "Statistic (WOBA, WRC_PLUS, ISO, BABIP, AVG, OBP, SLG, HR, BB, K_RATE, BB_RATE)" default("WRC_PLUS")
// @Param limit query integer false "Number of results" default(10)
// @Param min_pa query integer false "Minimum plate appearances" default(502)
// @Param team_id query string false "Filter by team ID"
// @Param league query string false "Filter by league (AL, NL)"
// @Success 200 {array} core.AdvancedBattingStats
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{season}/leaders/batting/advanced [get]
func (cr *ComputedRoutes) handleSeasonBattingLeaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	season := core.SeasonYear(getIntPathValue(r, "season"))

	stat := r.URL.Query().Get("stat")
	if stat == "" {
		stat = "WRC_PLUS"
	}

	limit := getIntQuery(r, "limit", 10)

	filter := core.AdvancedBattingFilter{}

	if minPAStr := r.URL.Query().Get("min_pa"); minPAStr != "" {
		minPA := getIntQuery(r, "min_pa", 502)
		filter.MinPA = &minPA
	}

	if teamIDStr := r.URL.Query().Get("team_id"); teamIDStr != "" {
		teamID := core.TeamID(teamIDStr)
		filter.TeamID = &teamID
	}

	if leagueStr := r.URL.Query().Get("league"); leagueStr != "" {
		league := core.LeagueID(leagueStr)
		filter.League = &league
	}

	leaders, err := cr.advancedRepo.SeasonBattingLeaders(ctx, season, stat, limit, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, leaders)
}

// handleSeasonPitchingLeaders godoc
// @Summary Get advanced pitching leaders
// @Description Get top pitchers by advanced pitching stats (FIP, ERA, WHIP, K/9, etc.)
// @Tags computed, leaders, pitching
// @Accept json
// @Produce json
// @Param season path integer true "Season year"
// @Param stat query string false "Statistic (FIP, ERA, WHIP, K_PER_9, BB_PER_9, HR_PER_9, SO)" default("FIP")
// @Param limit query integer false "Number of results" default(10)
// @Param min_ip query number false "Minimum innings pitched" default(162)
// @Param team_id query string false "Filter by team ID"
// @Success 200 {array} core.AdvancedPitchingStats
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{season}/leaders/pitching/advanced [get]
func (cr *ComputedRoutes) handleSeasonPitchingLeaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	season := core.SeasonYear(getIntPathValue(r, "season"))
	stat := r.URL.Query().Get("stat")
	if stat == "" {
		stat = "FIP"
	}

	limit := getIntQuery(r, "limit", 10)
	filter := core.AdvancedPitchingFilter{}

	if minIPStr := r.URL.Query().Get("min_ip"); minIPStr != "" {
		minIPFloat, err := strconv.ParseFloat(minIPStr, 64)
		if err != nil {
			writeBadRequest(w, "invalid min_ip: must be a number")
			return
		}
		filter.MinIP = &minIPFloat
	}

	if teamIDStr := r.URL.Query().Get("team_id"); teamIDStr != "" {
		teamID := core.TeamID(teamIDStr)
		filter.TeamID = &teamID
	}

	leaders, err := cr.advancedRepo.SeasonPitchingLeaders(ctx, season, stat, limit, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, leaders)
}

// handleSeasonWARLeaders godoc
// @Summary Get WAR leaders
// @Description Get top players by Wins Above Replacement
// @Tags computed, leaders, war
// @Accept json
// @Produce json
// @Param season path integer true "Season year"
// @Param limit query integer false "Number of results" default(10)
// @Param min_pa query integer false "Minimum plate appearances" default(502)
// @Param team_id query string false "Filter by team ID"
// @Success 200 {array} core.PlayerWARSummary
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{season}/leaders/war [get]
func (cr *ComputedRoutes) handleSeasonWARLeaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	season := core.SeasonYear(getIntPathValue(r, "season"))
	limit := getIntQuery(r, "limit", 10)
	filter := core.WARFilter{}

	if minPAStr := r.URL.Query().Get("min_pa"); minPAStr != "" {
		minPA := getIntQuery(r, "min_pa", 502)
		filter.MinPA = &minPA
	}

	if teamIDStr := r.URL.Query().Get("team_id"); teamIDStr != "" {
		teamID := core.TeamID(teamIDStr)
		filter.TeamID = &teamID
	}

	leaders, err := cr.advancedRepo.SeasonWARLeaders(ctx, season, limit, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, leaders)
}

// handleGameWinProbabilitySummary godoc
// @Summary Get game win probability summary
// @Description Get summary statistics for a game's win probability including biggest swings
// @Tags computed, games, leverage
// @Accept json
// @Produce json
// @Param game_id path string true "Game ID"
// @Success 200 {object} core.GameWinProbabilitySummary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{game_id}/win-probability/summary [get]
func (cr *ComputedRoutes) handleGameWinProbabilitySummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gameID := core.GameID(r.PathValue("game_id"))
	summary, err := cr.leverageRepo.GameWinProbabilitySummary(ctx, gameID)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// handlePlayerHighLeveragePAs godoc
// @Summary Get high leverage plate appearances
// @Description Get high-leverage plate appearances for a player in a season
// @Tags computed, players, leverage
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID"
// @Param season query integer false "Season year" default(2024)
// @Param min_li query number false "Minimum leverage index" default(1.5)
// @Success 200 {array} core.PlateAppearanceLeverage
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{player_id}/leverage/high [get]
func (cr *ComputedRoutes) handlePlayerHighLeveragePAs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playerID := core.PlayerID(r.PathValue("player_id"))
	season := core.SeasonYear(getIntQuery(r, "season", 2024))
	minLI := 1.5
	if minLIStr := r.URL.Query().Get("min_li"); minLIStr != "" {
		minLIFloat, err := strconv.ParseFloat(minLIStr, 64)
		if err != nil {
			writeBadRequest(w, "invalid min_li: must be a number")
			return
		}
		minLI = minLIFloat
	}

	leverages, err := cr.leverageRepo.PlayerHighLeveragePAs(ctx, playerID, season, minLI)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, leverages)
}

// handlePlayerLeverageSummary godoc
// @Summary Get player leverage summary
// @Description Get aggregated leverage metrics for a player in a season
// @Tags computed, players, leverage
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID"
// @Param season query integer false "Season year" default(2024)
// @Param role query string false "Role filter (batter, pitcher)" default("")
// @Success 200 {object} core.PlayerLeverageSummary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{player_id}/leverage/summary [get]
func (cr *ComputedRoutes) handlePlayerLeverageSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playerID := core.PlayerID(r.PathValue("player_id"))
	season := core.SeasonYear(getIntQuery(r, "season", 2024))
	role := r.URL.Query().Get("role")
	summary, err := cr.leverageRepo.PlayerLeverageSummary(ctx, playerID, season, role)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
