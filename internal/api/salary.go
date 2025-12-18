package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type SalaryRoutes struct {
	repo core.SalaryRepository
}

func NewSalaryRoutes(repo core.SalaryRepository) *SalaryRoutes {
	return &SalaryRoutes{repo: repo}
}

func (sr *SalaryRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/salaries/summary", sr.handleListSalarySummary)
	mux.HandleFunc("GET /v1/salaries/summary/{year}", sr.handleGetSalarySummary)
}

// handleListSalarySummary godoc
// @Summary List all salary summaries
// @Description Get salary aggregates (total, average, median) for all years
// @Tags salaries
// @Accept json
// @Produce json
// @Success 200 {object} SalarySummaryResponse
// @Failure 500 {object} ErrorResponse
// @Router /salaries/summary [get]
func (sr *SalaryRoutes) handleListSalarySummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	summaries, err := sr.repo.List(ctx)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, NewSalarySummaryResponse(summaries))
}

// handleGetSalarySummary godoc
// @Summary Get salary summary for a specific year
// @Description Get salary aggregates (total, average, median) for a specific season
// @Tags salaries
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Success 200 {object} core.SalarySummary
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /salaries/summary/{year} [get]
func (sr *SalaryRoutes) handleGetSalarySummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	year := core.SeasonYear(getIntPathValue(r, "year"))
	summary, err := sr.repo.Get(ctx, year)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	if summary == nil {
		writeNotFound(w, "salary summary for year")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
