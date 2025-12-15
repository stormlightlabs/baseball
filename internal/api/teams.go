package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type TeamRoutes struct {
	repo     core.TeamRepository
	gameRepo core.GameRepository
}

func NewTeamRoutes(repo core.TeamRepository, gameRepo core.GameRepository) *TeamRoutes {
	return &TeamRoutes{repo: repo, gameRepo: gameRepo}
}

func (tr *TeamRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/teams", tr.handleListTeams)
	mux.HandleFunc("GET /v1/teams/{id}", tr.handleGetTeam)
	mux.HandleFunc("GET /v1/teams/{id}/daily-stats", tr.handleTeamDailyStats)
	mux.HandleFunc("GET /v1/seasons", tr.handleListSeasons)
	mux.HandleFunc("GET /v1/seasons/{year}/teams", tr.handleSeasonTeams)
	mux.HandleFunc("GET /v1/seasons/{year}/teams/{team_id}/roster", tr.handleTeamRoster)
	mux.HandleFunc("GET /v1/seasons/{year}/teams/{team_id}/batting", tr.handleTeamBatting)
	mux.HandleFunc("GET /v1/seasons/{year}/teams/{team_id}/pitching", tr.handleTeamPitching)
	mux.HandleFunc("GET /v1/seasons/{year}/teams/{team_id}/fielding", tr.handleTeamFielding)
	mux.HandleFunc("GET /v1/seasons/{year}/teams/{team_id}/schedule", tr.handleTeamSchedule)
	mux.HandleFunc("GET /v1/seasons/{year}/teams/{team_id}/daily-logs", tr.handleTeamDailyLogs)
	mux.HandleFunc("GET /v1/franchises", tr.handleListFranchises)
	mux.HandleFunc("GET /v1/franchises/{id}", tr.handleGetFranchise)
}

// handleListTeams godoc
// @Summary List team seasons
// @Description List team seasons with optional year and league filters
// @Tags teams
// @Accept json
// @Produce json
// @Param year query integer false "Filter by year"
// @Param league query string false "Filter by league (AL, NL)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /teams [get]
func (tr *TeamRoutes) handleListTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.TeamFilter{
		Pagination: *core.NewPagination(getIntQuery(r, "page", 1), getIntQuery(r, "per_page", 50)),
	}

	if year := r.URL.Query().Get("year"); year != "" {
		y := core.SeasonYear(getIntQuery(r, "year", 0))
		filter.Year = &y
	}

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	teams, err := tr.repo.ListTeamSeasons(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := tr.repo.CountTeamSeasons(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, NewPaginatedResponse(teams, filter.Pagination.Page, filter.Pagination.PerPage, total))
}

// handleGetTeam godoc
// @Summary Get team season
// @Description Get a single team-season record
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param year query integer false "Year" default(2024)
// @Success 200 {object} core.TeamSeason
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /teams/{id} [get]
func (tr *TeamRoutes) handleGetTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.TeamID(r.PathValue("id"))
	year := core.SeasonYear(getIntQuery(r, "year", 2024))

	team, err := tr.repo.GetTeamSeason(ctx, id, year)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, team)
}

// handleSeasonTeams godoc
// @Summary Get all teams for a season
// @Description List all teams that played in a specific season
// @Tags teams
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param league query string false "Filter by league (AL, NL)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/teams [get]
func (tr *TeamRoutes) handleSeasonTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntQuery(r, "year", 2024))

	filter := core.TeamFilter{
		Year:       &year,
		Pagination: *core.NewPagination(getIntQuery(r, "page", 1), getIntQuery(r, "per_page", 50)),
	}

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	teams, err := tr.repo.ListTeamSeasons(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := tr.repo.CountTeamSeasons(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, NewPaginatedResponse(teams, filter.Pagination.Page, filter.Pagination.PerPage, total))
}

// handleListFranchises godoc
// @Summary List franchises
// @Description List all baseball franchises with optional active filter
// @Tags franchises
// @Accept json
// @Produce json
// @Param active query boolean false "Filter to only active franchises"
// @Success 200 {object} FranchisesResponse
// @Failure 500 {object} ErrorResponse
// @Router /franchises [get]
func (tr *TeamRoutes) handleListFranchises(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	onlyActive := r.URL.Query().Get("active") == "true"

	franchises, err := tr.repo.ListFranchises(ctx, onlyActive)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, FranchisesResponse{
		Franchises: franchises,
		Total:      len(franchises),
	})
}

// handleListSeasons godoc
// @Summary List all seasons
// @Description Get summary of all available seasons with league and team counts
// @Tags seasons
// @Accept json
// @Produce json
// @Success 200 {array} core.Season
// @Failure 500 {object} ErrorResponse
// @Router /seasons [get]
func (tr *TeamRoutes) handleListSeasons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	seasons, err := tr.repo.ListSeasons(ctx)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, seasons)
}

// handleTeamRoster godoc
// @Summary Get team roster
// @Description Get roster for a specific team and season with basic stats
// @Tags teams
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param team_id path string true "Team ID"
// @Success 200 {array} core.RosterPlayer
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/teams/{team_id}/roster [get]
func (tr *TeamRoutes) handleTeamRoster(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	teamID := core.TeamID(r.PathValue("team_id"))

	roster, err := tr.repo.Roster(ctx, year, teamID)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, roster)
}

// handleTeamBatting godoc
// @Summary Get team batting stats
// @Description Get aggregated batting statistics for a team with optional per-player splits
// @Tags teams
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param team_id path string true "Team ID"
// @Param players query boolean false "Include per-player splits"
// @Success 200 {object} core.TeamBattingStats
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/teams/{team_id}/batting [get]
func (tr *TeamRoutes) handleTeamBatting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	teamID := core.TeamID(r.PathValue("team_id"))
	includePlayers := r.URL.Query().Get("players") == "true"

	stats, err := tr.repo.BattingStats(ctx, year, teamID, includePlayers)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// handleTeamPitching godoc
// @Summary Get team pitching stats
// @Description Get aggregated pitching statistics for a team with optional per-player splits
// @Tags teams
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param team_id path string true "Team ID"
// @Param players query boolean false "Include per-player splits"
// @Success 200 {object} core.TeamPitchingStats
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/teams/{team_id}/pitching [get]
func (tr *TeamRoutes) handleTeamPitching(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	teamID := core.TeamID(r.PathValue("team_id"))
	includePlayers := r.URL.Query().Get("players") == "true"

	stats, err := tr.repo.PitchingStats(ctx, year, teamID, includePlayers)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// handleTeamFielding godoc
// @Summary Get team fielding stats
// @Description Get aggregated fielding statistics for a team with optional per-player/position splits
// @Tags teams
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param team_id path string true "Team ID"
// @Param players query boolean false "Include per-player splits"
// @Success 200 {object} core.TeamFieldingStats
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/teams/{team_id}/fielding [get]
func (tr *TeamRoutes) handleTeamFielding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	teamID := core.TeamID(r.PathValue("team_id"))
	includePlayers := r.URL.Query().Get("players") == "true"

	stats, err := tr.repo.FieldingStats(ctx, year, teamID, includePlayers)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// handleGetFranchise godoc
// @Summary Get franchise
// @Description Get details for a specific franchise
// @Tags franchises
// @Accept json
// @Produce json
// @Param id path string true "Franchise ID"
// @Success 200 {object} core.Franchise
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /franchises/{id} [get]
func (tr *TeamRoutes) handleGetFranchise(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.FranchiseID(r.PathValue("id"))

	franchise, err := tr.repo.GetFranchise(ctx, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, franchise)
}

// handleTeamSchedule godoc
// @Summary Get team schedule
// @Description Get the game schedule for a team in a season (alias for /seasons/{year}/teams/{team_id}/games)
// @Tags teams, games
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param team_id path string true "Team ID"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(100)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/teams/{team_id}/schedule [get]
func (tr *TeamRoutes) handleTeamSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	teamID := core.TeamID(r.PathValue("team_id"))

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 100),
	}

	filter := core.GameFilter{
		Season:     &year,
		Pagination: pagination,
	}

	homeFilter := filter
	homeFilter.HomeTeam = &teamID
	homeGames, err := tr.gameRepo.List(ctx, homeFilter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	awayFilter := filter
	awayFilter.AwayTeam = &teamID
	awayGames, err := tr.gameRepo.List(ctx, awayFilter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	allGames := append(homeGames, awayGames...)

	homeCount, err := tr.gameRepo.Count(ctx, homeFilter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	awayCount, err := tr.gameRepo.Count(ctx, awayFilter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, NewPaginatedResponse(allGames, pagination.Page, pagination.PerPage, homeCount+awayCount))
}

// handleTeamDailyLogs godoc
// @Summary Get team daily logs
// @Description Get team performance aggregated by date for a season
// @Tags teams, games
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param team_id path string true "Team ID"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(100)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/teams/{team_id}/daily-logs [get]
func (tr *TeamRoutes) handleTeamDailyLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	teamID := core.TeamID(r.PathValue("team_id"))
	pagination := core.NewPagination(getIntQuery(r, "page", 1), getIntQuery(r, "per_page", 100))
	filter := core.GameFilter{Season: &year, Pagination: *core.NewPagination(1, 1000)}

	homeFilter := filter
	homeFilter.HomeTeam = &teamID
	homeGames, err := tr.gameRepo.List(ctx, homeFilter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	awayFilter := filter
	awayFilter.AwayTeam = &teamID
	awayGames, err := tr.gameRepo.List(ctx, awayFilter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	type DailyLog struct {
		Date        string `json:"date"`
		GamesPlayed int    `json:"games_played"`
		Wins        int    `json:"wins"`
		Losses      int    `json:"losses"`
		RunsScored  int    `json:"runs_scored"`
		RunsAllowed int    `json:"runs_allowed"`
		RunDiff     int    `json:"run_diff"`
	}

	dailyLogs := make(map[string]*DailyLog)

	processGames := func(games []core.Game, isHome bool) {
		for _, game := range games {
			dateKey := game.Date.Format("2006-01-02")
			if dailyLogs[dateKey] == nil {
				dailyLogs[dateKey] = &DailyLog{
					Date: dateKey,
				}
			}
			log := dailyLogs[dateKey]
			log.GamesPlayed++

			var runsScored, runsAllowed int
			if isHome {
				runsScored = game.HomeScore
				runsAllowed = game.AwayScore
			} else {
				runsScored = game.AwayScore
				runsAllowed = game.HomeScore
			}

			log.RunsScored += runsScored
			log.RunsAllowed += runsAllowed

			if runsScored > runsAllowed {
				log.Wins++
			} else {
				log.Losses++
			}
		}
	}

	processGames(homeGames, true)
	processGames(awayGames, false)

	logs := make([]DailyLog, 0, len(dailyLogs))
	for _, log := range dailyLogs {
		log.RunDiff = log.RunsScored - log.RunsAllowed
		logs = append(logs, *log)
	}

	start := (pagination.Page - 1) * pagination.PerPage
	end := start + pagination.PerPage
	if start > len(logs) {
		start = len(logs)
	}
	if end > len(logs) {
		end = len(logs)
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    logs[start:end],
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		Total:   len(logs),
	})
}

// handleTeamDailyStats godoc
// @Summary Get team daily statistics
// @Description Get per-game team statistics for daily performance tracking and analysis
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param season query integer false "Filter by season year"
// @Param date_from query string false "Filter by start date (YYYYMMDD)"
// @Param date_to query string false "Filter by end date (YYYYMMDD)"
// @Param result query string false "Filter by result (W, L, or T)"
// @Param sort_by query string false "Sort by field (date, runs, runs_allowed)" default("date")
// @Param sort_order query string false "Sort order (asc, desc)" default("desc")
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/teams/{id}/daily-stats [get]
func (tr *TeamRoutes) handleTeamDailyStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamID := core.TeamID(r.PathValue("id"))
	filter := core.TeamDailyStatsFilter{
		TeamID:     &teamID,
		Pagination: *core.NewPagination(getIntQuery(r, "page", 1), getIntQuery(r, "per_page", 50)),
	}

	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		season := core.SeasonYear(getIntQuery(r, "season", 0))
		filter.Season = &season
	}

	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		filter.DateFrom = &dateFrom
	}

	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		filter.DateTo = &dateTo
	}

	if result := r.URL.Query().Get("result"); result != "" {
		filter.Result = &result
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		filter.SortOrder = core.SortOrder(sortOrder)
	}

	stats, err := tr.repo.DailyStats(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := tr.repo.CountDailyStats(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, NewPaginatedResponse(stats, filter.Pagination.Page, filter.Pagination.PerPage, total))
}
