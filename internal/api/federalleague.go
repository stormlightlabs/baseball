package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type FederalLeagueRoutes struct {
	gameRepo core.GameRepository
	playRepo core.PlayRepository
	teamRepo core.TeamRepository
}

func NewFederalLeagueRoutes(gameRepo core.GameRepository, playRepo core.PlayRepository, teamRepo core.TeamRepository) *FederalLeagueRoutes {
	return &FederalLeagueRoutes{
		gameRepo: gameRepo,
		playRepo: playRepo,
		teamRepo: teamRepo,
	}
}

func (flr *FederalLeagueRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/federalleague/games", flr.handleListGames)
	mux.HandleFunc("GET /v1/federalleague/teams", flr.handleListTeams)
	mux.HandleFunc("GET /v1/federalleague/plays", flr.handleListPlays)
	mux.HandleFunc("GET /v1/federalleague/seasons/{year}/schedule", flr.handleSeasonSchedule)
	mux.HandleFunc("GET /v1/federalleague/seasons/{year}/teams/{team_id}/games", flr.handleTeamGames)
}

// handleListGames godoc
// @Summary List Federal League games
// @Description Get all games from the Federal League (1914-1915)
// @Tags federalleague, games
// @Accept json
// @Produce json
// @Param season query integer false "Filter by season year (1914 or 1915)"
// @Param home_team query string false "Filter by home team ID"
// @Param away_team query string false "Filter by away team ID"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /federalleague/games [get]
func (flr *FederalLeagueRoutes) handleListGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	league := core.LeagueID("FL")

	filter := core.GameFilter{
		League: &league,
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if season := r.URL.Query().Get("season"); season != "" {
		y := core.SeasonYear(getIntQuery(r, "season", 0))
		filter.Season = &y
	}

	if homeTeam := r.URL.Query().Get("home_team"); homeTeam != "" {
		t := core.TeamID(homeTeam)
		filter.HomeTeam = &t
	}

	if awayTeam := r.URL.Query().Get("away_team"); awayTeam != "" {
		t := core.TeamID(awayTeam)
		filter.AwayTeam = &t
	}

	games, err := flr.gameRepo.List(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := flr.gameRepo.Count(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    games,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleListTeams godoc
// @Summary List Federal League teams
// @Description Get all teams that played in the Federal League (1914-1915)
// @Tags federalleague, teams
// @Accept json
// @Produce json
// @Param year query integer false "Filter by season year (1914 or 1915)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /federalleague/teams [get]
func (flr *FederalLeagueRoutes) handleListTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	league := core.LeagueID("FL")

	filter := core.TeamFilter{
		League: &league,
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if year := r.URL.Query().Get("year"); year != "" {
		y := core.SeasonYear(getIntQuery(r, "year", 0))
		filter.Year = &y
	}

	teams, err := flr.teamRepo.ListTeamSeasons(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := flr.teamRepo.CountTeamSeasons(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    teams,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleListPlays godoc
// @Summary List Federal League plays
// @Description Get play-by-play data from Federal League games
// @Tags federalleague, plays
// @Accept json
// @Produce json
// @Param batter query string false "Filter by batter ID"
// @Param pitcher query string false "Filter by pitcher ID"
// @Param team query string false "Filter by team ID (batting team)"
// @Param date_from query string false "Start date (YYYYMMDD)"
// @Param date_to query string false "End date (YYYYMMDD)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(100)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /federalleague/plays [get]
func (flr *FederalLeagueRoutes) handleListPlays(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	league := core.LeagueID("FL")

	filter := core.PlayFilter{
		League: &league,
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 100),
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

	if team := r.URL.Query().Get("team"); team != "" {
		t := core.TeamID(team)
		filter.BatTeam = &t
	}

	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		filter.DateFrom = &dateFrom
	}

	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		filter.DateTo = &dateTo
	}

	plays, err := flr.playRepo.List(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := flr.playRepo.Count(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    plays,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleSeasonSchedule godoc
// @Summary Get Federal League season schedule
// @Description Get all games for a specific Federal League season (1914 or 1915)
// @Tags federalleague, games
// @Accept json
// @Produce json
// @Param year path integer true "Season year (1914 or 1915)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(100)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /federalleague/seasons/{year}/schedule [get]
func (flr *FederalLeagueRoutes) handleSeasonSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	league := core.LeagueID("FL")

	filter := core.GameFilter{
		Season: &year,
		League: &league,
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 100),
		},
	}

	games, err := flr.gameRepo.List(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := flr.gameRepo.Count(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    games,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleTeamGames godoc
// @Summary Get Federal League team games for a season
// @Description Get all games for a specific Federal League team in a season
// @Tags federalleague, games, teams
// @Accept json
// @Produce json
// @Param year path integer true "Season year (1914 or 1915)"
// @Param team_id path string true "Team ID"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(100)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /federalleague/seasons/{year}/teams/{team_id}/games [get]
func (flr *FederalLeagueRoutes) handleTeamGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	teamID := core.TeamID(r.PathValue("team_id"))
	league := core.LeagueID("FL")

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 100),
	}

	games, err := flr.gameRepo.ListByTeamSeason(ctx, teamID, year, pagination)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	filter := core.GameFilter{
		HomeTeam:   &teamID,
		Season:     &year,
		League:     &league,
		Pagination: pagination,
	}

	total, err := flr.gameRepo.Count(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    games,
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		Total:   total,
	})
}
