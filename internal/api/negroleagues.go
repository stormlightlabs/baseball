package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

// NegroLeaguesRoutes handles endpoints for Negro Leagues historical data (1903-1962)
//
// Negro Leagues data in Retrosheet uses various league identifiers including:
// - NAL (Negro American League)
// - NNL (Negro National League)
// - And potentially others
//
// This implementation provides dedicated endpoints for Negro Leagues data.
type NegroLeaguesRoutes struct {
	repo core.NegroLeaguesRepository
}

func NewNegroLeaguesRoutes(repo core.NegroLeaguesRepository) *NegroLeaguesRoutes {
	return &NegroLeaguesRoutes{
		repo: repo,
	}
}

func (nlr *NegroLeaguesRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/negroleagues/games", nlr.handleListGames)
	mux.HandleFunc("GET /v1/negroleagues/teams", nlr.handleListTeams)
	mux.HandleFunc("GET /v1/negroleagues/plays", nlr.handleListPlays)
	mux.HandleFunc("GET /v1/negroleagues/seasons/{year}/schedule", nlr.handleSeasonSchedule)
	mux.HandleFunc("GET /v1/negroleagues/seasons/{year}/teams/{team_id}/games", nlr.handleTeamGames)
}

// handleListGames godoc
// @Summary List Negro Leagues games
// @Description Get all games from the Negro Leagues (1935-1949)
// @Tags negroleagues, games
// @Accept json
// @Produce json
// @Param season query integer false "Filter by season year"
// @Param home_team query string false "Filter by home team ID"
// @Param away_team query string false "Filter by away team ID"
// @Param league query string false "Filter by specific league code (NAL, NNL, etc.)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /negroleagues/games [get]
func (nlr *NegroLeaguesRoutes) handleListGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.GameFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if leagueParam := r.URL.Query().Get("league"); leagueParam != "" {
		league := core.LeagueID(leagueParam)
		filter.League = &league
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

	games, err := nlr.repo.ListGames(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := nlr.repo.CountGames(ctx, filter)
	if err != nil {
		writeError(w, err)
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
// @Summary List Negro Leagues teams
// @Description Get all teams that played in the Negro Leagues (1935-1949)
// @Tags negroleagues, teams
// @Accept json
// @Produce json
// @Param year query integer false "Filter by season year"
// @Param league query string false "Filter by specific league code (NAL, NNL, etc.)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /negroleagues/teams [get]
func (nlr *NegroLeaguesRoutes) handleListTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.TeamFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if leagueParam := r.URL.Query().Get("league"); leagueParam != "" {
		league := core.LeagueID(leagueParam)
		filter.League = &league
	}

	if year := r.URL.Query().Get("year"); year != "" {
		y := core.SeasonYear(getIntQuery(r, "year", 0))
		filter.Year = &y
	}

	teams, err := nlr.repo.ListTeamSeasons(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := nlr.repo.CountTeamSeasons(ctx, filter)
	if err != nil {
		writeError(w, err)
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
// @Summary List Negro Leagues plays
// @Description Get play-by-play data from Negro Leagues games
// @Tags negroleagues, plays
// @Accept json
// @Produce json
// @Param batter query string false "Filter by batter ID"
// @Param pitcher query string false "Filter by pitcher ID"
// @Param team query string false "Filter by team ID (batting team)"
// @Param league query string false "Filter by specific league code (NAL, NNL, etc.)"
// @Param date_from query string false "Start date (YYYYMMDD)"
// @Param date_to query string false "End date (YYYYMMDD)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(100)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /negroleagues/plays [get]
func (nlr *NegroLeaguesRoutes) handleListPlays(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.PlayFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 100),
		},
	}

	if leagueParam := r.URL.Query().Get("league"); leagueParam != "" {
		league := core.LeagueID(leagueParam)
		filter.League = &league
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

	plays, err := nlr.repo.ListPlays(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := nlr.repo.CountPlays(ctx, filter)
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

// handleSeasonSchedule godoc
// @Summary Get Negro Leagues season schedule
// @Description Get all games for a specific Negro Leagues season
// @Tags negroleagues, games
// @Accept json
// @Produce json
// @Param year path integer true "Season year (1935-1949)"
// @Param league query string false "Filter by specific league code (NAL, NNL, etc.)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(100)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /negroleagues/seasons/{year}/schedule [get]
func (nlr *NegroLeaguesRoutes) handleSeasonSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))

	filter := core.GameFilter{
		Season: &year,
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 100),
		},
	}

	var league *core.LeagueID
	if leagueParam := r.URL.Query().Get("league"); leagueParam != "" {
		l := core.LeagueID(leagueParam)
		league = &l
		filter.League = league
	}

	games, err := nlr.repo.GetSeasonSchedule(ctx, year, league, filter.Pagination)
	if err != nil {
		writeError(w, err)
		return
	}

	total, err := nlr.repo.CountGames(ctx, filter)
	if err != nil {
		writeError(w, err)
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
// @Summary Get Negro Leagues team games for a season
// @Description Get all games for a specific Negro Leagues team in a season
// @Tags negroleagues, games, teams
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param team_id path string true "Team ID"
// @Param league query string false "Filter by specific league code (NAL, NNL, etc.)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(100)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /negroleagues/seasons/{year}/teams/{team_id}/games [get]
func (nlr *NegroLeaguesRoutes) handleTeamGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	teamID := core.TeamID(r.PathValue("team_id"))

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 100),
	}

	games, err := nlr.repo.GetTeamGames(ctx, teamID, year, pagination)
	if err != nil {
		writeError(w, err)
		return
	}

	filter := core.GameFilter{
		Season:     &year,
		Pagination: pagination,
	}

	if leagueParam := r.URL.Query().Get("league"); leagueParam != "" {
		league := core.LeagueID(leagueParam)
		filter.League = &league
	}

	filter.HomeTeam = &teamID
	homeCount, err := nlr.repo.CountGames(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	filter.HomeTeam = nil
	filter.AwayTeam = &teamID
	awayCount, err := nlr.repo.CountGames(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	total := homeCount + awayCount

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    games,
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		Total:   total,
	})
}
