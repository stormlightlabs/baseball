package api

import (
	"context"
	"net/http"
	"time"

	"stormlightlabs.org/baseball/internal/core"
	"stormlightlabs.org/baseball/internal/seed"
)

// TODO: we'll update to semantic versioning when ready
const apiVersion = "ALPHA" // 1.0.0
const countModeHeader = "X-Count-Mode"

type MetaRoutes struct {
	repo          core.MetaRepository
	crosswalkRepo core.CrosswalkRepository
}

func NewMetaRoutes(repo core.MetaRepository, crosswalkRepo core.CrosswalkRepository) *MetaRoutes {
	return &MetaRoutes{repo: repo, crosswalkRepo: crosswalkRepo}
}

func (mr *MetaRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/ready", mr.handleReady)
	mux.HandleFunc("GET /v1/meta", mr.handleMeta)
	mux.HandleFunc("GET /v1/meta/datasets", mr.handleDatasetStatus)
	mux.HandleFunc("GET /v1/meta/readiness", mr.handleReadiness)
	mux.HandleFunc("GET /v1/meta/constants/woba", mr.handleWOBAConstants)
	mux.HandleFunc("GET /v1/meta/constants/league", mr.handleLeagueConstants)
	mux.HandleFunc("GET /v1/meta/constants/park-factors", mr.handleParkFactors)
	mux.HandleFunc("GET /v1/meta/crosswalk/teams", mr.handleTeamCrosswalkList)
	mux.HandleFunc("GET /v1/meta/crosswalk/teams/by-team/{team_id}", mr.handleTeamCrosswalkByTeam)
	mux.HandleFunc("GET /v1/meta/crosswalk/teams/by-franchise/{franchise_id}", mr.handleTeamCrosswalkByFranchise)
	mux.HandleFunc("GET /v1/meta/crosswalk/teams/by-mlbam/{mlbam_team_id}", mr.handleTeamCrosswalkByMLBAM)
	mux.HandleFunc("GET /v1/meta/crosswalk/players", mr.handlePlayerCrosswalkList)
	mux.HandleFunc("GET /v1/meta/crosswalk/players/by-player/{player_id}", mr.handlePlayerCrosswalkByPlayer)
	mux.HandleFunc("GET /v1/meta/crosswalk/players/by-retro/{retro_id}", mr.handlePlayerCrosswalkByRetro)
	mux.HandleFunc("GET /v1/meta/crosswalk/players/by-mlbam/{mlbam_id}", mr.handlePlayerCrosswalkByMLBAM)
}

type metaResponse struct {
	Version      string                     `json:"version"`
	GeneratedAt  time.Time                  `json:"generated_at"`
	Coverage     map[string]datasetCoverage `json:"coverage"`
	EraLabels    map[string]string          `json:"era_labels"`
	SchemaHashes map[string]string          `json:"schema_hashes"`
	Datasets     []core.DatasetStatus       `json:"datasets"`
}

type datasetCoverage struct {
	From *core.SeasonYear `json:"from,omitempty"`
	To   *core.SeasonYear `json:"to,omitempty"`
}

// handleMeta godoc
//
//	@Summary		API metadata
//	@Description	Returns API version, dataset freshness, coverage, era label mappings, and schema fingerprints
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			strict	query		bool	false	"Use strict exact row counts (falls back to lightweight mode on error/timeout)"
//	@Success		200	{object}	metaResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/meta [get]
func (mr *MetaRoutes) handleMeta(w http.ResponseWriter, r *http.Request) {
	datasets, mode, err := mr.loadDatasetsWithMode(r)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set(countModeHeader, string(mode))

	ctx := r.Context()

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
		EraLabels:    eraLabelsMap(),
		SchemaHashes: schemaHashes,
		Datasets:     datasets,
	}

	writeJSON(w, http.StatusOK, resp)
}

func eraLabelsMap() map[string]string {
	eras := seed.GetAllEras()
	labels := make(map[string]string, len(eras))
	for _, era := range eras {
		labels[era.ShortName] = era.Name
	}
	return labels
}

// handleDatasetStatus godoc
//
//	@Summary		Dataset status
//	@Description	Returns dataset ETL metadata and coverage
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			strict	query		bool	false	"Use strict exact row counts (falls back to lightweight mode on error/timeout)"
//	@Success		200	{array}		core.DatasetStatus
//	@Failure		500	{object}	ErrorResponse
//	@Router			/meta/datasets [get]
func (mr *MetaRoutes) handleDatasetStatus(w http.ResponseWriter, r *http.Request) {
	datasets, mode, err := mr.loadDatasetsWithMode(r)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set(countModeHeader, string(mode))
	writeJSON(w, http.StatusOK, datasets)
}

// handleReadiness godoc
//
//	@Summary		Dataset readiness
//	@Description	Returns whether the required datasets are loaded for the core API routes
//	@Tags			meta, health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	core.ReadinessStatus
//	@Failure		500	{object}	ErrorResponse
//	@Router			/meta/readiness [get]
func (mr *MetaRoutes) handleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx := core.WithCountMode(r.Context(), core.CountModeLightweight)
	readiness, err := mr.repo.Readiness(ctx)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set(countModeHeader, string(core.EffectiveCountModeFromContext(ctx)))
	writeJSON(w, http.StatusOK, readiness)
}

// handleReady godoc
//
//	@Summary		Readiness check
//	@Description	Returns HTTP 200 when required datasets are loaded, or 503 when the API is live but not fully ready
//	@Tags			health, meta
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	core.ReadinessStatus
//	@Failure		503	{object}	core.ReadinessStatus
//	@Failure		500	{object}	ErrorResponse
//	@Router			/ready [get]
func (mr *MetaRoutes) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx := core.WithCountMode(r.Context(), core.CountModeLightweight)
	readiness, err := mr.repo.Readiness(ctx)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set(countModeHeader, string(core.EffectiveCountModeFromContext(ctx)))

	status := http.StatusOK
	if !readiness.Ready {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, readiness)
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

func (mr *MetaRoutes) loadDatasetsWithMode(r *http.Request) ([]core.DatasetStatus, core.CountMode, error) {
	requestedMode := core.CountModeLightweight
	if isTruthyParam(r.URL.Query().Get("strict")) {
		requestedMode = core.CountModeStrict
	}

	ctx := core.WithCountMode(r.Context(), requestedMode)
	if requestedMode == core.CountModeStrict {
		strictCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		datasets, err := mr.repo.DatasetStatuses(strictCtx)
		if err == nil {
			return datasets, core.EffectiveCountModeFromContext(strictCtx), nil
		}

		fallbackCtx := core.WithCountMode(r.Context(), core.CountModeLightweight)
		datasets, fallbackErr := mr.repo.DatasetStatuses(fallbackCtx)
		if fallbackErr != nil {
			return nil, "", err
		}
		return datasets, core.CountModeFallback, nil
	}

	datasets, err := mr.repo.DatasetStatuses(ctx)
	if err != nil {
		return nil, "", err
	}
	return datasets, core.EffectiveCountModeFromContext(ctx), nil
}

// handleWOBAConstants godoc
//
//	@Summary		wOBA constants
//	@Description	Returns season-specific wOBA calculation constants from FanGraphs
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			season	query		int	false	"Season year (returns all if omitted)"
//	@Success		200		{array}		core.WOBAConstant
//	@Failure		500		{object}	ErrorResponse
//	@Router			/meta/constants/woba [get]
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
//
//	@Summary		League constants
//	@Description	Returns league-specific constants for wRC+ and WAR calculations
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			season	query		int		false	"Season year (returns all if omitted)"
//	@Param			league	query		string	false	"League (AL or NL)"
//	@Success		200		{array}		core.LeagueConstant
//	@Failure		500		{object}	ErrorResponse
//	@Router			/meta/constants/league [get]
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
//
//	@Summary		Park factors
//	@Description	Returns FanGraphs park factors for seasons
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			season	query		int		false	"Season year (returns all if omitted)"
//	@Param			team	query		string	false	"Team ID filter"
//	@Success		200		{array}		core.ParkFactorRow
//	@Failure		500		{object}	ErrorResponse
//	@Router			/meta/constants/park-factors [get]
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

func isTruthyParam(value string) bool {
	switch value {
	case "1", "true", "TRUE", "True", "yes", "YES", "on", "ON":
		return true
	default:
		return false
	}
}

// handleTeamCrosswalkList godoc
//
//	@Summary		Team crosswalk list
//	@Description	List team/franchise/MLBAM ID crosswalk rows. Defaults to current season unless all_seasons=true.
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			season			query		int		false	"Season year (defaults to current year when all_seasons=false)"
//	@Param			all_seasons		query		bool	false	"Return all seasons"
//	@Param			team_id			query		string	false	"Local team ID filter"
//	@Param			franchise_id	query		string	false	"Local franchise ID filter"
//	@Param			mlbam_team_id	query		int		false	"MLBAM team ID filter"
//	@Success		200				{object}	core.TeamCrosswalkResponse
//	@Failure		500				{object}	ErrorResponse
//	@Router			/meta/crosswalk/teams [get]
func (mr *MetaRoutes) handleTeamCrosswalkList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := core.TeamCrosswalkFilter{
		AllSeasons: isTruthyParam(r.URL.Query().Get("all_seasons")),
	}

	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		s := core.SeasonYear(getIntQuery(r, "season", 0))
		if s > 0 {
			filter.Season = &s
		}
	}
	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		id := core.TeamID(teamID)
		filter.TeamID = &id
	}
	if franchiseID := r.URL.Query().Get("franchise_id"); franchiseID != "" {
		id := core.FranchiseID(franchiseID)
		filter.FranchiseID = &id
	}
	if mlbamIDStr := r.URL.Query().Get("mlbam_team_id"); mlbamIDStr != "" {
		id := getIntQuery(r, "mlbam_team_id", 0)
		if id > 0 {
			filter.MLBAMTeamID = &id
		}
	}

	rows, err := mr.crosswalkRepo.ListTeamCrosswalk(ctx, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := core.TeamCrosswalkResponse{
		AllSeasons: filter.AllSeasons,
		Total:      len(rows),
		Rows:       rows,
	}
	if !filter.AllSeasons {
		if filter.Season != nil {
			resp.Season = filter.Season
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleTeamCrosswalkByTeam godoc
//
//	@Summary		Team crosswalk by team_id
//	@Description	Get crosswalk rows for a local team ID.
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			team_id	path		string	true	"Local team ID"
//	@Param			season	query		int		false	"Season year"
//	@Success		200		{object}	core.TeamCrosswalkResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/meta/crosswalk/teams/by-team/{team_id} [get]
func (mr *MetaRoutes) handleTeamCrosswalkByTeam(w http.ResponseWriter, r *http.Request) {
	teamID := core.TeamID(r.PathValue("team_id"))
	filter := core.TeamCrosswalkFilter{TeamID: &teamID, AllSeasons: true}
	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		s := core.SeasonYear(getIntQuery(r, "season", 0))
		if s > 0 {
			filter.Season = &s
			filter.AllSeasons = false
		}
	}
	rows, err := mr.crosswalkRepo.ListTeamCrosswalk(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, core.TeamCrosswalkResponse{Season: filter.Season, AllSeasons: filter.AllSeasons, Total: len(rows), Rows: rows})
}

// handleTeamCrosswalkByFranchise godoc
//
//	@Summary		Team crosswalk by franchise_id
//	@Description	Get crosswalk rows for a local franchise ID.
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			franchise_id	path		string	true	"Local franchise ID"
//	@Param			season			query		int		false	"Season year"
//	@Success		200				{object}	core.TeamCrosswalkResponse
//	@Failure		500				{object}	ErrorResponse
//	@Router			/meta/crosswalk/teams/by-franchise/{franchise_id} [get]
func (mr *MetaRoutes) handleTeamCrosswalkByFranchise(w http.ResponseWriter, r *http.Request) {
	franchiseID := core.FranchiseID(r.PathValue("franchise_id"))
	filter := core.TeamCrosswalkFilter{FranchiseID: &franchiseID, AllSeasons: true}
	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		s := core.SeasonYear(getIntQuery(r, "season", 0))
		if s > 0 {
			filter.Season = &s
			filter.AllSeasons = false
		}
	}
	rows, err := mr.crosswalkRepo.ListTeamCrosswalk(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, core.TeamCrosswalkResponse{Season: filter.Season, AllSeasons: filter.AllSeasons, Total: len(rows), Rows: rows})
}

// handleTeamCrosswalkByMLBAM godoc
//
//	@Summary		Team crosswalk by MLBAM team ID
//	@Description	Get crosswalk rows for an MLBAM team ID.
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			mlbam_team_id	path		int	true	"MLBAM team ID"
//	@Param			season			query		int	false	"Season year"
//	@Success		200				{object}	core.TeamCrosswalkResponse
//	@Failure		500				{object}	ErrorResponse
//	@Router			/meta/crosswalk/teams/by-mlbam/{mlbam_team_id} [get]
func (mr *MetaRoutes) handleTeamCrosswalkByMLBAM(w http.ResponseWriter, r *http.Request) {
	mlbamTeamID := getIntPathValue(r, "mlbam_team_id")
	filter := core.TeamCrosswalkFilter{MLBAMTeamID: &mlbamTeamID, AllSeasons: true}
	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		s := core.SeasonYear(getIntQuery(r, "season", 0))
		if s > 0 {
			filter.Season = &s
			filter.AllSeasons = false
		}
	}
	rows, err := mr.crosswalkRepo.ListTeamCrosswalk(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, core.TeamCrosswalkResponse{Season: filter.Season, AllSeasons: filter.AllSeasons, Total: len(rows), Rows: rows})
}

// handlePlayerCrosswalkList godoc
//
//	@Summary		Player crosswalk list
//	@Description	List player ID mappings across local and MLBAM identifiers.
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			player_id	query		string	false	"Local player ID filter"
//	@Param			retro_id	query		string	false	"Retrosheet ID filter"
//	@Param			mlbam_id	query		int		false	"MLBAM person ID filter"
//	@Success		200			{object}	core.PlayerCrosswalkResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/meta/crosswalk/players [get]
func (mr *MetaRoutes) handlePlayerCrosswalkList(w http.ResponseWriter, r *http.Request) {
	filter := core.PlayerCrosswalkFilter{}
	if playerID := r.URL.Query().Get("player_id"); playerID != "" {
		id := core.PlayerID(playerID)
		filter.PlayerID = &id
	}
	if retroID := r.URL.Query().Get("retro_id"); retroID != "" {
		id := core.RetroPlayerID(retroID)
		filter.RetroID = &id
	}
	if mlbamID := r.URL.Query().Get("mlbam_id"); mlbamID != "" {
		id := getIntQuery(r, "mlbam_id", 0)
		if id > 0 {
			filter.MLBAMID = &id
		}
	}

	rows, err := mr.crosswalkRepo.ListPlayerCrosswalk(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, core.PlayerCrosswalkResponse{Total: len(rows), Rows: rows})
}

// handlePlayerCrosswalkByPlayer godoc
//
//	@Summary		Player crosswalk by player_id
//	@Description	Get crosswalk mappings for a local player ID.
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			player_id	path		string	true	"Local player ID"
//	@Success		200			{object}	core.PlayerCrosswalkResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/meta/crosswalk/players/by-player/{player_id} [get]
func (mr *MetaRoutes) handlePlayerCrosswalkByPlayer(w http.ResponseWriter, r *http.Request) {
	playerID := core.PlayerID(r.PathValue("player_id"))
	rows, err := mr.crosswalkRepo.ListPlayerCrosswalk(r.Context(), core.PlayerCrosswalkFilter{PlayerID: &playerID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, core.PlayerCrosswalkResponse{Total: len(rows), Rows: rows})
}

// handlePlayerCrosswalkByRetro godoc
//
//	@Summary		Player crosswalk by retro_id
//	@Description	Get crosswalk mappings for a Retrosheet player ID.
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			retro_id	path		string	true	"Retrosheet player ID"
//	@Success		200			{object}	core.PlayerCrosswalkResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/meta/crosswalk/players/by-retro/{retro_id} [get]
func (mr *MetaRoutes) handlePlayerCrosswalkByRetro(w http.ResponseWriter, r *http.Request) {
	retroID := core.RetroPlayerID(r.PathValue("retro_id"))
	rows, err := mr.crosswalkRepo.ListPlayerCrosswalk(r.Context(), core.PlayerCrosswalkFilter{RetroID: &retroID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, core.PlayerCrosswalkResponse{Total: len(rows), Rows: rows})
}

// handlePlayerCrosswalkByMLBAM godoc
//
//	@Summary		Player crosswalk by MLBAM ID
//	@Description	Get crosswalk mappings for an MLBAM person ID.
//	@Tags			meta
//	@Accept			json
//	@Produce		json
//	@Param			mlbam_id	path		int	true	"MLBAM person ID"
//	@Success		200			{object}	core.PlayerCrosswalkResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/meta/crosswalk/players/by-mlbam/{mlbam_id} [get]
func (mr *MetaRoutes) handlePlayerCrosswalkByMLBAM(w http.ResponseWriter, r *http.Request) {
	mlbamID := getIntPathValue(r, "mlbam_id")
	rows, err := mr.crosswalkRepo.ListPlayerCrosswalk(r.Context(), core.PlayerCrosswalkFilter{MLBAMID: &mlbamID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, core.PlayerCrosswalkResponse{Total: len(rows), Rows: rows})
}
