package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type responseDetailsEnricher struct {
	db *sql.DB
}

func newResponseDetailsEnricher(db *sql.DB) *responseDetailsEnricher {
	if db == nil {
		return nil
	}
	return &responseDetailsEnricher{db: db}
}

func requestWantsDetails(r *http.Request) bool {
	if r == nil || r.Method != http.MethodGet {
		return false
	}
	for _, raw := range r.URL.Query()["include"] {
		for _, token := range strings.Split(raw, ",") {
			if strings.EqualFold(strings.TrimSpace(token), "details") {
				return true
			}
		}
	}
	return false
}

type bufferedResponseWriter struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{header: make(http.Header)}
}

func (w *bufferedResponseWriter) Header() http.Header { return w.header }

func (w *bufferedResponseWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}

func (w *bufferedResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(p)
}

func (w *bufferedResponseWriter) StatusCode() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

type idCollector struct {
	localTeamIDs      map[string]struct{}
	localFranchiseIDs map[string]struct{}
	localPlayerIDs    map[string]struct{}
	retroIDs          map[string]struct{}
	mlbamTeamIDs      map[int]struct{}
	mlbamPlayerIDs    map[int]struct{}
}

func newIDCollector() *idCollector {
	return &idCollector{
		localTeamIDs:      map[string]struct{}{},
		localFranchiseIDs: map[string]struct{}{},
		localPlayerIDs:    map[string]struct{}{},
		retroIDs:          map[string]struct{}{},
		mlbamTeamIDs:      map[int]struct{}{},
		mlbamPlayerIDs:    map[int]struct{}{},
	}
}

func (c *idCollector) addString(target map[string]struct{}, value any) {
	s := toTrimmedString(value)
	if s == "" {
		return
	}
	target[s] = struct{}{}
}

func (c *idCollector) addInt(target map[int]struct{}, value any) {
	n, ok := toInt(value)
	if !ok || n <= 0 {
		return
	}
	target[n] = struct{}{}
}

func toTrimmedString(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

func toInt(v any) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int64:
		return int(t), true
	case float64:
		if t == float64(int(t)) {
			return int(t), true
		}
		return 0, false
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(t))
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

func collectIDsFromJSON(value any, parentKey string, collector *idCollector) {
	switch node := value.(type) {
	case map[string]any:
		for key, child := range node {
			lower := strings.ToLower(key)
			switch lower {
			case "team_id", "teamid", "home_team", "away_team", "team_lookup":
				collector.addString(collector.localTeamIDs, child)
			case "franchise_id", "franchiseid":
				collector.addString(collector.localFranchiseIDs, child)
			case "player_id", "playerid":
				collector.addString(collector.localPlayerIDs, child)
			case "retro_id", "retroid":
				collector.addString(collector.retroIDs, child)
			case "mlbam_team_id", "team_mlbam_id", "mlb_team_id", "teammlbid":
				collector.addInt(collector.mlbamTeamIDs, child)
			case "mlbam_id", "mlb_id", "player_mlbam_id", "playermlbid":
				collector.addInt(collector.mlbamPlayerIDs, child)
			case "team":
				if childMap, ok := child.(map[string]any); ok {
					collector.addInt(collector.mlbamTeamIDs, childMap["id"])
				}
			case "player", "batter", "pitcher", "ondeck", "inhole", "first", "second", "third":
				if childMap, ok := child.(map[string]any); ok {
					collector.addInt(collector.mlbamPlayerIDs, childMap["id"])
				}
			case "currentteam":
				if childMap, ok := child.(map[string]any); ok {
					collector.addInt(collector.mlbamTeamIDs, childMap["id"])
				}
			case "id":
				if parentKey == "team" || parentKey == "currentteam" {
					collector.addInt(collector.mlbamTeamIDs, child)
				}
				if parentKey == "player" || parentKey == "batter" || parentKey == "pitcher" {
					collector.addInt(collector.mlbamPlayerIDs, child)
				}
			}
			collectIDsFromJSON(child, lower, collector)
		}
	case []any:
		for _, child := range node {
			collectIDsFromJSON(child, parentKey, collector)
		}
	}
}

func setToSortedStrings(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for v := range values {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func setToSortedInts(values map[int]struct{}) []int {
	out := make([]int, 0, len(values))
	for v := range values {
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}

func makePlaceholders(start, count int) string {
	parts := make([]string, 0, count)
	for i := 0; i < count; i++ {
		parts = append(parts, fmt.Sprintf("$%d", start+i))
	}
	return strings.Join(parts, ",")
}

func (e *responseDetailsEnricher) queryTeamDetails(ctx context.Context, ids []string) (map[string]any, error) {
	out := map[string]any{}
	if len(ids) == 0 {
		return out, nil
	}

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	query := fmt.Sprintf(`
		SELECT DISTINCT ON (t."teamID") t."teamID", t."franchID", t.name, t."lgID", t."yearID"
		FROM "Teams" t
		WHERE t."teamID" IN (%s)
		ORDER BY t."teamID", t."yearID" DESC
	`, makePlaceholders(1, len(ids)))

	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var teamID, franchiseID, name, league string
		var season int
		if err := rows.Scan(&teamID, &franchiseID, &name, &league, &season); err != nil {
			return nil, err
		}
		out[teamID] = map[string]any{
			"team_id":      teamID,
			"franchise_id": franchiseID,
			"name":         name,
			"league":       league,
			"season":       season,
		}
	}
	return out, rows.Err()
}

func (e *responseDetailsEnricher) queryFranchiseDetails(ctx context.Context, ids []string) (map[string]any, error) {
	out := map[string]any{}
	if len(ids) == 0 {
		return out, nil
	}
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	query := fmt.Sprintf(`
		SELECT "franchID", "franchName", "active", "NAassoc"
		FROM "TeamsFranchises"
		WHERE "franchID" IN (%s)
	`, makePlaceholders(1, len(ids)))

	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		var active, naAssoc sql.NullString
		if err := rows.Scan(&id, &name, &active, &naAssoc); err != nil {
			return nil, err
		}
		out[id] = map[string]any{
			"franchise_id": id,
			"name":         name,
			"active":       active.String,
			"na_assoc":     naAssoc.String,
		}
	}
	return out, rows.Err()
}

func (e *responseDetailsEnricher) queryPlayerDetails(ctx context.Context, ids []string) (map[string]any, error) {
	out := map[string]any{}
	if len(ids) == 0 {
		return out, nil
	}
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	query := fmt.Sprintf(`
		SELECT "playerID", "retroID", "nameFirst", "nameLast", "nameGiven", "debut", "finalGame"
		FROM "People"
		WHERE "playerID" IN (%s)
	`, makePlaceholders(1, len(ids)))

	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, firstName, lastName, givenName sql.NullString
		var retroID sql.NullString
		var debut, finalGame sql.NullString
		if err := rows.Scan(&id, &retroID, &firstName, &lastName, &givenName, &debut, &finalGame); err != nil {
			return nil, err
		}
		if !id.Valid {
			continue
		}
		out[id.String] = map[string]any{
			"player_id":  id.String,
			"retro_id":   retroID.String,
			"first_name": firstName.String,
			"last_name":  lastName.String,
			"given_name": givenName.String,
			"debut":      debut.String,
			"final_game": finalGame.String,
		}
	}
	return out, rows.Err()
}

func (e *responseDetailsEnricher) queryMLBAMPlayers(ctx context.Context, ids []int) (map[string]any, error) {
	out := map[string]any{}
	if len(ids) == 0 {
		return out, nil
	}
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	query := fmt.Sprintf(`
		SELECT mlbam_id, lahman_id, retro_id, bbref_id, full_name, source, confidence
		FROM player_mlbam_map
		WHERE mlbam_id IN (%s)
	`, makePlaceholders(1, len(ids)))
	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mlbamID int
		var lahmanID, retroID, bbrefID, fullName, source, confidence sql.NullString
		if err := rows.Scan(&mlbamID, &lahmanID, &retroID, &bbrefID, &fullName, &source, &confidence); err != nil {
			return nil, err
		}
		out[strconv.Itoa(mlbamID)] = map[string]any{
			"mlbam_id":   mlbamID,
			"player_id":  lahmanID.String,
			"retro_id":   retroID.String,
			"bbref_id":   bbrefID.String,
			"full_name":  fullName.String,
			"source":     source.String,
			"confidence": confidence.String,
		}
	}
	return out, rows.Err()
}

func (e *responseDetailsEnricher) queryMLBAMTeams(ctx context.Context, ids []int) (map[string]any, error) {
	out := map[string]any{}
	if len(ids) == 0 {
		return out, nil
	}
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	query := fmt.Sprintf(`
		SELECT DISTINCT ON (mlbam_team_id)
			season, mlbam_team_id, mlb_abbreviation, mlb_team_name, local_team_id, local_franchise_id, confidence, match_method
		FROM team_mlbam_map
		WHERE mlbam_team_id IN (%s)
		ORDER BY mlbam_team_id, season DESC
	`, makePlaceholders(1, len(ids)))
	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var season, mlbamTeamID int
		var abbr, teamName, localTeamID, localFranchiseID, confidence, method sql.NullString
		if err := rows.Scan(&season, &mlbamTeamID, &abbr, &teamName, &localTeamID, &localFranchiseID, &confidence, &method); err != nil {
			return nil, err
		}
		out[strconv.Itoa(mlbamTeamID)] = map[string]any{
			"season":           season,
			"mlbam_team_id":    mlbamTeamID,
			"mlb_abbreviation": abbr.String,
			"mlb_team_name":    teamName.String,
			"team_id":          localTeamID.String,
			"franchise_id":     localFranchiseID.String,
			"confidence":       confidence.String,
			"match_method":     method.String,
		}
	}
	return out, rows.Err()
}

func (e *responseDetailsEnricher) queryCrosswalkMaps(ctx context.Context, collector *idCollector) (map[string]any, error) {
	crosswalk := map[string]any{
		"retro_to_lahman":       map[string]any{},
		"lahman_to_retro":       map[string]any{},
		"mlbam_player_to_local": map[string]any{},
		"local_player_to_mlbam": map[string]any{},
		"team_to_franchise":     map[string]any{},
		"franchise_to_team":     map[string]any{},
		"mlbam_team_to_local":   map[string]any{},
		"local_team_to_mlbam":   map[string]any{},
	}

	retroIDs := setToSortedStrings(collector.retroIDs)
	if len(retroIDs) > 0 {
		args := make([]any, len(retroIDs))
		for i, id := range retroIDs {
			args[i] = id
		}
		query := fmt.Sprintf(`SELECT retro_id, lahman_id FROM player_id_map WHERE retro_id IN (%s)`, makePlaceholders(1, len(retroIDs)))
		rows, err := e.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var retroID, lahmanID string
			if err := rows.Scan(&retroID, &lahmanID); err != nil {
				rows.Close()
				return nil, err
			}
			crosswalk["retro_to_lahman"].(map[string]any)[retroID] = lahmanID
			crosswalk["lahman_to_retro"].(map[string]any)[lahmanID] = retroID
		}
		rows.Close()
	}

	playerIDs := setToSortedStrings(collector.localPlayerIDs)
	if len(playerIDs) > 0 {
		args := make([]any, len(playerIDs))
		for i, id := range playerIDs {
			args[i] = id
		}
		query := fmt.Sprintf(`SELECT lahman_id, mlbam_id FROM player_mlbam_map WHERE lahman_id IN (%s)`, makePlaceholders(1, len(playerIDs)))
		rows, err := e.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var lahmanID sql.NullString
			var mlbamID int
			if err := rows.Scan(&lahmanID, &mlbamID); err != nil {
				rows.Close()
				return nil, err
			}
			if !lahmanID.Valid || lahmanID.String == "" {
				continue
			}
			crosswalk["local_player_to_mlbam"].(map[string]any)[lahmanID.String] = mlbamID
			crosswalk["mlbam_player_to_local"].(map[string]any)[strconv.Itoa(mlbamID)] = lahmanID.String
		}
		rows.Close()
	}

	mlbamPlayerIDs := setToSortedInts(collector.mlbamPlayerIDs)
	if len(mlbamPlayerIDs) > 0 {
		args := make([]any, len(mlbamPlayerIDs))
		for i, id := range mlbamPlayerIDs {
			args[i] = id
		}
		query := fmt.Sprintf(`SELECT mlbam_id, lahman_id FROM player_mlbam_map WHERE mlbam_id IN (%s)`, makePlaceholders(1, len(mlbamPlayerIDs)))
		rows, err := e.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var mlbamID int
			var lahmanID sql.NullString
			if err := rows.Scan(&mlbamID, &lahmanID); err != nil {
				rows.Close()
				return nil, err
			}
			if lahmanID.Valid && lahmanID.String != "" {
				crosswalk["mlbam_player_to_local"].(map[string]any)[strconv.Itoa(mlbamID)] = lahmanID.String
				crosswalk["local_player_to_mlbam"].(map[string]any)[lahmanID.String] = mlbamID
			}
		}
		rows.Close()
	}

	teamIDs := setToSortedStrings(collector.localTeamIDs)
	if len(teamIDs) > 0 {
		args := make([]any, len(teamIDs))
		for i, id := range teamIDs {
			args[i] = id
		}
		query := fmt.Sprintf(`
			SELECT DISTINCT ON (team_id) team_id, franchise_id
			FROM team_franchise_map
			WHERE team_id IN (%s)
			ORDER BY team_id, season DESC
		`, makePlaceholders(1, len(teamIDs)))
		rows, err := e.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var teamID, franchiseID string
			if err := rows.Scan(&teamID, &franchiseID); err != nil {
				rows.Close()
				return nil, err
			}
			crosswalk["team_to_franchise"].(map[string]any)[teamID] = franchiseID
			if _, ok := collector.localFranchiseIDs[franchiseID]; !ok {
				collector.localFranchiseIDs[franchiseID] = struct{}{}
			}
		}
		rows.Close()

		query = fmt.Sprintf(`
			SELECT DISTINCT ON (local_team_id) local_team_id, mlbam_team_id
			FROM team_mlbam_map
			WHERE local_team_id IN (%s)
			ORDER BY local_team_id, season DESC
		`, makePlaceholders(1, len(teamIDs)))
		rows, err = e.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var teamID sql.NullString
			var mlbamID sql.NullInt64
			if err := rows.Scan(&teamID, &mlbamID); err != nil {
				rows.Close()
				return nil, err
			}
			if teamID.Valid && mlbamID.Valid {
				id := int(mlbamID.Int64)
				crosswalk["local_team_to_mlbam"].(map[string]any)[teamID.String] = id
				crosswalk["mlbam_team_to_local"].(map[string]any)[strconv.Itoa(id)] = teamID.String
				collector.mlbamTeamIDs[id] = struct{}{}
			}
		}
		rows.Close()
	}

	franchiseIDs := setToSortedStrings(collector.localFranchiseIDs)
	if len(franchiseIDs) > 0 {
		args := make([]any, len(franchiseIDs))
		for i, id := range franchiseIDs {
			args[i] = id
		}
		query := fmt.Sprintf(`
			SELECT DISTINCT ON (franchise_id) franchise_id, team_id
			FROM team_franchise_map
			WHERE franchise_id IN (%s)
			ORDER BY franchise_id, season DESC
		`, makePlaceholders(1, len(franchiseIDs)))
		rows, err := e.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var franchiseID, teamID string
			if err := rows.Scan(&franchiseID, &teamID); err != nil {
				rows.Close()
				return nil, err
			}
			crosswalk["franchise_to_team"].(map[string]any)[franchiseID] = teamID
		}
		rows.Close()
	}

	mlbamTeamIDs := setToSortedInts(collector.mlbamTeamIDs)
	if len(mlbamTeamIDs) > 0 {
		args := make([]any, len(mlbamTeamIDs))
		for i, id := range mlbamTeamIDs {
			args[i] = id
		}
		query := fmt.Sprintf(`
			SELECT DISTINCT ON (mlbam_team_id) mlbam_team_id, local_team_id
			FROM team_mlbam_map
			WHERE mlbam_team_id IN (%s)
			ORDER BY mlbam_team_id, season DESC
		`, makePlaceholders(1, len(mlbamTeamIDs)))
		rows, err := e.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var mlbamID int
			var teamID sql.NullString
			if err := rows.Scan(&mlbamID, &teamID); err != nil {
				rows.Close()
				return nil, err
			}
			if teamID.Valid && teamID.String != "" {
				crosswalk["mlbam_team_to_local"].(map[string]any)[strconv.Itoa(mlbamID)] = teamID.String
				crosswalk["local_team_to_mlbam"].(map[string]any)[teamID.String] = mlbamID
			}
		}
		rows.Close()
	}

	return crosswalk, nil
}

func (e *responseDetailsEnricher) buildDetails(ctx context.Context, collector *idCollector) (map[string]any, error) {
	teamIDs := setToSortedStrings(collector.localTeamIDs)
	franchiseIDs := setToSortedStrings(collector.localFranchiseIDs)
	playerIDs := setToSortedStrings(collector.localPlayerIDs)
	retoIDs := setToSortedStrings(collector.retroIDs)
	mlbamTeamIDs := setToSortedInts(collector.mlbamTeamIDs)
	mlbamPlayerIDs := setToSortedInts(collector.mlbamPlayerIDs)

	// Add local player IDs found through retro IDs.
	if len(retoIDs) > 0 {
		args := make([]any, len(retoIDs))
		for i, id := range retoIDs {
			args[i] = id
		}
		query := fmt.Sprintf(`SELECT lahman_id FROM player_id_map WHERE retro_id IN (%s)`, makePlaceholders(1, len(retoIDs)))
		rows, err := e.db.QueryContext(ctx, query, args...)
		if err == nil {
			for rows.Next() {
				var lahmanID sql.NullString
				if scanErr := rows.Scan(&lahmanID); scanErr == nil && lahmanID.Valid && lahmanID.String != "" {
					collector.localPlayerIDs[lahmanID.String] = struct{}{}
				}
			}
			rows.Close()
		}
		playerIDs = setToSortedStrings(collector.localPlayerIDs)
	}

	teams, err := e.queryTeamDetails(ctx, teamIDs)
	if err != nil {
		return nil, err
	}
	for _, v := range teams {
		if row, ok := v.(map[string]any); ok {
			if franchiseID, ok := row["franchise_id"].(string); ok && franchiseID != "" {
				collector.localFranchiseIDs[franchiseID] = struct{}{}
			}
		}
	}
	franchiseIDs = setToSortedStrings(collector.localFranchiseIDs)

	franchises, err := e.queryFranchiseDetails(ctx, franchiseIDs)
	if err != nil {
		return nil, err
	}
	players, err := e.queryPlayerDetails(ctx, playerIDs)
	if err != nil {
		return nil, err
	}

	crosswalk, err := e.queryCrosswalkMaps(ctx, collector)
	if err != nil {
		return nil, err
	}

	for _, v := range crosswalk["local_player_to_mlbam"].(map[string]any) {
		if id, ok := toInt(v); ok && id > 0 {
			collector.mlbamPlayerIDs[id] = struct{}{}
		}
	}
	for _, v := range crosswalk["local_team_to_mlbam"].(map[string]any) {
		if id, ok := toInt(v); ok && id > 0 {
			collector.mlbamTeamIDs[id] = struct{}{}
		}
	}
	mlbamTeamIDs = setToSortedInts(collector.mlbamTeamIDs)
	mlbamPlayerIDs = setToSortedInts(collector.mlbamPlayerIDs)

	mlbamTeams, err := e.queryMLBAMTeams(ctx, mlbamTeamIDs)
	if err != nil {
		return nil, err
	}
	mlbamPlayers, err := e.queryMLBAMPlayers(ctx, mlbamPlayerIDs)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"teams":         teams,
		"franchises":    franchises,
		"players":       players,
		"mlbam_teams":   mlbamTeams,
		"mlbam_players": mlbamPlayers,
		"crosswalk":     crosswalk,
	}, nil
}

func (e *responseDetailsEnricher) Enrich(ctx context.Context, payload []byte) ([]byte, bool, error) {
	if e == nil || len(payload) == 0 {
		return payload, false, nil
	}
	var root any
	if err := json.Unmarshal(payload, &root); err != nil {
		return payload, false, nil
	}

	collector := newIDCollector()
	collectIDsFromJSON(root, "", collector)
	details, err := e.buildDetails(ctx, collector)
	if err != nil {
		return payload, false, err
	}

	switch typed := root.(type) {
	case map[string]any:
		meta := map[string]any{}
		if existingMeta, ok := typed["meta"].(map[string]any); ok {
			for k, v := range existingMeta {
				meta[k] = v
			}
		}
		meta["details"] = details
		typed["meta"] = meta
		updated, err := json.Marshal(typed)
		return updated, err == nil, err
	case []any:
		wrapped := map[string]any{
			"data": typed,
			"meta": map[string]any{"details": details},
		}
		updated, err := json.Marshal(wrapped)
		return updated, err == nil, err
	default:
		return payload, false, nil
	}
}
