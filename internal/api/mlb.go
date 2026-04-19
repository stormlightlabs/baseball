package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/cache"
	"stormlightlabs.org/baseball/internal/core"
)

const mlbStatsAPIBase string = "https://statsapi.mlb.com/api"

var mlbProxyCatalog = []core.MLBProxyCatalogItem{
	{Route: "/v1/mlb/people", Target: "/v1/people", Description: "Player search and metadata from MLBAM"},
	{Route: "/v1/mlb/people/{id}", Target: "/v1/people/{personId}", Description: "Single player lookup"},
	{Route: "/v1/mlb/teams", Target: "/v1/teams", Description: "MLB team reference and roster metadata"},
	{Route: "/v1/mlb/teams/{id}", Target: "/v1/teams/{teamId}", Description: "Single team details"},
	{Route: "/v1/mlb/crosswalk/teams", Target: "local+mlb", Description: "Crosswalk MLB team IDs to local team/franchise IDs"},
	{Route: "/v1/mlb/schedule", Target: "/v1/schedule", Description: "Daily/season schedule with game metadata"},
	{Route: "/v1/mlb/seasons", Target: "/v1/seasons", Description: "Season directory with league/division data"},
	{Route: "/v1/mlb/stats", Target: "/v1/stats", Description: "MLB-wide stats queries"},
	{Route: "/v1/mlb/standings", Target: "/v1/standings", Description: "League/division standings"},
	{Route: "/v1/mlb/awards", Target: "/v1/awards", Description: "Awards directory and recipients"},
	{Route: "/v1/mlb/awards/{id}", Target: "/v1/awards/{awardId}", Description: "Single MLB awards endpoint"},
	{Route: "/v1/mlb/venues", Target: "/v1/venues", Description: "Ballpark directory"},
}

// MLBRoutes proxies select statsapi.mlb.com endpoints through /v1/mlb with HTTP caching
type MLBRoutes struct {
	client   *http.Client
	cache    *cache.Client
	teamRepo core.TeamRepository
	baseURL  string
}

func NewMLBStatsAPIRoutes(cacheClient *cache.Client, teamRepo core.TeamRepository) *MLBRoutes {
	return &MLBRoutes{
		client:   &http.Client{Timeout: 10 * time.Second},
		cache:    cacheClient,
		teamRepo: teamRepo,
		baseURL:  mlbStatsAPIBase,
	}
}

func (mr *MLBRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/mlb", mr.handleMLBOverview)
	mux.HandleFunc("GET /v1/mlb/people", mr.handleMLBPeople)
	mux.HandleFunc("GET /v1/mlb/people/{id}", mr.handleMLBPerson)
	mux.HandleFunc("GET /v1/mlb/teams", mr.handleMLBTeams)
	mux.HandleFunc("GET /v1/mlb/teams/{id}", mr.handleMLBTeam)
	mux.HandleFunc("GET /v1/mlb/crosswalk/teams", mr.handleMLBTeamCrosswalk)
	mux.HandleFunc("GET /v1/mlb/schedule", mr.handleMLBSchedule)
	mux.HandleFunc("GET /v1/mlb/seasons", mr.handleMLBSeasons)
	mux.HandleFunc("GET /v1/mlb/stats", mr.handleMLBStats)
	mux.HandleFunc("GET /v1/mlb/standings", mr.handleMLBStandings)
	mux.HandleFunc("GET /v1/mlb/awards", mr.handleMLBAwards)
	mux.HandleFunc("GET /v1/mlb/awards/{id}", mr.handleMLBAward)
	mux.HandleFunc("GET /v1/mlb/venues", mr.handleMLBVenues)
}

// handleMLBOverview godoc
//
//	@Summary		MLB Stats proxy catalog
//	@Description	Lists available MLB Stats API proxy routes surfaced under /v1/mlb. All endpoints default to sportId=1 (Major League Baseball) unless specified.
//	@Tags			mlb
//	@Produce		json
//	@Success		200	{object}	core.MLBOverviewResponse
//	@Router			/mlb [get]
func (mr *MLBRoutes) handleMLBOverview(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, core.MLBOverviewResponse{
		BaseURL: "/v1/mlb",
		Target:  mlbStatsAPIBase,
		Routes:  mlbProxyCatalog,
	})
}

// handleMLBPeople godoc
//
//	@Summary		MLB people search
//	@Description	Proxy to MLB Stats API /v1/people for live roster metadata. Defaults to sportId=1 (Major League Baseball) if not provided.
//	@Tags			mlb
//	@Accept			json
//	@Produce		json
//	@Param			personIds	query		string	false	"Comma-separated MLBAM personIds"
//	@Param			sportId		query		string	false	"Filter by sportId (defaults to 1 for MLB)"
//	@Param			hydrate		query		string	false	"Hydrate relationships"
//	@Success		200			{object}	core.MLBPeopleResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/mlb/people [get]
func (mr *MLBRoutes) handleMLBPeople(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "people")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBPeopleResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBPerson godoc
//
//	@Summary		MLB person by ID
//	@Description	Proxy to MLB Stats API /v1/people/{personId}
//	@Tags			mlb
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"MLBAM personId"
//	@Param			hydrate	query		string	false	"Hydrate relationships"
//	@Success		200		{object}	core.MLBPeopleResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/mlb/people/{id} [get]
func (mr *MLBRoutes) handleMLBPerson(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "people", r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBPeopleResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBTeams godoc
//
//	@Summary		MLB teams
//	@Description	Proxy to MLB Stats API /v1/teams. Defaults to sportId=1 (Major League Baseball) if not provided.
//	@Tags			mlb
//	@Accept			json
//	@Produce		json
//	@Param			sportId	query		string	false	"Sport filter (defaults to 1 for MLB)"
//	@Param			season	query		string	false	"Season year"
//	@Success		200		{object}	core.MLBTeamsResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/mlb/teams [get]
func (mr *MLBRoutes) handleMLBTeams(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "teams")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBTeamsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBTeam godoc
//
//	@Summary		MLB team by ID
//	@Description	Proxy to MLB Stats API /v1/teams/{teamId}
//	@Tags			mlb
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"MLB teamId"
//	@Param			season	query		string	false	"Season year"
//	@Success		200		{object}	core.MLBTeamsResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/mlb/teams/{id} [get]
func (mr *MLBRoutes) handleMLBTeam(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "teams", r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBTeamsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

var mlbToLocalTeamCodeCandidates = map[string][]string{
	"ARI": {"ARI"},
	"ATL": {"ATL"},
	"BAL": {"BAL"},
	"BOS": {"BOS"},
	"CHC": {"CHN", "CHC"},
	"CWS": {"CHA", "CHW"},
	"CIN": {"CIN"},
	"CLE": {"CLE"},
	"COL": {"COL"},
	"DET": {"DET"},
	"HOU": {"HOU"},
	"KC":  {"KCA", "KCR"},
	"LAA": {"LAA", "ANA"},
	"LAD": {"LAN", "LAD"},
	"MIA": {"MIA", "FLO"},
	"MIL": {"MIL"},
	"MIN": {"MIN"},
	"NYM": {"NYN", "NYM"},
	"NYY": {"NYA", "NYY"},
	"ATH": {"OAK", "ATH"},
	"PHI": {"PHI"},
	"PIT": {"PIT"},
	"SD":  {"SDN", "SDP"},
	"SEA": {"SEA"},
	"SF":  {"SFN", "SFG"},
	"STL": {"SLN", "STL"},
	"TB":  {"TBA", "TBD", "TBR"},
	"TEX": {"TEX"},
	"TOR": {"TOR"},
	"WSH": {"WAS", "WSN"},
}

type localTeamLookup struct {
	byTeamID      map[string]core.TeamSeason
	byFranchiseID map[string]core.TeamSeason
	byName        map[string]core.TeamSeason
}

func normalizeMLBLookupKey(value string) string {
	if value == "" {
		return ""
	}

	upper := strings.ToUpper(value)
	var b strings.Builder
	b.Grow(len(upper))
	for _, r := range upper {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func buildLocalTeamLookup(localTeams []core.TeamSeason) localTeamLookup {
	lookup := localTeamLookup{
		byTeamID:      make(map[string]core.TeamSeason, len(localTeams)),
		byFranchiseID: make(map[string]core.TeamSeason, len(localTeams)),
		byName:        make(map[string]core.TeamSeason, len(localTeams)),
	}

	for _, localTeam := range localTeams {
		teamID := normalizeMLBLookupKey(string(localTeam.TeamID))
		if teamID != "" {
			lookup.byTeamID[teamID] = localTeam
		}

		franchiseID := normalizeMLBLookupKey(string(localTeam.FranchiseID))
		if franchiseID != "" {
			lookup.byFranchiseID[franchiseID] = localTeam
		}

		teamName := normalizeMLBLookupKey(localTeam.Name)
		if teamName != "" {
			lookup.byName[teamName] = localTeam
		}
	}

	return lookup
}

func localCodeCandidatesForMLBTeam(team core.MLBTeam) []string {
	seen := map[string]struct{}{}
	add := func(raw string, out *[]string) {
		normalized := normalizeMLBLookupKey(raw)
		if normalized == "" {
			return
		}
		if _, exists := seen[normalized]; exists {
			return
		}
		seen[normalized] = struct{}{}
		*out = append(*out, normalized)
	}

	var candidates []string
	add(team.Abbreviation, &candidates)
	add(team.TeamCode, &candidates)
	add(team.FileCode, &candidates)

	if known := mlbToLocalTeamCodeCandidates[normalizeMLBLookupKey(team.Abbreviation)]; len(known) > 0 {
		for _, candidate := range known {
			add(candidate, &candidates)
		}
	}

	return candidates
}

func localNameCandidatesForMLBTeam(team core.MLBTeam) []string {
	seen := map[string]struct{}{}
	add := func(raw string, out *[]string) {
		normalized := normalizeMLBLookupKey(raw)
		if normalized == "" {
			return
		}
		if _, exists := seen[normalized]; exists {
			return
		}
		seen[normalized] = struct{}{}
		*out = append(*out, normalized)
	}

	var candidates []string
	add(team.Name, &candidates)
	add(team.FranchiseName, &candidates)
	add(team.ClubName, &candidates)
	add(team.TeamName, &candidates)
	add(team.ShortName, &candidates)

	if team.LocationName != "" && team.TeamName != "" {
		add(fmt.Sprintf("%s %s", team.LocationName, team.TeamName), &candidates)
	}
	return candidates
}

func (mr *MLBRoutes) resolveLocalTeamSeason(ctx context.Context, requestedSeason int) (int, []core.TeamSeason, error) {
	if mr.teamRepo == nil {
		return requestedSeason, nil, nil
	}

	requested := core.SeasonYear(requestedSeason)
	filter := core.TeamFilter{Year: &requested, Pagination: *core.NewPagination(1, 200)}
	localTeams, err := mr.teamRepo.ListTeamSeasons(ctx, filter)
	if err != nil {
		return requestedSeason, nil, err
	}
	if len(localTeams) > 0 {
		return requestedSeason, localTeams, nil
	}

	latestRows, err := mr.teamRepo.ListTeamSeasons(ctx, core.TeamFilter{Pagination: *core.NewPagination(1, 1)})
	if err != nil {
		return requestedSeason, nil, err
	}
	if len(latestRows) == 0 {
		return requestedSeason, nil, nil
	}

	resolvedSeason := int(latestRows[0].Year)
	resolved := core.SeasonYear(resolvedSeason)
	localTeams, err = mr.teamRepo.ListTeamSeasons(ctx, core.TeamFilter{Year: &resolved, Pagination: *core.NewPagination(1, 200)})
	if err != nil {
		return requestedSeason, nil, err
	}

	return resolvedSeason, localTeams, nil
}

// handleMLBTeamCrosswalk godoc
//
//	@Summary		MLB team ID crosswalk
//	@Description	Map MLB Stats API team IDs/abbreviations to local team_id and franchise_id for a season.
//	@Tags			mlb
//	@Accept			json
//	@Produce		json
//	@Param			season	query		integer	false	"Season year (defaults to current year)"
//	@Success		200		{object}	core.MLBTeamCrosswalkResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/mlb/crosswalk/teams [get]
func (mr *MLBRoutes) handleMLBTeamCrosswalk(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "teams")
	if err != nil {
		writeError(w, err)
		return
	}

	requestedSeason := getIntQuery(r, "season", time.Now().Year())
	upstreamRequest := r.Clone(r.Context())
	query := upstreamRequest.URL.Query()
	query.Set("season", strconv.Itoa(requestedSeason))
	upstreamRequest.URL.RawQuery = query.Encode()

	body, statusCode, err := mr.fetchFromMLB(upstreamRequest, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var mlbTeams core.MLBTeamsResponse
	if err := json.Unmarshal(body, &mlbTeams); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to parse MLB teams response: %w", err))
		return
	}

	localSeason, localTeams, err := mr.resolveLocalTeamSeason(r.Context(), requestedSeason)
	if err != nil {
		writeError(w, err)
		return
	}

	localLookup := buildLocalTeamLookup(localTeams)
	response := core.MLBTeamCrosswalkResponse{
		RequestedSeason: requestedSeason,
		LocalSeason:     localSeason,
		Rows:            make([]core.MLBTeamCrosswalkRow, 0, len(mlbTeams.Teams)),
	}

	for _, mlbTeam := range mlbTeams.Teams {
		row := core.MLBTeamCrosswalkRow{
			MLBTeamID:        mlbTeam.ID,
			MLBAbbreviation:  mlbTeam.Abbreviation,
			MLBTeamCode:      mlbTeam.TeamCode,
			MLBFileCode:      mlbTeam.FileCode,
			MLBTeamName:      mlbTeam.Name,
			MLBFranchiseName: mlbTeam.FranchiseName,
			MLBClubName:      mlbTeam.ClubName,
		}

		var matchedTeam core.TeamSeason
		matched := false
		for _, candidate := range localCodeCandidatesForMLBTeam(mlbTeam) {
			if team, ok := localLookup.byTeamID[candidate]; ok {
				matchedTeam = team
				row.MatchMethod = "team_id"
				row.Confidence = "high"
				matched = true
				break
			}
			if team, ok := localLookup.byFranchiseID[candidate]; ok {
				matchedTeam = team
				row.MatchMethod = "franchise_id"
				row.Confidence = "high"
				matched = true
				break
			}
		}

		if !matched {
			for _, candidate := range localNameCandidatesForMLBTeam(mlbTeam) {
				if team, ok := localLookup.byName[candidate]; ok {
					matchedTeam = team
					row.MatchMethod = "name"
					row.Confidence = "medium"
					matched = true
					break
				}
			}
		}

		if matched {
			localTeamID := matchedTeam.TeamID
			localFranchiseID := matchedTeam.FranchiseID
			localLeague := matchedTeam.League
			row.LocalTeamID = &localTeamID
			row.LocalFranchiseID = &localFranchiseID
			row.LocalLeague = &localLeague
			row.LocalTeamName = matchedTeam.Name
			response.Matched++
		} else {
			row.Confidence = "none"
			response.Unmatched++
		}

		response.Rows = append(response.Rows, row)
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, response)
}

// handleMLBSchedule godoc
//
//	@Summary		MLB schedule
//	@Description	Proxy to MLB Stats API /v1/schedule. Defaults to sportId=1 (Major League Baseball) if not provided.
//	@Tags			mlb
//	@Accept			json
//	@Produce		json
//	@Param			sportId	query		string	false	"Sport filter (defaults to 1 for MLB)"
//	@Param			teamId	query		string	false	"Team filter"
//	@Param			season	query		string	false	"Season year"
//	@Param			date	query		string	false	"Specific date (YYYY-MM-DD)"
//	@Param			hydrate	query		string	false	"Hydrate payload sections (e.g. linescore,team)"
//	@Success		200		{object}	core.MLBScheduleResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/mlb/schedule [get]
func (mr *MLBRoutes) handleMLBSchedule(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "schedule")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBScheduleResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBSeasons godoc
//
//	@Summary		MLB seasons
//	@Description	Proxy to MLB Stats API /v1/seasons. Defaults to sportId=1 (Major League Baseball) if not provided.
//	@Tags			mlb
//	@Accept			json
//	@Produce		json
//	@Param			sportId	query		string	false	"Sport filter (defaults to 1 for MLB)"
//	@Param			season	query		string	false	"Season year"
//	@Success		200		{object}	core.MLBSeasonsResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/mlb/seasons [get]
func (mr *MLBRoutes) handleMLBSeasons(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "seasons")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBSeasonsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBStats godoc
//
//	@Summary		MLB stats queries
//	@Description	Proxy to MLB Stats API /v1/stats for ad-hoc stats lookups. Note: sportId defaults to 1 (Major League Baseball).
//	@Tags			mlb
//	@Accept			json
//	@Produce		json
//	@Param			stats		query		string	true	"Stat group(s) to query"
//	@Param			group		query		string	true	"Grouping (e.g., hitting, pitching)"
//	@Param			season		query		string	false	"Season year"
//	@Param			gameType	query		string	false	"Game type (R, S, etc.)"
//	@Success		200			{object}	core.MLBStatsResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/mlb/stats [get]
func (mr *MLBRoutes) handleMLBStats(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "stats")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBStatsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBStandings godoc
//
//	@Summary		MLB standings
//	@Description	Proxy to MLB Stats API /v1/standings. Note: sportId defaults to 1 (Major League Baseball).
//	@Tags			mlb
//	@Accept			json
//	@Produce		json
//	@Param			leagueId		query		string	false	"League filter"
//	@Param			season			query		string	false	"Season year"
//	@Param			standingsTypes	query		string	false	"Standings type (byLeague, etc.)"
//	@Success		200				{object}	core.MLBStandingsResponse
//	@Failure		500				{object}	ErrorResponse
//	@Router			/mlb/standings [get]
func (mr *MLBRoutes) handleMLBStandings(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "standings")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBStandingsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBAwards godoc
//
//	@Summary		MLB awards catalog
//	@Description	Proxy to MLB Stats API /v1/awards. Defaults to sportId=1 (Major League Baseball) if not provided.
//	@Tags			mlb
//	@Accept			json
//	@Produce		json
//	@Param			sportId	query		string	false	"Sport filter (defaults to 1 for MLB)"
//	@Param			season	query		string	false	"Season year"
//	@Success		200		{object}	core.MLBAwardsResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/mlb/awards [get]
func (mr *MLBRoutes) handleMLBAwards(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "awards")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBAwardsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBAward godoc
//
//	@Summary		MLB award by ID
//	@Description	Proxy to MLB Stats API /v1/awards/{awardId}
//	@Tags			mlb
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"MLB awardId"
//	@Param			season	query		string	false	"Season year"
//	@Param			hydrate	query		string	false	"Hydrate relationships"
//	@Success		200		{object}	core.MLBAwardsResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/mlb/awards/{id} [get]
func (mr *MLBRoutes) handleMLBAward(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "awards", r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBAwardsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// handleMLBVenues godoc
//
//	@Summary		MLB venues directory
//	@Description	Proxy to MLB Stats API /v1/venues. Defaults to sportId=1 (Major League Baseball) if not provided.
//	@Tags			mlb
//	@Accept			json
//	@Produce		json
//	@Param			venueIds	query		string	false	"Comma-separated venue IDs"
//	@Param			season		query		string	false	"Season year"
//	@Param			sportId		query		string	false	"Sport filter (defaults to 1 for MLB)"
//	@Success		200			{object}	core.MLBVenuesResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/mlb/venues [get]
func (mr *MLBRoutes) handleMLBVenues(w http.ResponseWriter, r *http.Request) {
	target, err := url.JoinPath(mr.baseURL, "v1", "venues")
	if err != nil {
		writeError(w, err)
		return
	}

	body, statusCode, err := mr.fetchFromMLB(r, target)
	if err != nil {
		writeError(w, err)
		return
	}

	var result core.MLBVenuesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		writeInternalServerError(w, fmt.Errorf("failed to parse MLB API response: %w", err))
		return
	}

	w.Header().Set("X-Proxy-Target", "statsapi.mlb.com")
	writeJSON(w, statusCode, result)
}

// fetchFromMLB handles the common HTTP request logic for all MLB proxy endpoints.
// It adds sportId=1 as default if not provided, sets User-Agent, and returns the response body and status code.
func (mr *MLBRoutes) fetchFromMLB(r *http.Request, target string) ([]byte, int, error) {
	if mr.cache == nil {
		return mr.fetchFromMLBUncached(r, target)
	}

	query := r.URL.Query()
	if query.Get("sportId") == "" {
		query.Set("sportId", "1")
	}
	pathAndQuery := fmt.Sprintf("%s?%s", target, query.Encode())
	cacheKey := mr.cache.UpstreamKey(http.MethodGet, "statsapi.mlb.com", pathAndQuery)

	if cachedEntry, ok := mr.cache.GetHTTPCache(r.Context(), cacheKey); ok {
		return cachedEntry.Body, cachedEntry.Status, nil
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		return nil, 0, err
	}
	req.URL.RawQuery = query.Encode()
	req.Header.Set("User-Agent", "Stormlight-Baseball-MLBProxy/1.0")

	mr.cache.AddConditionalHeaders(r.Context(), cacheKey, req)

	resp, err := mr.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		upstreamConfig := cache.DefaultUpstreamConfig()
		ttl := upstreamConfig.DetermineTTL(resp)
		_ = mr.cache.RefreshHTTPCache(r.Context(), cacheKey, ttl)

		if cachedEntry, ok := mr.cache.GetHTTPCache(r.Context(), cacheKey); ok {
			return cachedEntry.Body, cachedEntry.Status, nil
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		upstreamConfig := cache.DefaultUpstreamConfig()
		ttl := upstreamConfig.DetermineTTL(resp)
		_ = mr.cache.CacheHTTPResponse(r.Context(), cacheKey, resp, body, ttl)
	}

	if resp.StatusCode >= 400 {
		retryAfter := resp.Header.Get("Retry-After")
		_ = mr.cache.CacheNegativeResponse(r.Context(), cacheKey, resp.StatusCode, string(body), retryAfter)
	}

	return body, resp.StatusCode, nil
}

// fetchFromMLBUncached makes an uncached request to MLB Stats API (fallback when cache unavailable)
func (mr *MLBRoutes) fetchFromMLBUncached(r *http.Request, target string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		return nil, 0, err
	}

	query := r.URL.Query()
	if query.Get("sportId") == "" {
		query.Set("sportId", "1")
	}
	req.URL.RawQuery = query.Encode()
	req.Header.Set("User-Agent", "Stormlight-Baseball-MLBProxy/1.0")

	resp, err := mr.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}
