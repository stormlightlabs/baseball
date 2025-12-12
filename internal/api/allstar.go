package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type AllStarRoutes struct {
	gameRepo core.GameRepository
}

func NewAllStarRoutes(gameRepo core.GameRepository) *AllStarRoutes {
	return &AllStarRoutes{gameRepo: gameRepo}
}

func (ar *AllStarRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/allstar/games", ar.handleListAllStarGames)
	mux.HandleFunc("GET /v1/allstar/games/{id}", ar.handleGetAllStarGame)
}

// handleListAllStarGames godoc
// @Summary List All-Star games
// @Description Get All-Star Game history from Retrosheet data
// @Tags allstar, games
// @Accept json
// @Produce json
// @Param year query integer false "Filter by season year"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /allstar/games [get]
func (ar *AllStarRoutes) handleListAllStarGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := core.GameFilter{
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 50),
		},
	}

	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		year := core.SeasonYear(getIntQuery(r, "year", 0))
		filter.Season = &year
	}

	games, err := ar.gameRepo.List(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	allStarGames := filterAllStarGames(games)

	total := len(allStarGames)
	page := filter.Pagination.Page
	perPage := filter.Pagination.PerPage

	start := (page - 1) * perPage
	end := start + perPage
	if start > len(allStarGames) {
		start = len(allStarGames)
	}
	if end > len(allStarGames) {
		end = len(allStarGames)
	}

	paginatedGames := allStarGames[start:end]

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    paginatedGames,
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// handleGetAllStarGame godoc
// @Summary Get All-Star game details
// @Description Get detailed information and boxscore for a specific All-Star game
// @Tags allstar, games
// @Accept json
// @Produce json
// @Param id path string true "Game ID (format: YYYYMMDD + game_number + home_team)"
// @Success 200 {object} AllStarGameResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /allstar/games/{id} [get]
func (ar *AllStarRoutes) handleGetAllStarGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := core.GameID(r.PathValue("id"))

	game, err := ar.gameRepo.GetByID(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}

	boxscore, err := ar.gameRepo.GetBoxscore(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, AllStarGameResponse{
		Game:     *game,
		Boxscore: *boxscore,
	})
}

// filterAllStarGames identifies All-Star games from a list of games.
// All-Star games typically have game_number = 0 and are played in July.
func filterAllStarGames(games []core.Game) []core.Game {
	var allStarGames []core.Game
	for _, game := range games {
		if game.Date.Month() == 7 && string(game.ID)[8] == '0' {
			allStarGames = append(allStarGames, game)
		}
	}
	return allStarGames
}

// AllStarGameResponse wraps All-Star game details with boxscore
type AllStarGameResponse struct {
	Game     core.Game     `json:"game"`
	Boxscore core.Boxscore `json:"boxscore"`
}
