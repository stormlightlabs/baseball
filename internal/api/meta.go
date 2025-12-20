package api

import (
	"net/http"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

const apiVersion = "1.0.0"

type MetaRoutes struct {
	repo core.MetaRepository
}

func NewMetaRoutes(repo core.MetaRepository) *MetaRoutes {
	return &MetaRoutes{repo: repo}
}

func (mr *MetaRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/meta", mr.handleMeta)
	mux.HandleFunc("GET /v1/meta/datasets", mr.handleDatasetStatus)
	mux.HandleFunc("GET /v1/meta/constants/woba", mr.handleWOBAConstants)
	mux.HandleFunc("GET /v1/meta/constants/league", mr.handleLeagueConstants)
	mux.HandleFunc("GET /v1/meta/constants/park-factors", mr.handleParkFactors)
}

type metaResponse struct {
	Version      string                     `json:"version"`
	GeneratedAt  time.Time                  `json:"generated_at"`
	Coverage     map[string]datasetCoverage `json:"coverage"`
	SchemaHashes map[string]string          `json:"schema_hashes"`
	Datasets     []core.DatasetStatus       `json:"datasets"`
}

type datasetCoverage struct {
	From *core.SeasonYear `json:"from,omitempty"`
	To   *core.SeasonYear `json:"to,omitempty"`
}

// handleMeta godoc
// @Summary API metadata
// @Description Returns API version, dataset freshness, coverage, and schema fingerprints
// @Tags meta
// @Accept json
// @Produce json
// @Success 200 {object} metaResponse
// @Failure 500 {object} ErrorResponse
// @Router /meta [get]
func (mr *MetaRoutes) handleMeta(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	datasets, err := mr.repo.DatasetStatuses(ctx)
	if err != nil {
		writeError(w, err)
		return
	}

	minLahman, maxLahman, minRetro, maxRetro, err := mr.repo.SeasonCoverage(ctx)
	if err != nil {
		writeError(w, err)
		return
	}

	schemaHashes, err := mr.repo.SchemaHashes(ctx)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := metaResponse{
		Version:     apiVersion,
		GeneratedAt: time.Now().UTC(),
		Coverage: map[string]datasetCoverage{
			"lahman":     makeCoverage(minLahman, maxLahman),
			"retrosheet": makeCoverage(minRetro, maxRetro),
		},
		SchemaHashes: schemaHashes,
		Datasets:     datasets,
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleDatasetStatus godoc
// @Summary Dataset status
// @Description Returns dataset ETL metadata and coverage
// @Tags meta
// @Accept json
// @Produce json
// @Success 200 {array} core.DatasetStatus
// @Failure 500 {object} ErrorResponse
// @Router /meta/datasets [get]
func (mr *MetaRoutes) handleDatasetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	datasets, err := mr.repo.DatasetStatuses(ctx)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, datasets)
}

func makeCoverage(from, to core.SeasonYear) datasetCoverage {
	c := datasetCoverage{}
	if from != 0 {
		f := from
		c.From = &f
	}
	if to != 0 {
		t := to
		c.To = &t
	}
	return c
}

// handleWOBAConstants godoc
// @Summary wOBA constants
// @Description Returns season-specific wOBA calculation constants from FanGraphs
// @Tags meta
// @Accept json
// @Produce json
// @Param season query int false "Season year (returns all if omitted)"
// @Success 200 {array} core.WOBAConstant
// @Failure 500 {object} ErrorResponse
// @Router /meta/constants/woba [get]
func (mr *MetaRoutes) handleWOBAConstants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var season *core.SeasonYear
	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		s := core.SeasonYear(getIntQuery(r, "season", 0))
		if s > 0 {
			season = &s
		}
	}

	constants, err := mr.repo.WOBAConstants(ctx, season)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, constants)
}

// handleLeagueConstants godoc
// @Summary League constants
// @Description Returns league-specific constants for wRC+ and WAR calculations
// @Tags meta
// @Accept json
// @Produce json
// @Param season query int false "Season year (returns all if omitted)"
// @Param league query string false "League (AL or NL)"
// @Success 200 {array} core.LeagueConstant
// @Failure 500 {object} ErrorResponse
// @Router /meta/constants/league [get]
func (mr *MetaRoutes) handleLeagueConstants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var season *core.SeasonYear
	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		s := core.SeasonYear(getIntQuery(r, "season", 0))
		if s > 0 {
			season = &s
		}
	}

	var league *core.LeagueID
	if leagueStr := r.URL.Query().Get("league"); leagueStr != "" {
		l := core.LeagueID(leagueStr)
		league = &l
	}

	constants, err := mr.repo.LeagueConstants(ctx, season, league)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, constants)
}

// handleParkFactors godoc
// @Summary Park factors
// @Description Returns FanGraphs park factors for seasons
// @Tags meta
// @Accept json
// @Produce json
// @Param season query int false "Season year (returns all if omitted)"
// @Param team query string false "Team ID filter"
// @Success 200 {array} core.ParkFactorRow
// @Failure 500 {object} ErrorResponse
// @Router /meta/constants/park-factors [get]
func (mr *MetaRoutes) handleParkFactors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var season *core.SeasonYear
	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		s := core.SeasonYear(getIntQuery(r, "season", 0))
		if s > 0 {
			season = &s
		}
	}

	var teamID *core.TeamID
	if teamStr := r.URL.Query().Get("team"); teamStr != "" {
		t := core.TeamID(teamStr)
		teamID = &t
	}

	factors, err := mr.repo.ParkFactors(ctx, season, teamID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, factors)
}
