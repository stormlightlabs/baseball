package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type PostseasonRoutes struct {
	repo core.PostseasonRepository
}

func NewPostseasonRoutes(repo core.PostseasonRepository) *PostseasonRoutes {
	return &PostseasonRoutes{repo: repo}
}

func (pr *PostseasonRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/seasons/{year}/postseason/series", pr.handleSeasonPostseasonSeries)
}

// handleSeasonPostseasonSeries godoc
// @Summary Get postseason series for a season
// @Description Get all postseason series for a specific year
// @Tags postseason
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Success 200 {object} PostseasonSeriesResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/postseason/series [get]
func (pr *PostseasonRoutes) handleSeasonPostseasonSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	year := getIntPathValue(r, "year")

	series, err := pr.repo.ListSeries(ctx, core.SeasonYear(year))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PostseasonSeriesResponse{
		Year:   core.SeasonYear(year),
		Series: series,
	})
}

// PostseasonSeriesResponse wraps postseason series for a season
type PostseasonSeriesResponse struct {
	Year   core.SeasonYear         `json:"year"`
	Series []core.PostseasonSeries `json:"series"`
}
