package api

import (
	"net/http"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

type StandingsRoutes struct {
	repo core.StandingsRepository
}

func NewStandingsRoutes(repo core.StandingsRepository) *StandingsRoutes {
	return &StandingsRoutes{repo: repo}
}

func (sr *StandingsRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/standings", sr.handleSeasonStandings)
}

// handleSeasonStandings godoc
//
//	@Summary		Get season standings
//	@Description	Returns season standings. Current-season requests read from local current_season tables; historical seasons read from Lahman Teams records.
//	@Tags			stats, teams
//	@Accept			json
//	@Produce		json
//	@Param			season	query		integer	false	"Season year (defaults to current year)"
//	@Success		200		{object}	StandingsResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/standings [get]
func (sr *StandingsRoutes) handleSeasonStandings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	season := core.SeasonYear(getIntQuery(r, "season", time.Now().Year()))

	standings, lastUpdated, err := sr.repo.SeasonStandings(ctx, season)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, StandingsResponse{
		Season:      season,
		LastUpdated: lastUpdated,
		Standings:   standings,
	})
}
