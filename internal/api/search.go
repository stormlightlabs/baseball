package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type SearchRoutes struct {
	playerRepo core.PlayerRepository
	teamRepo   core.TeamRepository
	parkRepo   core.ParkRepository
	gameRepo   core.GameRepository
}

func NewSearchRoutes(playerRepo core.PlayerRepository, teamRepo core.TeamRepository, parkRepo core.ParkRepository, gameRepo core.GameRepository) *SearchRoutes {
	return &SearchRoutes{
		playerRepo: playerRepo,
		teamRepo:   teamRepo,
		parkRepo:   parkRepo,
		gameRepo:   gameRepo,
	}
}

func (sr *SearchRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/search/players", sr.handleSearchPlayers)
	mux.HandleFunc("GET /v1/search/teams", sr.handleSearchTeams)
	mux.HandleFunc("GET /v1/search/parks", sr.handleSearchParks)
	mux.HandleFunc("GET /v1/search/games", sr.handleSearchGames)
}

// handleSearchPlayers godoc
// @Summary Search players
// @Description Fuzzy player search with filters for name, position, era, and more
// @Tags search, players
// @Accept json
// @Produce json
// @Param q query string false "Search query (searches first and last name)"
// @Param debut_year query integer false "Filter by debut year"
// @Param position query string false "Filter by position"
// @Param bats query string false "Filter by batting hand (R, L, B)"
// @Param throws query string false "Filter by throwing hand (R, L)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /search/players [get]
func (sr *SearchRoutes) handleSearchPlayers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.PlayerFilter{
		NameQuery: r.URL.Query().Get("q"),
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if debutYear := r.URL.Query().Get("debut_year"); debutYear != "" {
		year := core.SeasonYear(getIntQuery(r, "debut_year", 0))
		filter.DebutYear = &year
	}

	players, err := sr.playerRepo.List(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := sr.playerRepo.Count(ctx, filter)
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

// handleSearchTeams godoc
// @Summary Search teams
// @Description Search teams by name, city, or franchise
// @Tags search, teams
// @Accept json
// @Produce json
// @Param q query string false "Search query (searches team name, team ID, and franchise ID)"
// @Param year query integer false "Filter by season year"
// @Param league query string false "Filter by league (AL, NL)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /search/teams [get]
func (sr *SearchRoutes) handleSearchTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.TeamFilter{
		NameQuery: r.URL.Query().Get("q"),
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if year := r.URL.Query().Get("year"); year != "" {
		y := core.SeasonYear(getIntQuery(r, "year", 0))
		filter.Year = &y
	}

	if league := r.URL.Query().Get("league"); league != "" {
		lg := core.LeagueID(league)
		filter.League = &lg
	}

	teams, err := sr.teamRepo.ListTeamSeasons(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := sr.teamRepo.CountTeamSeasons(ctx, filter)
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

// handleSearchParks godoc
// @Summary Search parks
// @Description Ballpark lookup by name, city, state, or park ID
// @Tags search, parks
// @Accept json
// @Produce json
// @Param q query string false "Search query (searches park name, city, state, and park ID)"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /search/parks [get]
func (sr *SearchRoutes) handleSearchParks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.ParkFilter{
		NameQuery: r.URL.Query().Get("q"),
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	parks, err := sr.parkRepo.List(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	// TODO: count method for parks
	total := len(parks)
	if len(parks) == filter.Pagination.PerPage {
		total = filter.Pagination.PerPage * filter.Pagination.Page
	}

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    parks,
		Page:    filter.Pagination.Page,
		PerPage: filter.Pagination.PerPage,
		Total:   total,
	})
}

// handleSearchGames godoc
// @Summary Search games with natural language
// @Description Natural language game search supporting queries like "yankees red sox 2004 alcs game 7" or "dodgers giants 2014 nlcs"
// @Tags search, games
// @Accept json
// @Produce json
// @Param q query string true "Natural language search query"
// @Param limit query integer false "Maximum number of results" default(50)
// @Success 200 {array} core.Game
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /search/games [get]
func (sr *SearchRoutes) handleSearchGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query().Get("q")
	if query == "" {
		writeBadRequest(w, "query parameter 'q' is required")
		return
	}

	limit := getIntQuery(r, "limit", 50)
	if limit > 200 {
		limit = 200
	}

	games, err := sr.gameRepo.SearchGamesNL(ctx, query, limit)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, games)
}
