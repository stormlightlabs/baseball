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
	mux.HandleFunc("GET /v1/players/{id}/stats/batting", pr.handlePlayerBattingStats)
	mux.HandleFunc("GET /v1/players/{id}/stats/pitching", pr.handlePlayerPitchingStats)
	mux.HandleFunc("GET /v1/players/{id}/awards", pr.handlePlayerAwards)
	mux.HandleFunc("GET /v1/players/{id}/hall-of-fame", pr.handlePlayerHallOfFame)
	mux.HandleFunc("GET /v1/players/{id}/game-logs", pr.handlePlayerGameLogs)
	mux.HandleFunc("GET /v1/players/{id}/appearances", pr.handlePlayerAppearances)
	mux.HandleFunc("GET /v1/players/{id}/teams", pr.handlePlayerTeams)
	mux.HandleFunc("GET /v1/players/{id}/salaries", pr.handlePlayerSalaries)
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
		writeInternalServerError(w, err)
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
		writeInternalServerError(w, err)
		return
	}

	total, err := pr.repo.Count(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
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
		writeInternalServerError(w, err)
		return
	}

	pitching, err := pr.repo.PitchingSeasons(ctx, id)
	if err != nil {
		writeInternalServerError(w, err)
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
		writeInternalServerError(w, err)
		return
	}

	total, err := pr.awardRepo.CountAwardResults(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
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
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, HallOfFameResponse{
		Records: records,
	})
}

// handlePlayerGameLogs godoc
// @Summary Get player game logs
// @Description Get list of games where the player appeared in the starting lineup
// @Tags players
// @Accept json
// @Produce json
// @Param id path string true "Player ID"
// @Param season query integer false "Filter by season"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{id}/game-logs [get]
func (pr *PlayerRoutes) handlePlayerGameLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.PlayerID(r.PathValue("id"))

	filter := core.GameFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if season := r.URL.Query().Get("season"); season != "" {
		y := core.SeasonYear(getIntQuery(r, "season", 0))
		filter.Season = &y
	}

	games, err := pr.repo.GameLogs(ctx, id, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    games,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   len(games),
	})
}

// handlePlayerAppearances godoc
// @Summary Get player appearances
// @Description Get appearance records by position for a player across all seasons
// @Tags players
// @Accept json
// @Produce json
// @Param id path string true "Player ID"
// @Success 200 {object} []core.PlayerAppearance
// @Failure 500 {object} ErrorResponse
// @Router /players/{id}/appearances [get]
func (pr *PlayerRoutes) handlePlayerAppearances(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.PlayerID(r.PathValue("id"))

	appearances, err := pr.repo.Appearances(ctx, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, appearances)
}

// handlePlayerTeams godoc
// @Summary Get player's team history
// @Description List every season/team combination a player appeared in
// @Tags players
// @Accept json
// @Produce json
// @Param id path string true "Player ID"
// @Success 200 {array} core.PlayerTeamSeason
// @Failure 500 {object} ErrorResponse
// @Router /players/{id}/teams [get]
func (pr *PlayerRoutes) handlePlayerTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.PlayerID(r.PathValue("id"))

	teams, err := pr.repo.Teams(ctx, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, teams)
}

// handlePlayerSalaries godoc
// @Summary Get player's salary history
// @Description Return all salary records for a player from Lahman Salaries table
// @Tags players
// @Accept json
// @Produce json
// @Param id path string true "Player ID"
// @Success 200 {array} core.PlayerSalary
// @Failure 500 {object} ErrorResponse
// @Router /players/{id}/salaries [get]
func (pr *PlayerRoutes) handlePlayerSalaries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.PlayerID(r.PathValue("id"))

	salaries, err := pr.repo.Salaries(ctx, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, salaries)
}

// handlePlayerBattingStats godoc
// @Summary Get player's batting statistics
// @Description Get comprehensive batting statistics for a player including career totals and season-by-season breakdowns
// @Tags players, stats
// @Accept json
// @Produce json
// @Param id path string true "Player ID"
// @Success 200 {object} PlayerBattingStatsResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{id}/stats/batting [get]
func (pr *PlayerRoutes) handlePlayerBattingStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.PlayerID(r.PathValue("id"))

	seasons, err := pr.repo.BattingSeasons(ctx, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	var career core.PlayerBattingSeason
	career.PlayerID = id
	career.TeamID = core.TeamID("TOTAL")
	career.League = core.LeagueID("")

	for _, season := range seasons {
		career.G += season.G
		career.PA += season.PA
		career.AB += season.AB
		career.R += season.R
		career.H += season.H
		career.Doubles += season.Doubles
		career.Triples += season.Triples
		career.HR += season.HR
		career.RBI += season.RBI
		career.SB += season.SB
		career.CS += season.CS
		career.BB += season.BB
		career.SO += season.SO
		career.HBP += season.HBP
		career.SF += season.SF
	}

	if career.AB > 0 {
		career.AVG = float64(career.H) / float64(career.AB)
	}

	if career.AB+career.BB+career.HBP+career.SF > 0 {
		career.OBP = float64(career.H+career.BB+career.HBP) / float64(career.AB+career.BB+career.HBP+career.SF)
	}

	if career.AB > 0 {
		career.SLG = float64(career.H+career.Doubles+(2*career.Triples)+(3*career.HR)) / float64(career.AB)
	}

	career.OPS = career.OBP + career.SLG

	writeJSON(w, http.StatusOK, PlayerBattingStatsResponse{
		Career:  career,
		Seasons: seasons,
	})
}

// handlePlayerPitchingStats godoc
// @Summary Get player's pitching statistics
// @Description Get comprehensive pitching statistics for a player including career totals and season-by-season breakdowns
// @Tags players, stats
// @Accept json
// @Produce json
// @Param id path string true "Player ID"
// @Success 200 {object} PlayerPitchingStatsResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{id}/stats/pitching [get]
func (pr *PlayerRoutes) handlePlayerPitchingStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.PlayerID(r.PathValue("id"))

	seasons, err := pr.repo.PitchingSeasons(ctx, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	var career core.PlayerPitchingSeason
	career.PlayerID = id
	career.TeamID = core.TeamID("TOTAL")
	career.League = core.LeagueID("")

	for _, season := range seasons {
		career.W += season.W
		career.L += season.L
		career.G += season.G
		career.GS += season.GS
		career.CG += season.CG
		career.SHO += season.SHO
		career.SV += season.SV
		career.IPOuts += season.IPOuts
		career.H += season.H
		career.ER += season.ER
		career.HR += season.HR
		career.BB += season.BB
		career.SO += season.SO
		career.WP += season.WP
		career.HBP += season.HBP
		career.BK += season.BK
	}

	if career.IPOuts > 0 {
		ip := float64(career.IPOuts) / 3.0
		career.ERA = (float64(career.ER) * 9.0) / ip
		career.WHIP = float64(career.BB+career.H) / ip
		career.KPer9 = (float64(career.SO) * 9.0) / ip
		career.BBPer9 = (float64(career.BB) * 9.0) / ip
		career.HRPer9 = (float64(career.HR) * 9.0) / ip
	}

	writeJSON(w, http.StatusOK, PlayerPitchingStatsResponse{
		Career:  career,
		Seasons: seasons,
	})
}
