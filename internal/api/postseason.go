package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type PostseasonRoutes struct {
	repo     core.PostseasonRepository
	gameRepo core.GameRepository
}

func NewPostseasonRoutes(repo core.PostseasonRepository, gameRepo core.GameRepository) *PostseasonRoutes {
	return &PostseasonRoutes{repo: repo, gameRepo: gameRepo}
}

func (pr *PostseasonRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/seasons/{year}/postseason/series", pr.handleSeasonPostseasonSeries)
	mux.HandleFunc("GET /v1/seasons/{year}/postseason/games", pr.handlePostseasonGames)
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
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, PostseasonSeriesResponse{
		Year:   core.SeasonYear(year),
		Series: series,
	})
}

// handlePostseasonGames godoc
// @Summary Get postseason games for a season
// @Description Get all postseason games from Retrosheet data for a specific year
// @Tags postseason, games
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(100)
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/postseason/games [get]
func (pr *PostseasonRoutes) handlePostseasonGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))

	isPostseason := true
	filter := core.GameFilter{
		Season:       &year,
		IsPostseason: &isPostseason,
		Pagination: core.Pagination{
			Page:    getIntQuery(r, "page", 1),
			PerPage: getIntQuery(r, "per_page", 100),
		},
	}

	games, err := pr.gameRepo.List(ctx, filter)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total, err := pr.gameRepo.Count(ctx, filter)
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

// PostseasonSeriesResponse wraps postseason series for a season
type PostseasonSeriesResponse struct {
	Year   core.SeasonYear         `json:"year"`
	Series []core.PostseasonSeries `json:"series"`
}
