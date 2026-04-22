package seed

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/core"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
)

const (
	defaultCurrentSeasonSyncTimeout = 20 * time.Second
	currentSeasonSyncAll            = "all"
	currentSeasonSyncStats          = "stats"
	currentSeasonSyncStandings      = "standings"
	currentSeasonSyncSchedule       = "schedule"
	currentSeasonSyncRosters        = "rosters"
)

var currentSeasonSyncTypeSet = map[string]struct{}{
	currentSeasonSyncAll:       {},
	currentSeasonSyncStats:     {},
	currentSeasonSyncStandings: {},
	currentSeasonSyncSchedule:  {},
	currentSeasonSyncRosters:   {},
}

type currentSeasonCrosswalk struct {
	playerByMLBID map[int]string
	teamByMLBID   map[int]currentSeasonTeamCrosswalk
}

type currentSeasonTeamCrosswalk struct {
	TeamID      string
	FranchiseID string
}

type currentSeasonBattingRow struct {
	MLBID     int
	PlayerID  *string
	Season    int
	TeamMLBID int
	TeamID    *string

	G   *int
	PA  *int
	AB  *int
	R   *int
	H   *int
	B2  *int
	B3  *int
	HR  *int
	RBI *int
	SB  *int
	CS  *int
	BB  *int
	SO  *int
	HBP *int
	SF  *int
	SH  *int

	AVG *float64
	OBP *float64
	SLG *float64
	OPS *float64
}

type currentSeasonPitchingRow struct {
	MLBID     int
	PlayerID  *string
	Season    int
	TeamMLBID int
	TeamID    *string

	W   *int
	L   *int
	G   *int
	GS  *int
	SV  *int
	IP  *float64
	H   *int
	R   *int
	ER  *int
	HR  *int
	BB  *int
	SO  *int
	HBP *int

	ERA  *float64
	WHIP *float64
}

type currentSeasonStandingRow struct {
	Season       int
	DivisionID   int
	DivisionName string
	TeamMLBID    int
	TeamID       *string
	FranchiseID  *string

	W       *int
	L       *int
	PCT     *float64
	GB      *string
	WCGB    *string
	Streak  *string
	L10     *string
	RunDiff *int
	RS      *int
	RA      *int
}

type currentSeasonGameRow struct {
	GamePK int
	Season int

	GameDate     string
	Status       string
	AwayMLBID    int
	AwayTeamID   *string
	AwayScore    *int
	HomeMLBID    int
	HomeTeamID   *string
	HomeScore    *int
	Venue        *string
	Innings      *int
	DayNight     *string
	DoubleHeader *string
}

func SyncCurrentSeasonStats(ctx context.Context, database *db.DB, season int, httpClient *http.Client) (int64, error) {
	if database == nil {
		return 0, fmt.Errorf("database is required")
	}
	if season <= 0 {
		return 0, fmt.Errorf("season must be > 0")
	}

	crosswalk, err := loadCurrentSeasonCrosswalk(ctx, database, season)
	if err != nil {
		return 0, err
	}

	hittingPayload, err := fetchCurrentSeasonStats(ctx, httpClient, season, "hitting")
	if err != nil {
		return 0, err
	}
	pitchingPayload, err := fetchCurrentSeasonStats(ctx, httpClient, season, "pitching")
	if err != nil {
		return 0, err
	}

	battingRows, missingBatters, missingBattingTeams := parseCurrentSeasonBattingRows(hittingPayload, season, crosswalk)
	pitchingRows, missingPitchers, missingPitchingTeams := parseCurrentSeasonPitchingRows(pitchingPayload, season, crosswalk)
	logCurrentSeasonCrosswalkGaps("player", mergeMissingIDCounts(missingBatters, missingPitchers))
	logCurrentSeasonCrosswalkGaps("team", mergeMissingIDCounts(missingBattingTeams, missingPitchingTeams))

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin current-season stats tx: %w", err)
	}
	defer tx.Rollback()

	battingRowsWritten, err := upsertCurrentSeasonBattingRows(ctx, tx, battingRows)
	if err != nil {
		return 0, err
	}
	pitchingRowsWritten, err := upsertCurrentSeasonPitchingRows(ctx, tx, pitchingRows)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit current-season stats tx: %w", err)
	}

	totalRows := battingRowsWritten + pitchingRowsWritten
	echo.Infof("Current-season stats sync completed season=%d batting_rows=%d pitching_rows=%d", season, battingRowsWritten, pitchingRowsWritten)
	return totalRows, nil
}

func SyncCurrentSeasonStandings(ctx context.Context, database *db.DB, season int, httpClient *http.Client) (int64, error) {
	if database == nil {
		return 0, fmt.Errorf("database is required")
	}
	if season <= 0 {
		return 0, fmt.Errorf("season must be > 0")
	}

	crosswalk, err := loadCurrentSeasonCrosswalk(ctx, database, season)
	if err != nil {
		return 0, err
	}

	payload, err := fetchCurrentSeasonStandings(ctx, httpClient, season)
	if err != nil {
		return 0, err
	}

	rows, missingTeams := parseCurrentSeasonStandingRows(payload, season, crosswalk)
	logCurrentSeasonCrosswalkGaps("team", missingTeams)

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin current-season standings tx: %w", err)
	}
	defer tx.Rollback()

	written, err := upsertCurrentSeasonStandingRows(ctx, tx, rows)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit current-season standings tx: %w", err)
	}

	echo.Infof("Current-season standings sync completed season=%d rows=%d", season, written)
	return written, nil
}

func SyncCurrentSeasonSchedule(ctx context.Context, database *db.DB, season int, httpClient *http.Client) (int64, error) {
	if database == nil {
		return 0, fmt.Errorf("database is required")
	}
	if season <= 0 {
		return 0, fmt.Errorf("season must be > 0")
	}

	crosswalk, err := loadCurrentSeasonCrosswalk(ctx, database, season)
	if err != nil {
		return 0, err
	}

	payload, err := fetchCurrentSeasonSchedule(ctx, httpClient, season)
	if err != nil {
		return 0, err
	}

	rows, missingTeams := parseCurrentSeasonGameRows(payload, season, crosswalk)
	logCurrentSeasonCrosswalkGaps("team", missingTeams)

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin current-season schedule tx: %w", err)
	}
	defer tx.Rollback()

	written, err := upsertCurrentSeasonGameRows(ctx, tx, rows)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit current-season schedule tx: %w", err)
	}

	echo.Infof("Current-season schedule sync completed season=%d rows=%d", season, written)
	return written, nil
}

func SyncCurrentSeasonRosters(ctx context.Context, database *db.DB, season int, httpClient *http.Client) (int64, error) {
	if database == nil {
		return 0, fmt.Errorf("database is required")
	}
	if season <= 0 {
		return 0, fmt.Errorf("season must be > 0")
	}

	crosswalk, err := loadCurrentSeasonCrosswalk(ctx, database, season)
	if err != nil {
		return 0, err
	}

	mlbTeams, err := fetchMLBTeamsForSeason(ctx, season)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch MLB teams for roster sync season=%d: %w", season, err)
	}

	rowsScanned := int64(0)
	missingPlayers := map[int]int{}
	missingTeams := map[int]int{}
	for _, team := range mlbTeams.Teams {
		if team.ID <= 0 {
			echo.Infof("Skipping malformed roster team entry season=%d missing_team_id", season)
			continue
		}
		if _, ok := crosswalk.teamByMLBID[team.ID]; !ok {
			missingTeams[team.ID]++
		}

		rosterPayload, err := fetchCurrentSeasonTeamRoster(ctx, httpClient, season, team.ID)
		if err != nil {
			return rowsScanned, fmt.Errorf("failed fetching team roster team_mlb_id=%d season=%d: %w", team.ID, season, err)
		}

		for _, entry := range rosterPayload.Roster {
			if entry.Person == nil || entry.Person.ID <= 0 {
				echo.Infof("Skipping malformed roster entry season=%d team_mlb_id=%d", season, team.ID)
				continue
			}
			rowsScanned++
			if _, ok := crosswalk.playerByMLBID[entry.Person.ID]; !ok {
				missingPlayers[entry.Person.ID]++
			}
		}
	}

	logCurrentSeasonCrosswalkGaps("player", missingPlayers)
	logCurrentSeasonCrosswalkGaps("team", missingTeams)
	echo.Infof("Current-season roster sync completed season=%d scanned_active_players=%d", season, rowsScanned)
	return rowsScanned, nil
}

func executeCurrentSeasonSync(ctx context.Context, database *db.DB, job *db.ETLJob, httpClient *http.Client) (int64, error) {
	syncType, err := normalizeCurrentSeasonSyncType(stringFromAny(job.Scope["sync_type"]))
	if err != nil {
		return 0, err
	}
	season := resolveCurrentSeasonFromJob(job)

	echo.Infof("current-season-sync started job=%d sync_type=%s season=%d", job.ID, syncType, season)

	rows := int64(0)
	switch syncType {
	case currentSeasonSyncAll:
		written, err := SyncCurrentSeasonStats(ctx, database, season, httpClient)
		rows += written
		if err != nil {
			return rows, err
		}
		written, err = SyncCurrentSeasonStandings(ctx, database, season, httpClient)
		rows += written
		if err != nil {
			return rows, err
		}
		written, err = SyncCurrentSeasonSchedule(ctx, database, season, httpClient)
		rows += written
		if err != nil {
			return rows, err
		}
		written, err = SyncCurrentSeasonRosters(ctx, database, season, httpClient)
		rows += written
		if err != nil {
			return rows, err
		}
	case currentSeasonSyncStats:
		written, err := SyncCurrentSeasonStats(ctx, database, season, httpClient)
		rows += written
		if err != nil {
			return rows, err
		}
	case currentSeasonSyncStandings:
		written, err := SyncCurrentSeasonStandings(ctx, database, season, httpClient)
		rows += written
		if err != nil {
			return rows, err
		}
	case currentSeasonSyncSchedule:
		written, err := SyncCurrentSeasonSchedule(ctx, database, season, httpClient)
		rows += written
		if err != nil {
			return rows, err
		}
	case currentSeasonSyncRosters:
		written, err := SyncCurrentSeasonRosters(ctx, database, season, httpClient)
		rows += written
		if err != nil {
			return rows, err
		}
	default:
		return 0, fmt.Errorf("unsupported current-season sync_type %q", syncType)
	}

	echo.Successf("✓ current-season-sync completed job=%d sync_type=%s season=%d rows=%d", job.ID, syncType, season, rows)
	return rows, nil
}

func normalizeCurrentSeasonSyncType(raw string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return currentSeasonSyncAll, nil
	}
	if _, ok := currentSeasonSyncTypeSet[value]; ok {
		return value, nil
	}
	return "", fmt.Errorf("invalid sync_type %q (allowed: all,stats,standings,schedule,rosters)", raw)
}

func resolveCurrentSeasonFromJob(job *db.ETLJob) int {
	season := intFromAny(job.Scope["season"])
	if season > 0 {
		return season
	}
	years, err := intSliceFromAny(job.Scope["years"])
	if err == nil && len(years) > 0 && years[0] > 0 {
		return years[0]
	}
	return time.Now().Year()
}

func classifyCurrentSeasonSyncFailure(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return "cancelled", false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return "network", true
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "invalid sync_type") ||
		strings.Contains(msg, "malformed") ||
		strings.Contains(msg, "failed to decode") ||
		strings.Contains(msg, "response shape") {
		return "validation", false
	}

	if strings.Contains(msg, "http") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "tls") ||
		strings.Contains(msg, "temporary") ||
		strings.Contains(msg, "status code") {
		return "network", true
	}

	return "db_write", false
}

func parseCurrentSeasonBattingRows(
	payload core.MLBStatsResponse,
	defaultSeason int,
	crosswalk currentSeasonCrosswalk,
) ([]currentSeasonBattingRow, map[int]int, map[int]int) {
	rows := make([]currentSeasonBattingRow, 0, 2048)
	missingPlayers := map[int]int{}
	missingTeams := map[int]int{}
	for _, stat := range payload.Stats {
		for _, split := range stat.Splits {
			row, ok, reason := buildCurrentSeasonBattingRow(split, defaultSeason, crosswalk)
			if !ok {
				echo.Infof("Skipping malformed batting split season=%d reason=%s", defaultSeason, reason)
				continue
			}
			if row.PlayerID == nil {
				missingPlayers[row.MLBID]++
			}
			if row.TeamID == nil {
				missingTeams[row.TeamMLBID]++
			}
			rows = append(rows, row)
		}
	}
	return rows, missingPlayers, missingTeams
}

func parseCurrentSeasonPitchingRows(
	payload core.MLBStatsResponse,
	defaultSeason int,
	crosswalk currentSeasonCrosswalk,
) ([]currentSeasonPitchingRow, map[int]int, map[int]int) {
	rows := make([]currentSeasonPitchingRow, 0, 2048)
	missingPlayers := map[int]int{}
	missingTeams := map[int]int{}
	for _, stat := range payload.Stats {
		for _, split := range stat.Splits {
			row, ok, reason := buildCurrentSeasonPitchingRow(split, defaultSeason, crosswalk)
			if !ok {
				echo.Infof("Skipping malformed pitching split season=%d reason=%s", defaultSeason, reason)
				continue
			}
			if row.PlayerID == nil {
				missingPlayers[row.MLBID]++
			}
			if row.TeamID == nil {
				missingTeams[row.TeamMLBID]++
			}
			rows = append(rows, row)
		}
	}
	return rows, missingPlayers, missingTeams
}

func parseCurrentSeasonStandingRows(
	payload core.MLBStandingsResponse,
	defaultSeason int,
	crosswalk currentSeasonCrosswalk,
) ([]currentSeasonStandingRow, map[int]int) {
	rows := make([]currentSeasonStandingRow, 0, 64)
	missingTeams := map[int]int{}
	for _, record := range payload.Records {
		for _, teamRecord := range record.TeamRecords {
			row, ok, reason := buildCurrentSeasonStandingRow(record, teamRecord, defaultSeason, crosswalk)
			if !ok {
				echo.Infof("Skipping malformed standings row season=%d reason=%s", defaultSeason, reason)
				continue
			}
			if row.TeamID == nil || row.FranchiseID == nil {
				missingTeams[row.TeamMLBID]++
			}
			rows = append(rows, row)
		}
	}
	return rows, missingTeams
}

func parseCurrentSeasonGameRows(
	payload core.MLBScheduleResponse,
	defaultSeason int,
	crosswalk currentSeasonCrosswalk,
) ([]currentSeasonGameRow, map[int]int) {
	rows := make([]currentSeasonGameRow, 0, payload.TotalGames)
	missingTeams := map[int]int{}
	for _, dateBucket := range payload.Dates {
		for _, game := range dateBucket.Games {
			row, ok, reason := buildCurrentSeasonGameRow(dateBucket, game, defaultSeason, crosswalk)
			if !ok {
				echo.Infof("Skipping malformed game row season=%d game_pk=%d reason=%s", defaultSeason, game.GamePk, reason)
				continue
			}
			if row.AwayTeamID == nil {
				missingTeams[row.AwayMLBID]++
			}
			if row.HomeTeamID == nil {
				missingTeams[row.HomeMLBID]++
			}
			rows = append(rows, row)
		}
	}
	return rows, missingTeams
}

func buildCurrentSeasonBattingRow(
	split core.MLBStatSplit,
	defaultSeason int,
	crosswalk currentSeasonCrosswalk,
) (currentSeasonBattingRow, bool, string) {
	if split.Player == nil || split.Player.ID <= 0 {
		return currentSeasonBattingRow{}, false, "missing player.id"
	}
	if split.Team == nil || split.Team.ID <= 0 {
		return currentSeasonBattingRow{}, false, "missing team.id"
	}

	season := parseSeasonValue(split.Season, defaultSeason)
	if season <= 0 {
		return currentSeasonBattingRow{}, false, "missing season"
	}

	playerID := optionalLookup(crosswalk.playerByMLBID, split.Player.ID)
	teamMap := crosswalk.teamByMLBID[split.Team.ID]
	teamID := optionalString(teamMap.TeamID)

	row := currentSeasonBattingRow{
		MLBID:     split.Player.ID,
		PlayerID:  playerID,
		Season:    season,
		TeamMLBID: split.Team.ID,
		TeamID:    teamID,
		G:         parseStatInt(split.Stat.GamesPlayed),
		PA:        parseStatInt(split.Stat.PlateAppearances),
		AB:        parseStatInt(split.Stat.AtBats),
		R:         parseStatInt(split.Stat.Runs),
		H:         parseStatInt(split.Stat.Hits),
		B2:        parseStatInt(split.Stat.Doubles),
		B3:        parseStatInt(split.Stat.Triples),
		HR:        parseStatInt(split.Stat.HomeRuns),
		RBI:       parseStatInt(split.Stat.RBI),
		SB:        parseStatInt(split.Stat.StolenBases),
		CS:        parseStatInt(split.Stat.CaughtStealing),
		BB:        parseStatInt(split.Stat.BaseOnBalls),
		SO:        parseStatInt(split.Stat.StrikeOuts),
		HBP:       parseStatInt(split.Stat.HitByPitch),
		SF:        parseStatInt(split.Stat.SacFlies),
		SH:        parseStatInt(split.Stat.SacBunts),
		AVG:       parseStatFloat(split.Stat.Avg),
		OBP:       parseStatFloat(split.Stat.OBP),
		SLG:       parseStatFloat(split.Stat.SLG),
		OPS:       parseStatFloat(split.Stat.OPS),
	}
	return row, true, ""
}

func buildCurrentSeasonPitchingRow(
	split core.MLBStatSplit,
	defaultSeason int,
	crosswalk currentSeasonCrosswalk,
) (currentSeasonPitchingRow, bool, string) {
	if split.Player == nil || split.Player.ID <= 0 {
		return currentSeasonPitchingRow{}, false, "missing player.id"
	}
	if split.Team == nil || split.Team.ID <= 0 {
		return currentSeasonPitchingRow{}, false, "missing team.id"
	}

	season := parseSeasonValue(split.Season, defaultSeason)
	if season <= 0 {
		return currentSeasonPitchingRow{}, false, "missing season"
	}

	playerID := optionalLookup(crosswalk.playerByMLBID, split.Player.ID)
	teamMap := crosswalk.teamByMLBID[split.Team.ID]
	teamID := optionalString(teamMap.TeamID)

	row := currentSeasonPitchingRow{
		MLBID:     split.Player.ID,
		PlayerID:  playerID,
		Season:    season,
		TeamMLBID: split.Team.ID,
		TeamID:    teamID,
		W:         parseStatInt(split.Stat.Wins),
		L:         parseStatInt(split.Stat.Losses),
		G:         parseStatInt(split.Stat.GamesPlayed),
		GS:        parseStatInt(split.Stat.GamesStarted),
		SV:        parseStatInt(split.Stat.Saves),
		IP:        parseStatFloat(split.Stat.InningsPitched),
		H:         parseStatInt(split.Stat.Hits),
		R:         parseStatInt(split.Stat.Runs),
		ER:        parseStatInt(split.Stat.EarnedRuns),
		HR:        parseStatInt(split.Stat.HomeRuns),
		BB:        parseStatInt(split.Stat.BaseOnBalls),
		SO:        parseStatInt(split.Stat.StrikeOuts),
		HBP:       parseStatInt(split.Stat.HitByPitch),
		ERA:       parseStatFloat(split.Stat.ERA),
		WHIP:      parseStatFloat(split.Stat.WHIP),
	}
	return row, true, ""
}

func buildCurrentSeasonStandingRow(
	record core.MLBStandingsRecord,
	teamRecord core.MLBTeamRecord,
	defaultSeason int,
	crosswalk currentSeasonCrosswalk,
) (currentSeasonStandingRow, bool, string) {
	if record.Division == nil || record.Division.ID <= 0 {
		return currentSeasonStandingRow{}, false, "missing division"
	}
	if teamRecord.Team == nil || teamRecord.Team.ID <= 0 {
		return currentSeasonStandingRow{}, false, "missing team.id"
	}

	season := parseSeasonValue(record.Season, defaultSeason)
	if season <= 0 {
		return currentSeasonStandingRow{}, false, "missing season"
	}

	teamMap := crosswalk.teamByMLBID[teamRecord.Team.ID]
	streakCode := ""
	if teamRecord.Streak != nil {
		streakCode = teamRecord.Streak.StreakCode
	}
	streak := optionalString(streakCode)
	lastTen := optionalString(extractLastTenRecord(teamRecord))

	row := currentSeasonStandingRow{
		Season:       season,
		DivisionID:   record.Division.ID,
		DivisionName: strings.TrimSpace(record.Division.Name),
		TeamMLBID:    teamRecord.Team.ID,
		TeamID:       optionalString(teamMap.TeamID),
		FranchiseID:  optionalString(teamMap.FranchiseID),
		W:            intPointer(teamRecord.Wins),
		L:            intPointer(teamRecord.Losses),
		PCT:          parseDecimalString(teamRecord.WinningPercentage),
		GB:           optionalString(teamRecord.GamesBack),
		WCGB:         optionalString(teamRecord.WildCardGamesBack),
		Streak:       streak,
		L10:          lastTen,
		RunDiff:      parseSignedIntString(teamRecord.RunDifferential),
		RS:           intPointer(teamRecord.RunsScored),
		RA:           intPointer(teamRecord.RunsAllowed),
	}
	if row.DivisionName == "" {
		row.DivisionName = fmt.Sprintf("Division %d", row.DivisionID)
	}
	return row, true, ""
}

func buildCurrentSeasonGameRow(
	dateBucket core.MLBScheduleDate,
	game core.MLBGame,
	defaultSeason int,
	crosswalk currentSeasonCrosswalk,
) (currentSeasonGameRow, bool, string) {
	if game.GamePk <= 0 {
		return currentSeasonGameRow{}, false, "missing game_pk"
	}
	if game.Teams == nil || game.Teams.Away == nil || game.Teams.Home == nil {
		return currentSeasonGameRow{}, false, "missing game teams"
	}
	if game.Teams.Away.Team == nil || game.Teams.Home.Team == nil {
		return currentSeasonGameRow{}, false, "missing away/home team ids"
	}

	season := parseSeasonValue(game.Season, defaultSeason)
	if season <= 0 {
		season = defaultSeason
	}

	awayMLBID := game.Teams.Away.Team.ID
	homeMLBID := game.Teams.Home.Team.ID
	if awayMLBID <= 0 || homeMLBID <= 0 {
		return currentSeasonGameRow{}, false, "invalid away/home team ids"
	}

	gameDate := strings.TrimSpace(game.OfficialDate)
	if gameDate == "" {
		gameDate = strings.TrimSpace(dateBucket.Date)
	}
	if gameDate == "" && len(game.GameDate) >= 10 {
		gameDate = game.GameDate[:10]
	}
	if gameDate == "" {
		return currentSeasonGameRow{}, false, "missing game date"
	}

	status := ""
	if game.Status != nil {
		status = strings.TrimSpace(game.Status.DetailedState)
		if status == "" {
			status = strings.TrimSpace(game.Status.AbstractGameState)
		}
	}
	if status == "" {
		status = "Unknown"
	}

	awayTeamID := optionalString(crosswalk.teamByMLBID[awayMLBID].TeamID)
	homeTeamID := optionalString(crosswalk.teamByMLBID[homeMLBID].TeamID)
	venueName := ""
	if game.Venue != nil {
		venueName = game.Venue.Name
	}

	row := currentSeasonGameRow{
		GamePK:       game.GamePk,
		Season:       season,
		GameDate:     gameDate,
		Status:       status,
		AwayMLBID:    awayMLBID,
		AwayTeamID:   awayTeamID,
		AwayScore:    parseGameScore(game.Teams.Away, status),
		HomeMLBID:    homeMLBID,
		HomeTeamID:   homeTeamID,
		HomeScore:    parseGameScore(game.Teams.Home, status),
		Venue:        optionalString(venueName),
		Innings:      parseInnings(game),
		DayNight:     optionalString(game.DayNight),
		DoubleHeader: optionalString(game.DoubleHeader),
	}
	return row, true, ""
}

func parseGameScore(team *core.MLBGameTeamWrapper, status string) *int {
	if team == nil {
		return nil
	}
	score := team.Score
	if score == 0 && looksLikeUnplayedStatus(status) {
		return nil
	}
	return intPointer(score)
}

func looksLikeUnplayedStatus(status string) bool {
	value := strings.ToLower(strings.TrimSpace(status))
	return strings.Contains(value, "scheduled") ||
		strings.Contains(value, "pre-game") ||
		strings.Contains(value, "pre game") ||
		strings.Contains(value, "delayed start")
}

func parseInnings(game core.MLBGame) *int {
	if game.Linescore != nil && game.Linescore.CurrentInning > 0 {
		return intPointer(game.Linescore.CurrentInning)
	}
	if game.ScheduledInnings > 0 {
		return intPointer(game.ScheduledInnings)
	}
	return nil
}

func extractLastTenRecord(teamRecord core.MLBTeamRecord) string {
	for _, split := range teamRecord.LastTenRecords {
		if strings.EqualFold(strings.TrimSpace(split.Type), "lastTen") {
			return fmt.Sprintf("%d-%d", split.Wins, split.Losses)
		}
	}
	if teamRecord.Records != nil {
		for _, split := range teamRecord.Records.SplitRecords {
			if strings.EqualFold(strings.TrimSpace(split.Type), "lastTen") {
				return fmt.Sprintf("%d-%d", split.Wins, split.Losses)
			}
		}
	}
	return ""
}

func parseSeasonValue(raw string, fallback int) int {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func parseStatInt(raw json.RawMessage) *int {
	if len(raw) == 0 {
		return nil
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return intPointer(n)
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return intPointer(int(f))
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return nil
	}
	text = normalizeNumericString(text)
	if text == "" {
		return nil
	}
	if n, err := strconv.Atoi(text); err == nil {
		return intPointer(n)
	}
	if f, err := strconv.ParseFloat(text, 64); err == nil {
		return intPointer(int(f))
	}
	return nil
}

func parseStatFloat(raw json.RawMessage) *float64 {
	if len(raw) == 0 {
		return nil
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return floatPointer(f)
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return nil
	}
	text = normalizeNumericString(text)
	if text == "" {
		return nil
	}
	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil
	}
	return floatPointer(f)
}

func parseDecimalString(raw string) *float64 {
	value := normalizeNumericString(raw)
	if value == "" {
		return nil
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	return floatPointer(f)
}

func parseSignedIntString(raw string) *int {
	value := normalizeNumericString(raw)
	if value == "" {
		return nil
	}
	n, err := strconv.Atoi(value)
	if err == nil {
		return intPointer(n)
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	return intPointer(int(f))
}

func normalizeNumericString(raw string) string {
	value := strings.TrimSpace(raw)
	switch value {
	case "", "-", "--", "N/A", "n/a":
		return ""
	}
	if strings.HasPrefix(value, ".") {
		return "0" + value
	}
	return value
}

func optionalLookup(values map[int]string, key int) *string {
	if len(values) == 0 || key <= 0 {
		return nil
	}
	value, ok := values[key]
	if !ok {
		return nil
	}
	return optionalString(value)
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func intPointer(value int) *int { return &value }

func floatPointer(value float64) *float64 { return &value }

func mergeMissingIDCounts(a, b map[int]int) map[int]int {
	if len(a) == 0 && len(b) == 0 {
		return map[int]int{}
	}
	out := make(map[int]int, len(a)+len(b))
	for id, count := range a {
		out[id] += count
	}
	for id, count := range b {
		out[id] += count
	}
	return out
}

func logCurrentSeasonCrosswalkGaps(kind string, values map[int]int) {
	if len(values) == 0 {
		return
	}
	ids := make([]int, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	limit := 12
	if len(ids) < limit {
		limit = len(ids)
	}
	sample := ids[:limit]
	echo.Infof("Current-season crosswalk gaps kind=%s unmatched_ids=%d sample=%v", kind, len(ids), sample)
}

func loadCurrentSeasonCrosswalk(ctx context.Context, database *db.DB, season int) (currentSeasonCrosswalk, error) {
	playerByMLBID := make(map[int]string, 16000)
	playerRows, err := database.QueryContext(ctx, `
		SELECT mlbam_id, lahman_id
		FROM player_mlbam_map
		WHERE mlbam_id IS NOT NULL
	`)
	if err != nil {
		return currentSeasonCrosswalk{}, fmt.Errorf("failed loading player crosswalk map: %w", err)
	}
	defer playerRows.Close()
	for playerRows.Next() {
		var mlbID int
		var lahmanID sql.NullString
		if err := playerRows.Scan(&mlbID, &lahmanID); err != nil {
			return currentSeasonCrosswalk{}, fmt.Errorf("failed scanning player crosswalk row: %w", err)
		}
		if mlbID > 0 && lahmanID.Valid && strings.TrimSpace(lahmanID.String) != "" {
			playerByMLBID[mlbID] = strings.TrimSpace(lahmanID.String)
		}
	}
	if err := playerRows.Err(); err != nil {
		return currentSeasonCrosswalk{}, fmt.Errorf("failed iterating player crosswalk rows: %w", err)
	}

	teamByMLBID := make(map[int]currentSeasonTeamCrosswalk, 64)
	teamRows, err := database.QueryContext(ctx, `
		SELECT DISTINCT ON (mlbam_team_id)
			mlbam_team_id,
			local_team_id,
			local_franchise_id
		FROM team_mlbam_map
		WHERE season <= $1
		ORDER BY
			mlbam_team_id,
			CASE WHEN season = $1 THEN 0 ELSE 1 END,
			season DESC
	`, season)
	if err != nil {
		return currentSeasonCrosswalk{}, fmt.Errorf("failed loading team crosswalk map: %w", err)
	}
	defer teamRows.Close()
	for teamRows.Next() {
		var mlbTeamID int
		var teamID sql.NullString
		var franchiseID sql.NullString
		if err := teamRows.Scan(&mlbTeamID, &teamID, &franchiseID); err != nil {
			return currentSeasonCrosswalk{}, fmt.Errorf("failed scanning team crosswalk row: %w", err)
		}
		if mlbTeamID <= 0 {
			continue
		}
		teamByMLBID[mlbTeamID] = currentSeasonTeamCrosswalk{
			TeamID:      strings.TrimSpace(teamID.String),
			FranchiseID: strings.TrimSpace(franchiseID.String),
		}
	}
	if err := teamRows.Err(); err != nil {
		return currentSeasonCrosswalk{}, fmt.Errorf("failed iterating team crosswalk rows: %w", err)
	}

	return currentSeasonCrosswalk{
		playerByMLBID: playerByMLBID,
		teamByMLBID:   teamByMLBID,
	}, nil
}

func fetchCurrentSeasonStats(ctx context.Context, httpClient *http.Client, season int, group string) (core.MLBStatsResponse, error) {
	payload := core.MLBStatsResponse{}
	err := fetchCurrentSeasonAPIJSON(ctx, httpClient, []string{"v1", "stats"}, map[string]string{
		"stats":      "season",
		"group":      group,
		"season":     strconv.Itoa(season),
		"playerPool": "all",
		"sportId":    "1",
	}, &payload)
	if err != nil {
		return core.MLBStatsResponse{}, err
	}
	if len(payload.Stats) == 0 {
		echo.Infof("MLB stats payload returned no stat groups season=%d group=%s", season, group)
	}
	return payload, nil
}

func fetchCurrentSeasonStandings(ctx context.Context, httpClient *http.Client, season int) (core.MLBStandingsResponse, error) {
	payload := core.MLBStandingsResponse{}
	err := fetchCurrentSeasonAPIJSON(ctx, httpClient, []string{"v1", "standings"}, map[string]string{
		"season":         strconv.Itoa(season),
		"standingsTypes": "regularSeason",
		"sportId":        "1",
	}, &payload)
	if err != nil {
		return core.MLBStandingsResponse{}, err
	}
	return payload, nil
}

func fetchCurrentSeasonSchedule(ctx context.Context, httpClient *http.Client, season int) (core.MLBScheduleResponse, error) {
	payload := core.MLBScheduleResponse{}
	err := fetchCurrentSeasonAPIJSON(ctx, httpClient, []string{"v1", "schedule"}, map[string]string{
		"season":  strconv.Itoa(season),
		"sportId": "1",
		"hydrate": "linescore,team",
	}, &payload)
	if err != nil {
		return core.MLBScheduleResponse{}, err
	}
	return payload, nil
}

func fetchCurrentSeasonTeamRoster(ctx context.Context, httpClient *http.Client, season int, teamMLBID int) (core.MLBRosterResponse, error) {
	payload := core.MLBRosterResponse{}
	err := fetchCurrentSeasonAPIJSON(ctx, httpClient, []string{"v1", "teams", strconv.Itoa(teamMLBID), "roster"}, map[string]string{
		"rosterType": "active",
		"season":     strconv.Itoa(season),
		"sportId":    "1",
	}, &payload)
	if err != nil {
		return core.MLBRosterResponse{}, err
	}
	return payload, nil
}

func fetchCurrentSeasonAPIJSON(
	ctx context.Context,
	httpClient *http.Client,
	pathParts []string,
	query map[string]string,
	dest any,
) error {
	client := httpClient
	if client == nil {
		client = &http.Client{Timeout: defaultCurrentSeasonSyncTimeout}
	}

	target, err := url.JoinPath(mlbStatsAPIBaseURL, pathParts...)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	for key, value := range query {
		if strings.TrimSpace(value) == "" {
			continue
		}
		q.Set(key, value)
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", "Stormlight-Baseball-ETL/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		return fmt.Errorf("mlb api request failed status code %d url=%s body=%s", resp.StatusCode, req.URL.String(), strings.TrimSpace(string(body)))
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("failed to decode MLB API response url=%s: %w", req.URL.String(), err)
	}
	return nil
}

func upsertCurrentSeasonBattingRows(ctx context.Context, tx *sql.Tx, rows []currentSeasonBattingRow) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO current_season.batting (
			mlb_id, player_id, season, team_mlb_id, team_id,
			g, pa, ab, r, h, "2b", "3b", hr, rbi, sb, cs, bb, so, hbp, sf, sh,
			avg, obp, slg, ops, fetched_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21,
			$22, $23, $24, $25, NOW()
		)
		ON CONFLICT (mlb_id, season, team_mlb_id) DO UPDATE SET
			player_id = EXCLUDED.player_id,
			team_id = EXCLUDED.team_id,
			g = EXCLUDED.g,
			pa = EXCLUDED.pa,
			ab = EXCLUDED.ab,
			r = EXCLUDED.r,
			h = EXCLUDED.h,
			"2b" = EXCLUDED."2b",
			"3b" = EXCLUDED."3b",
			hr = EXCLUDED.hr,
			rbi = EXCLUDED.rbi,
			sb = EXCLUDED.sb,
			cs = EXCLUDED.cs,
			bb = EXCLUDED.bb,
			so = EXCLUDED.so,
			hbp = EXCLUDED.hbp,
			sf = EXCLUDED.sf,
			sh = EXCLUDED.sh,
			avg = EXCLUDED.avg,
			obp = EXCLUDED.obp,
			slg = EXCLUDED.slg,
			ops = EXCLUDED.ops,
			fetched_at = NOW()
	`)
	if err != nil {
		return 0, fmt.Errorf("failed preparing current-season batting upsert statement: %w", err)
	}
	defer stmt.Close()

	written := int64(0)
	for _, row := range rows {
		result, err := stmt.ExecContext(ctx,
			row.MLBID,
			stringValue(row.PlayerID),
			row.Season,
			row.TeamMLBID,
			stringValue(row.TeamID),
			intValue(row.G),
			intValue(row.PA),
			intValue(row.AB),
			intValue(row.R),
			intValue(row.H),
			intValue(row.B2),
			intValue(row.B3),
			intValue(row.HR),
			intValue(row.RBI),
			intValue(row.SB),
			intValue(row.CS),
			intValue(row.BB),
			intValue(row.SO),
			intValue(row.HBP),
			intValue(row.SF),
			intValue(row.SH),
			floatValue(row.AVG),
			floatValue(row.OBP),
			floatValue(row.SLG),
			floatValue(row.OPS),
		)
		if err != nil {
			return written, fmt.Errorf("failed upserting current-season batting row mlb_id=%d team_mlb_id=%d season=%d: %w", row.MLBID, row.TeamMLBID, row.Season, err)
		}
		affected, err := result.RowsAffected()
		if err == nil {
			written += affected
		} else {
			written++
		}
	}
	return written, nil
}

func upsertCurrentSeasonPitchingRows(ctx context.Context, tx *sql.Tx, rows []currentSeasonPitchingRow) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO current_season.pitching (
			mlb_id, player_id, season, team_mlb_id, team_id,
			w, l, g, gs, sv, ip, h, r, er, hr, bb, so, hbp, era, whip, fetched_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, NOW()
		)
		ON CONFLICT (mlb_id, season, team_mlb_id) DO UPDATE SET
			player_id = EXCLUDED.player_id,
			team_id = EXCLUDED.team_id,
			w = EXCLUDED.w,
			l = EXCLUDED.l,
			g = EXCLUDED.g,
			gs = EXCLUDED.gs,
			sv = EXCLUDED.sv,
			ip = EXCLUDED.ip,
			h = EXCLUDED.h,
			r = EXCLUDED.r,
			er = EXCLUDED.er,
			hr = EXCLUDED.hr,
			bb = EXCLUDED.bb,
			so = EXCLUDED.so,
			hbp = EXCLUDED.hbp,
			era = EXCLUDED.era,
			whip = EXCLUDED.whip,
			fetched_at = NOW()
	`)
	if err != nil {
		return 0, fmt.Errorf("failed preparing current-season pitching upsert statement: %w", err)
	}
	defer stmt.Close()

	written := int64(0)
	for _, row := range rows {
		result, err := stmt.ExecContext(ctx,
			row.MLBID,
			stringValue(row.PlayerID),
			row.Season,
			row.TeamMLBID,
			stringValue(row.TeamID),
			intValue(row.W),
			intValue(row.L),
			intValue(row.G),
			intValue(row.GS),
			intValue(row.SV),
			floatValue(row.IP),
			intValue(row.H),
			intValue(row.R),
			intValue(row.ER),
			intValue(row.HR),
			intValue(row.BB),
			intValue(row.SO),
			intValue(row.HBP),
			floatValue(row.ERA),
			floatValue(row.WHIP),
		)
		if err != nil {
			return written, fmt.Errorf("failed upserting current-season pitching row mlb_id=%d team_mlb_id=%d season=%d: %w", row.MLBID, row.TeamMLBID, row.Season, err)
		}
		affected, err := result.RowsAffected()
		if err == nil {
			written += affected
		} else {
			written++
		}
	}
	return written, nil
}

func upsertCurrentSeasonStandingRows(ctx context.Context, tx *sql.Tx, rows []currentSeasonStandingRow) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO current_season.standings (
			season, division_id, division_name, team_mlb_id, team_id, franchise_id,
			w, l, pct, gb, wc_gb, streak, l10, run_diff, rs, ra, fetched_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW()
		)
		ON CONFLICT (season, team_mlb_id) DO UPDATE SET
			division_id = EXCLUDED.division_id,
			division_name = EXCLUDED.division_name,
			team_id = EXCLUDED.team_id,
			franchise_id = EXCLUDED.franchise_id,
			w = EXCLUDED.w,
			l = EXCLUDED.l,
			pct = EXCLUDED.pct,
			gb = EXCLUDED.gb,
			wc_gb = EXCLUDED.wc_gb,
			streak = EXCLUDED.streak,
			l10 = EXCLUDED.l10,
			run_diff = EXCLUDED.run_diff,
			rs = EXCLUDED.rs,
			ra = EXCLUDED.ra,
			fetched_at = NOW()
	`)
	if err != nil {
		return 0, fmt.Errorf("failed preparing current-season standings upsert statement: %w", err)
	}
	defer stmt.Close()

	written := int64(0)
	for _, row := range rows {
		result, err := stmt.ExecContext(ctx,
			row.Season,
			row.DivisionID,
			row.DivisionName,
			row.TeamMLBID,
			stringValue(row.TeamID),
			stringValue(row.FranchiseID),
			intValue(row.W),
			intValue(row.L),
			floatValue(row.PCT),
			stringValue(row.GB),
			stringValue(row.WCGB),
			stringValue(row.Streak),
			stringValue(row.L10),
			intValue(row.RunDiff),
			intValue(row.RS),
			intValue(row.RA),
		)
		if err != nil {
			return written, fmt.Errorf("failed upserting current-season standings row team_mlb_id=%d season=%d: %w", row.TeamMLBID, row.Season, err)
		}
		affected, err := result.RowsAffected()
		if err == nil {
			written += affected
		} else {
			written++
		}
	}
	return written, nil
}

func upsertCurrentSeasonGameRows(ctx context.Context, tx *sql.Tx, rows []currentSeasonGameRow) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO current_season.games (
			game_pk, season, game_date, status,
			away_mlb_id, away_team_id, away_score,
			home_mlb_id, home_team_id, home_score,
			venue, innings, day_night, doubleheader, fetched_at
		) VALUES (
			$1, $2, $3::date, $4,
			$5, $6, $7,
			$8, $9, $10,
			$11, $12, $13, $14, NOW()
		)
		ON CONFLICT (game_pk) DO UPDATE SET
			season = EXCLUDED.season,
			game_date = EXCLUDED.game_date,
			status = EXCLUDED.status,
			away_mlb_id = EXCLUDED.away_mlb_id,
			away_team_id = EXCLUDED.away_team_id,
			away_score = EXCLUDED.away_score,
			home_mlb_id = EXCLUDED.home_mlb_id,
			home_team_id = EXCLUDED.home_team_id,
			home_score = EXCLUDED.home_score,
			venue = EXCLUDED.venue,
			innings = EXCLUDED.innings,
			day_night = EXCLUDED.day_night,
			doubleheader = EXCLUDED.doubleheader,
			fetched_at = NOW()
	`)
	if err != nil {
		return 0, fmt.Errorf("failed preparing current-season games upsert statement: %w", err)
	}
	defer stmt.Close()

	written := int64(0)
	for _, row := range rows {
		result, err := stmt.ExecContext(ctx,
			row.GamePK,
			row.Season,
			row.GameDate,
			row.Status,
			row.AwayMLBID,
			stringValue(row.AwayTeamID),
			intValue(row.AwayScore),
			row.HomeMLBID,
			stringValue(row.HomeTeamID),
			intValue(row.HomeScore),
			stringValue(row.Venue),
			intValue(row.Innings),
			stringValue(row.DayNight),
			stringValue(row.DoubleHeader),
		)
		if err != nil {
			return written, fmt.Errorf("failed upserting current-season game row game_pk=%d season=%d: %w", row.GamePK, row.Season, err)
		}
		affected, err := result.RowsAffected()
		if err == nil {
			written += affected
		} else {
			written++
		}
	}
	return written, nil
}

func stringValue(value *string) any {
	if value == nil {
		return nil
	}
	return strings.TrimSpace(*value)
}

func intValue(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func floatValue(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}
