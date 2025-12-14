package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type StatsRoutes struct {
	repo core.StatsRepository
}

func NewStatsRoutes(repo core.StatsRepository) *StatsRoutes {
	return &StatsRoutes{repo: repo}
}

func (sr *StatsRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/seasons/{year}/leaders/batting", sr.handleBattingLeaders)
	mux.HandleFunc("GET /v1/seasons/{year}/leaders/pitching", sr.handlePitchingLeaders)
	mux.HandleFunc("GET /v1/leaders/batting/career", sr.handleCareerBattingLeaders)
	mux.HandleFunc("GET /v1/leaders/pitching/career", sr.handleCareerPitchingLeaders)
	mux.HandleFunc("GET /v1/stats/batting", sr.handleQueryBattingStats)
	mux.HandleFunc("GET /v1/stats/pitching", sr.handleQueryPitchingStats)
	mux.HandleFunc("GET /v1/stats/fielding", sr.handleQueryFieldingStats)
	mux.HandleFunc("GET /v1/stats/teams/batting", sr.handleTeamBattingStats)
	mux.HandleFunc("GET /v1/stats/teams/pitching", sr.handleTeamPitchingStats)
	mux.HandleFunc("GET /v1/stats/teams/fielding", sr.handleTeamFieldingStats)
}

// handleBattingLeaders godoc
// @Summary Get batting leaders
// @Description Get season batting leaders for a specific statistic
// @Tags stats
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param stat query string false "Statistic (hr, avg, rbi, sb, h, r)" default("hr")
// @Param league query string false "Filter by league (AL, NL)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(10)
// @Success 200 {object} BattingLeadersResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/leaders/batting [get]
func (sr *StatsRoutes) handleBattingLeaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	year := core.SeasonYear(getIntPathValue(r, "year"))
	stat := r.URL.Query().Get("stat")
	if stat == "" {
		stat = "hr"
	}

	page := getIntQuery(r, "page", 1)
	perPage := getIntQuery(r, "per_page", 10)
	limit := perPage
	offset := (page - 1) * perPage

	var league *core.LeagueID
	if lg := r.URL.Query().Get("league"); lg != "" {
		lgID := core.LeagueID(lg)
		league = &lgID
	}

	leaders, err := sr.repo.SeasonBattingLeaders(ctx, year, stat, limit, offset, league)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	allLeaders, err := sr.repo.SeasonBattingLeaders(ctx, year, stat, 10000, 0, league)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	total := len(allLeaders)

	writeJSON(w, http.StatusOK, BattingLeadersResponse{
		Year:    year,
		Stat:    stat,
		League:  league,
		Leaders: leaders,
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// handlePitchingLeaders godoc
// @Summary Get pitching leaders
// @Description Get season pitching leaders for a specific statistic
// @Tags stats
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param stat query string false "Statistic (era, so, w, sv, ip)" default("era")
// @Param league query string false "Filter by league (AL, NL)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(10)
// @Success 200 {object} PitchingLeadersResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/leaders/pitching [get]
func (sr *StatsRoutes) handlePitchingLeaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	year := core.SeasonYear(getIntPathValue(r, "year"))
	stat := r.URL.Query().Get("stat")
	if stat == "" {
		stat = "era"
	}

	page := getIntQuery(r, "page", 1)
	perPage := getIntQuery(r, "per_page", 10)
	limit := perPage
	offset := (page - 1) * perPage

	var league *core.LeagueID
	if lg := r.URL.Query().Get("league"); lg != "" {
		lgID := core.LeagueID(lg)
		league = &lgID
	}

	leaders, err := sr.repo.SeasonPitchingLeaders(ctx, year, stat, limit, offset, league)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	allLeaders, err := sr.repo.SeasonPitchingLeaders(ctx, year, stat, 10000, 0, league)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	total := len(allLeaders)

	writeJSON(w, http.StatusOK, PitchingLeadersResponse{
		Year:    year,
		Stat:    stat,
		League:  league,
		Leaders: leaders,
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// handleQueryBattingStats godoc
// @Summary Query batting statistics
// @Description Flexible batting stats query with multiple filter options
// @Tags stats
// @Accept json
// @Produce json
// @Param player_id query string false "Filter by player ID"
// @Param team_id query string false "Filter by team ID"
// @Param season query integer false "Filter by specific season"
// @Param season_from query integer false "Filter by season range (start)"
// @Param season_to query integer false "Filter by season range (end)"
// @Param league query string false "Filter by league (AL, NL)"
// @Param min_ab query integer false "Minimum at-bats threshold" default(0)
// @Param sort_by query string false "Sort by stat (avg, hr, rbi, sb, h, r)" default("h")
// @Param sort_order query string false "Sort order (asc, desc)" default("desc")
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /stats/batting [get]
func (sr *StatsRoutes) handleQueryBattingStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.BattingStatsFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
		SortBy: r.URL.Query().Get("sort_by"),
	}

	if playerID := r.URL.Query().Get("player_id"); playerID != "" {
		pid := core.PlayerID(playerID)
		filter.PlayerID = &pid
	}

	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		tid := core.TeamID(teamID)
		filter.TeamID = &tid
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

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	if minAB := r.URL.Query().Get("min_ab"); minAB != "" {
		ab := getIntQuery(r, "min_ab", 0)
		filter.MinAB = &ab
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder == "asc" {
		filter.SortOrder = core.SortAsc
	} else {
		filter.SortOrder = core.SortDesc
	}

	stats, err := sr.repo.QueryBattingStats(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := sr.repo.QueryBattingStatsCount(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    stats,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleQueryPitchingStats godoc
// @Summary Query pitching statistics
// @Description Flexible pitching stats query with multiple filter options
// @Tags stats
// @Accept json
// @Produce json
// @Param player_id query string false "Filter by player ID"
// @Param team_id query string false "Filter by team ID"
// @Param season query integer false "Filter by specific season"
// @Param season_from query integer false "Filter by season range (start)"
// @Param season_to query integer false "Filter by season range (end)"
// @Param league query string false "Filter by league (AL, NL)"
// @Param min_ip query number false "Minimum innings pitched threshold" default(0)
// @Param min_gs query integer false "Minimum games started threshold" default(0)
// @Param sort_by query string false "Sort by stat (era, w, so, sv, ip)" default("so")
// @Param sort_order query string false "Sort order (asc, desc)" default("desc")
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /stats/pitching [get]
func (sr *StatsRoutes) handleQueryPitchingStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.PitchingStatsFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
		SortBy: r.URL.Query().Get("sort_by"),
	}

	if playerID := r.URL.Query().Get("player_id"); playerID != "" {
		pid := core.PlayerID(playerID)
		filter.PlayerID = &pid
	}

	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		tid := core.TeamID(teamID)
		filter.TeamID = &tid
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

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	if minIP := r.URL.Query().Get("min_ip"); minIP != "" {
		ip := float64(getIntQuery(r, "min_ip", 0))
		filter.MinIP = &ip
	}

	if minGS := r.URL.Query().Get("min_gs"); minGS != "" {
		gs := getIntQuery(r, "min_gs", 0)
		filter.MinGS = &gs
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder == "asc" {
		filter.SortOrder = core.SortAsc
	} else {
		filter.SortOrder = core.SortDesc
	}

	stats, err := sr.repo.QueryPitchingStats(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := sr.repo.QueryPitchingStatsCount(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    stats,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleQueryFieldingStats godoc
// @Summary Query fielding statistics
// @Description Flexible fielding stats query with multiple filter options
// @Tags stats
// @Accept json
// @Produce json
// @Param player_id query string false "Filter by player ID"
// @Param team_id query string false "Filter by team ID"
// @Param season query integer false "Filter by specific season"
// @Param season_from query integer false "Filter by season range (start)"
// @Param season_to query integer false "Filter by season range (end)"
// @Param league query string false "Filter by league (AL, NL)"
// @Param position query string false "Filter by position (1B, 2B, 3B, SS, OF, C, P, DH)"
// @Param min_g query integer false "Minimum games threshold" default(0)
// @Param sort_by query string false "Sort by stat (po, a, e, dp, fpct)" default("po")
// @Param sort_order query string false "Sort order (asc, desc)" default("desc")
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /stats/fielding [get]
func (sr *StatsRoutes) handleQueryFieldingStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.FieldingStatsFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
		SortBy: r.URL.Query().Get("sort_by"),
	}

	if playerID := r.URL.Query().Get("player_id"); playerID != "" {
		pid := core.PlayerID(playerID)
		filter.PlayerID = &pid
	}

	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		tid := core.TeamID(teamID)
		filter.TeamID = &tid
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

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	if position := r.URL.Query().Get("position"); position != "" {
		filter.Position = &position
	}

	if minG := r.URL.Query().Get("min_g"); minG != "" {
		g := getIntQuery(r, "min_g", 0)
		filter.MinG = &g
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder == "asc" {
		filter.SortOrder = core.SortAsc
	} else {
		filter.SortOrder = core.SortDesc
	}

	stats, err := sr.repo.QueryFieldingStats(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := sr.repo.QueryFieldingStatsCount(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    stats,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleCareerBattingLeaders godoc
// @Summary Get career batting leaders
// @Description Get all-time career batting leaders for a specific statistic
// @Tags stats
// @Accept json
// @Produce json
// @Param stat query string false "Statistic (hr, avg, rbi, sb, h, r, ops)" default("hr")
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(10)
// @Success 200 {object} CareerBattingLeadersResponse
// @Failure 500 {object} ErrorResponse
// @Router /leaders/batting/career [get]
func (sr *StatsRoutes) handleCareerBattingLeaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stat := r.URL.Query().Get("stat")
	if stat == "" {
		stat = "hr"
	}

	page := getIntQuery(r, "page", 1)
	perPage := getIntQuery(r, "per_page", 10)
	limit := perPage
	offset := (page - 1) * perPage

	leaders, err := sr.repo.CareerBattingLeaders(ctx, stat, limit, offset)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	allLeaders, err := sr.repo.CareerBattingLeaders(ctx, stat, 10000, 0)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	total := len(allLeaders)

	writeJSON(w, http.StatusOK, CareerBattingLeadersResponse{
		Stat:    stat,
		Leaders: leaders,
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// handleCareerPitchingLeaders godoc
// @Summary Get career pitching leaders
// @Description Get all-time career pitching leaders for a specific statistic
// @Tags stats
// @Accept json
// @Produce json
// @Param stat query string false "Statistic (era, so, w, sv, ip)" default("w")
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(10)
// @Success 200 {object} CareerPitchingLeadersResponse
// @Failure 500 {object} ErrorResponse
// @Router /leaders/pitching/career [get]
func (sr *StatsRoutes) handleCareerPitchingLeaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stat := r.URL.Query().Get("stat")
	if stat == "" {
		stat = "w"
	}

	page := getIntQuery(r, "page", 1)
	perPage := getIntQuery(r, "per_page", 10)
	limit := perPage
	offset := (page - 1) * perPage

	leaders, err := sr.repo.CareerPitchingLeaders(ctx, stat, limit, offset)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	allLeaders, err := sr.repo.CareerPitchingLeaders(ctx, stat, 10000, 0)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	total := len(allLeaders)

	writeJSON(w, http.StatusOK, CareerPitchingLeadersResponse{
		Stat:    stat,
		Leaders: leaders,
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// handleTeamBattingStats godoc
// @Summary Query team batting statistics
// @Description Flexible team batting stats query with filters for team, season range, and league
// @Tags stats
// @Accept json
// @Produce json
// @Param team_id query string false "Filter by team ID"
// @Param season query integer false "Filter by specific season"
// @Param season_from query integer false "Filter by season range (start)"
// @Param season_to query integer false "Filter by season range (end)"
// @Param league query string false "Filter by league (AL, NL)"
// @Param sort_by query string false "Sort by stat (hr, avg, obp, slg, ops, h, r, rbi, sb)" default("hr")
// @Param sort_order query string false "Sort order (asc, desc)" default("desc")
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /stats/teams/batting [get]
func (sr *StatsRoutes) handleTeamBattingStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.TeamStatsFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
		SortBy: r.URL.Query().Get("sort_by"),
	}

	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		tid := core.TeamID(teamID)
		filter.TeamID = &tid
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

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder == "asc" {
		filter.SortOrder = core.SortAsc
	} else {
		filter.SortOrder = core.SortDesc
	}

	stats, err := sr.repo.TeamBattingStats(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := sr.repo.TeamBattingStatsCount(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    stats,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleTeamPitchingStats godoc
// @Summary Query team pitching statistics
// @Description Flexible team pitching stats query with filters for team, season range, and league
// @Tags stats
// @Accept json
// @Produce json
// @Param team_id query string false "Filter by team ID"
// @Param season query integer false "Filter by specific season"
// @Param season_from query integer false "Filter by season range (start)"
// @Param season_to query integer false "Filter by season range (end)"
// @Param league query string false "Filter by league (AL, NL)"
// @Param sort_by query string false "Sort by stat (era, whip, w, l, sv, so, ip, cg, sho)" default("era")
// @Param sort_order query string false "Sort order (asc, desc)" default("desc")
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /stats/teams/pitching [get]
func (sr *StatsRoutes) handleTeamPitchingStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.TeamStatsFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
		SortBy: r.URL.Query().Get("sort_by"),
	}

	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		tid := core.TeamID(teamID)
		filter.TeamID = &tid
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

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder == "asc" {
		filter.SortOrder = core.SortAsc
	} else {
		filter.SortOrder = core.SortDesc
	}

	stats, err := sr.repo.TeamPitchingStats(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := sr.repo.TeamPitchingStatsCount(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    stats,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleTeamFieldingStats godoc
// @Summary Query team fielding statistics
// @Description Flexible team fielding stats query with filters for team, season range, and league
// @Tags stats
// @Accept json
// @Produce json
// @Param team_id query string false "Filter by team ID"
// @Param season query integer false "Filter by specific season"
// @Param season_from query integer false "Filter by season range (start)"
// @Param season_to query integer false "Filter by season range (end)"
// @Param league query string false "Filter by league (AL, NL)"
// @Param sort_by query string false "Sort by stat (po, a, e, dp, pb, fpct)" default("po")
// @Param sort_order query string false "Sort order (asc, desc)" default("desc")
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /stats/teams/fielding [get]
func (sr *StatsRoutes) handleTeamFieldingStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.TeamStatsFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
		SortBy: r.URL.Query().Get("sort_by"),
	}

	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		tid := core.TeamID(teamID)
		filter.TeamID = &tid
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

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder == "asc" {
		filter.SortOrder = core.SortAsc
	} else {
		filter.SortOrder = core.SortDesc
	}

	stats, err := sr.repo.TeamFieldingStats(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := sr.repo.TeamFieldingStatsCount(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    stats,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}
