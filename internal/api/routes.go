package api

import (
	"net/http"
	"time"

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

type GameRoutes struct {
	repo     core.GameRepository
	playRepo core.PlayRepository
}

func NewGameRoutes(repo core.GameRepository, playRepo core.PlayRepository) *GameRoutes {
	return &GameRoutes{repo: repo, playRepo: playRepo}
}

func (gr *GameRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/games", gr.handleListGames)
	mux.HandleFunc("GET /v1/games/{id}", gr.handleGetGame)
	mux.HandleFunc("GET /v1/games/{id}/boxscore", gr.handleGetBoxscore)
	mux.HandleFunc("GET /v1/games/{id}/summary", gr.handleGetGameSummary)
	mux.HandleFunc("GET /v1/games/{id}/events", gr.handleGameEvents)
	mux.HandleFunc("GET /v1/games/{id}/events/{event_seq}", gr.handleSingleEvent)
	mux.HandleFunc("GET /v1/seasons/{year}/schedule", gr.handleSeasonSchedule)
	mux.HandleFunc("GET /v1/seasons/{year}/dates/{date}/games", gr.handleGamesByDate)
	mux.HandleFunc("GET /v1/seasons/{year}/teams/{team_id}/games", gr.handleTeamGames)
	mux.HandleFunc("GET /v1/seasons/{year}/parks/{park_id}/games", gr.handleParkGames)
}

// handleGetGame godoc
// @Summary Get game by ID
// @Description Get detailed information for a specific game
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID (format: YYYYMMDD + game_number + home_team)"
// @Success 200 {object} core.Game
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{id} [get]
func (gr *GameRoutes) handleGetGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.GameID(r.PathValue("id"))

	game, err := gr.repo.GetByID(ctx, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, game)
}

// handleGetBoxscore godoc
// @Summary Get game boxscore
// @Description Get detailed boxscore statistics for a specific game including team stats and lineups
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID (format: YYYYMMDD + game_number + home_team)"
// @Success 200 {object} core.Boxscore
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{id}/boxscore [get]
func (gr *GameRoutes) handleGetBoxscore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.GameID(r.PathValue("id"))

	boxscore, err := gr.repo.GetBoxscore(ctx, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, boxscore)
}

// handleListGames godoc
// @Summary List games
// @Description Search and browse games with optional filters and pagination
// @Tags games
// @Accept json
// @Produce json
// @Param season query integer false "Filter by season year"
// @Param home_team query string false "Filter by home team ID"
// @Param away_team query string false "Filter by away team ID"
// @Param park_id query string false "Filter by park ID"
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /games [get]
func (gr *GameRoutes) handleListGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	if homeTeam := r.URL.Query().Get("home_team"); homeTeam != "" {
		t := core.TeamID(homeTeam)
		filter.HomeTeam = &t
	}

	if awayTeam := r.URL.Query().Get("away_team"); awayTeam != "" {
		t := core.TeamID(awayTeam)
		filter.AwayTeam = &t
	}

	if parkID := r.URL.Query().Get("park_id"); parkID != "" {
		p := core.ParkID(parkID)
		filter.ParkID = &p
	}

	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if d, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = &d
		}
	}

	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if d, err := time.Parse("2006-01-02", dateTo); err == nil {
			filter.DateTo = &d
		}
	}

	games, err := gr.repo.List(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := gr.repo.Count(ctx, filter)
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

// handleSeasonSchedule godoc
// @Summary Get season schedule
// @Description Get all games for a specific season
// @Tags games
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(100)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/schedule [get]
func (gr *GameRoutes) handleSeasonSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))

	filter := core.GameFilter{
		Season: &year,
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 100),
		},
	}

	games, err := gr.repo.List(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := gr.repo.Count(ctx, filter)
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

// handleGamesByDate godoc
// @Summary Get games by date
// @Description Get all games played on a specific date
// @Tags games
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param date path string true "Date (YYYY-MM-DD format)"
// @Success 200 {array} core.Game
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/dates/{date}/games [get]
func (gr *GameRoutes) handleGamesByDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := getIntPathValue(r, "year")
	dateStr := r.PathValue("date")

	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeBadRequest(w, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	if targetDate.Year() != year {
		writeBadRequest(w, "Date year must match season year")
		return
	}

	games, err := gr.repo.ListByDate(ctx, targetDate)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, games)
}

// handleTeamGames godoc
// @Summary Get team games for a season
// @Description Get all games for a specific team in a season
// @Tags games
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param team_id path string true "Team ID"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(100)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/teams/{team_id}/games [get]
func (gr *GameRoutes) handleTeamGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	teamID := core.TeamID(r.PathValue("team_id"))

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 100),
	}

	games, err := gr.repo.ListByTeamSeason(ctx, teamID, year, pagination)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	filter := core.GameFilter{
		HomeTeam:   &teamID,
		Season:     &year,
		Pagination: pagination,
	}

	total, err := gr.repo.Count(ctx, filter)
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

// handleParkGames godoc
// @Summary Get games at a park
// @Description Get all games played at a specific ballpark in a season
// @Tags games, parks
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param park_id path string true "Park ID"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(100)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/parks/{park_id}/games [get]
func (gr *GameRoutes) handleParkGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	parkID := core.ParkID(r.PathValue("park_id"))

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 100),
	}

	filter := core.GameFilter{
		Season:     &year,
		ParkID:     &parkID,
		Pagination: pagination,
	}

	games, err := gr.repo.List(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := gr.repo.Count(ctx, filter)
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

// handleGetGameSummary godoc
// @Summary Get game summary
// @Description Get narrative summary for a game including winning pitcher, save, and key events
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID (format: YYYYMMDD + game_number + home_team)"
// @Success 200 {object} map[string]any
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{id}/summary [get]
func (gr *GameRoutes) handleGetGameSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.GameID(r.PathValue("id"))

	game, err := gr.repo.GetByID(ctx, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	boxscore, err := gr.repo.GetBoxscore(ctx, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	summary := map[string]any{
		"game_id":     game.ID,
		"date":        game.Date,
		"home_team":   game.HomeTeam,
		"away_team":   game.AwayTeam,
		"home_score":  game.HomeScore,
		"away_score":  game.AwayScore,
		"innings":     game.Innings,
		"winner":      determineWinner(game),
		"home_lineup": boxscore.HomeLineup,
		"away_lineup": boxscore.AwayLineup,
	}

	writeJSON(w, http.StatusOK, summary)
}

// handleGameEvents godoc
// @Summary Get game events
// @Description Get all play-by-play events for a game (alias for /games/{id}/plays)
// @Tags games, plays
// @Accept json
// @Produce json
// @Param id path string true "Game ID (format: YYYYMMDD + game_number + home_team)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(200)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{id}/events [get]
func (gr *GameRoutes) handleGameEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gameID := core.GameID(r.PathValue("id"))

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 200),
	}

	plays, err := gr.playRepo.ListByGame(ctx, gameID, pagination)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	filter := core.PlayFilter{
		GameID:     &gameID,
		Pagination: core.Pagination{Page: 1, PerPage: 1},
	}
	total, err := gr.playRepo.Count(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    plays,
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		Total:   total,
	})
}

// handleSingleEvent godoc
// @Summary Get single event
// @Description Get a single play/event by sequence number
// @Tags games, plays
// @Accept json
// @Produce json
// @Param id path string true "Game ID (format: YYYYMMDD + game_number + home_team)"
// @Param event_seq path integer true "Event sequence number (play number)"
// @Success 200 {object} core.Play
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{id}/events/{event_seq} [get]
func (gr *GameRoutes) handleSingleEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gameID := core.GameID(r.PathValue("id"))
	eventSeq := getIntPathValue(r, "event_seq")

	pagination := core.Pagination{
		Page:    1,
		PerPage: 1000,
	}

	plays, err := gr.playRepo.ListByGame(ctx, gameID, pagination)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	for _, play := range plays {
		if play.PlayNum == eventSeq {
			writeJSON(w, http.StatusOK, play)
			return
		}
	}

	writeNotFound(w, "Event")
}

func determineWinner(game *core.Game) core.TeamID {
	if game.HomeScore > game.AwayScore {
		return game.HomeTeam
	}
	return game.AwayTeam
}
