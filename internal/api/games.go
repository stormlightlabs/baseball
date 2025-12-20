package api

import (
	"net/http"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

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
		writeError(w, err)
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
		writeError(w, err)
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
		writeError(w, err)
		return
	}

	total, err := gr.repo.Count(ctx, filter)
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
		writeError(w, err)
		return
	}

	total, err := gr.repo.Count(ctx, filter)
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
		writeError(w, err)
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
		writeError(w, err)
		return
	}

	filter := core.GameFilter{
		HomeTeam:   &teamID,
		Season:     &year,
		Pagination: pagination,
	}

	total, err := gr.repo.Count(ctx, filter)
	if err != nil {
		writeError(w, err)
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
		writeError(w, err)
		return
	}

	total, err := gr.repo.Count(ctx, filter)
	if err != nil {
		writeError(w, err)
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
		writeError(w, err)
		return
	}

	boxscore, err := gr.repo.GetBoxscore(ctx, id)
	if err != nil {
		writeError(w, err)
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
		writeError(w, err)
		return
	}

	filter := core.PlayFilter{
		GameID:     &gameID,
		Pagination: core.Pagination{Page: 1, PerPage: 1},
	}
	total, err := gr.playRepo.Count(ctx, filter)
	if err != nil {
		writeError(w, err)
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
		writeError(w, err)
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
