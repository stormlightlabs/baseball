package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type AllStarRoutes struct {
	awardRepo core.AwardRepository
}

func NewAllStarRoutes(awardRepo core.AwardRepository) *AllStarRoutes {
	return &AllStarRoutes{awardRepo: awardRepo}
}

func (ar *AllStarRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/allstar/games", ar.handleListAllStarGames)
	mux.HandleFunc("GET /v1/allstar/games/{id}", ar.handleGetAllStarGame)
}

// handleListAllStarGames godoc
// @Summary List All-Star games
// @Description Get All-Star Game history by joining Lahman participation data with Retrosheet game logs
// @Tags allstar, games
// @Accept json
// @Produce json
// @Param year query integer false "Filter by season year"
// @Success 200 {array} core.AllStarGame
// @Failure 500 {object} ErrorResponse
// @Router /allstar/games [get]
func (ar *AllStarRoutes) handleListAllStarGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var year *core.SeasonYear
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		y := core.SeasonYear(getIntQuery(r, "year", 0))
		year = &y
	}

	games, err := ar.awardRepo.ListAllStarGames(ctx, year)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, games)
}

// handleGetAllStarGame godoc
// @Summary Get All-Star game details
// @Description Get detailed information and participant list for a specific All-Star game sourced from Retrosheet game logs
// @Tags allstar, games
// @Accept json
// @Produce json
// @Param id path string true "Game ID (e.g., ALS202407160)"
// @Success 200 {object} core.AllStarGame
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /allstar/games/{id} [get]
func (ar *AllStarRoutes) handleGetAllStarGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	game, err := ar.awardRepo.GetAllStarGame(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, game)
}
