package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type StatsRoutes struct {
	repo core.StatsRepository
}

func NewStatsRoutes(repo core.StatsRepository) *StatsRoutes {
	return &StatsRoutes{repo: repo}
}

func (sr *StatsRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/seasons/{year}/leaders/batting", sr.handleBattingLeaders)
	mux.HandleFunc("GET /v1/seasons/{year}/leaders/pitching", sr.handlePitchingLeaders)
}

// handleBattingLeaders godoc
// @Summary Get batting leaders
// @Description Get season batting leaders for a specific statistic
// @Tags stats
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param stat query string false "Statistic (hr, avg, rbi, sb, h, r)" default("hr")
// @Param league query string false "Filter by league (AL, NL)"
// @Param limit query integer false "Number of results" default(10)
// @Param offset query integer false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/leaders/batting [get]
func (sr *StatsRoutes) handleBattingLeaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	year := core.SeasonYear(getIntQuery(r, "year", 2024))
	stat := r.URL.Query().Get("stat")
	if stat == "" {
		stat = "hr"
	}

	limit := getIntQuery(r, "limit", 10)
	offset := getIntQuery(r, "offset", 0)

	var league *core.LeagueID
	if lg := r.URL.Query().Get("league"); lg != "" {
		lgID := core.LeagueID(lg)
		league = &lgID
	}

	leaders, err := sr.repo.SeasonBattingLeaders(ctx, year, stat, limit, offset, league)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"year":    year,
		"stat":    stat,
		"league":  league,
		"leaders": leaders,
	})
}

// handlePitchingLeaders godoc
// @Summary Get pitching leaders
// @Description Get season pitching leaders for a specific statistic
// @Tags stats
// @Accept json
// @Produce json
// @Param year path integer true "Season year"
// @Param stat query string false "Statistic (era, so, w, sv, ip)" default("era")
// @Param league query string false "Filter by league (AL, NL)"
// @Param limit query integer false "Number of results" default(10)
// @Param offset query integer false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /seasons/{year}/leaders/pitching [get]
func (sr *StatsRoutes) handlePitchingLeaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	year := core.SeasonYear(getIntQuery(r, "year", 2024))
	stat := r.URL.Query().Get("stat")
	if stat == "" {
		stat = "era"
	}

	limit := getIntQuery(r, "limit", 10)
	offset := getIntQuery(r, "offset", 0)

	var league *core.LeagueID
	if lg := r.URL.Query().Get("league"); lg != "" {
		lgID := core.LeagueID(lg)
		league = &lgID
	}

	leaders, err := sr.repo.SeasonPitchingLeaders(ctx, year, stat, limit, offset, league)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"year":    year,
		"stat":    stat,
		"league":  league,
		"leaders": leaders,
	})
}
