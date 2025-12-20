package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

// AdvancedStatsRepository computes sabermetric stats.
type AdvancedStatsRepository struct {
	db *sql.DB
}

// NewAdvancedStatsRepository creates a new AdvancedStatsRepository.
func NewAdvancedStatsRepository(db *sql.DB) *AdvancedStatsRepository {
	return &AdvancedStatsRepository{db: db}
}

// PlayerAdvancedBatting computes wOBA, wRC+, ISO, BABIP, etc. for a player.
func (r *AdvancedStatsRepository) PlayerAdvancedBatting(ctx context.Context, playerID core.PlayerID, filter core.AdvancedBattingFilter) (*core.AdvancedBattingStats, error) {
	season := 2024
	if filter.Season != nil {
		season = int(*filter.Season)
	}

	var teamID sql.NullString
	if filter.TeamID != nil {
		teamID = sql.NullString{String: string(*filter.TeamID), Valid: true}
	}

	row := r.db.QueryRowContext(ctx, playerAdvancedBattingQuery, string(playerID), season, teamID)

	var stats core.AdvancedBattingStats
	var teamIDResult, leagueIDResult sql.NullString

	err := row.Scan(
		&stats.PA,
		&stats.AB,
		&stats.H,
		&stats.Doubles,
		&stats.Triples,
		&stats.HR,
		&stats.BB,
		&stats.IBB,
		&stats.HBP,
		&stats.SF,
		&stats.SH,
		&stats.SO,
		&stats.AVG,
		&stats.OBP,
		&stats.SLG,
		&stats.ISO,
		&stats.BABIP,
		&stats.KRate,
		&stats.BBRate,
		&stats.WOBA,
		&stats.WRAA,
		&stats.WRC,
		&stats.WRCPlus,
		&teamIDResult,
		&leagueIDResult,
	)
	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("batting stats", "")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query advanced batting stats: %w", err)
	}

	stats.PlayerID = playerID
	if teamIDResult.Valid {
		tid := core.TeamID(teamIDResult.String)
		stats.TeamID = &tid
	}

	var leagueID *core.LeagueID
	if leagueIDResult.Valid {
		lid := core.LeagueID(leagueIDResult.String)
		leagueID = &lid
	}

	stats.Context = core.StatContext{
		Season:      core.SeasonYear(season),
		League:      leagueID,
		Provider:    core.StatProviderInternal,
		ParkNeutral: false,
		RegSeason:   true,
	}

	stats.OPS = stats.OBP + stats.SLG
	return &stats, nil
}

// PlayerAdvancedBattingSplits returns advanced batting stats split by dimension.
// TODO: implement splits for advanced batting
func (r *AdvancedStatsRepository) PlayerAdvancedBattingSplits(ctx context.Context, playerID core.PlayerID, filter core.AdvancedBattingFilter) ([]core.AdvancedBattingStats, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// PlayerBaserunning calculates base running runs (wSB) for a player.
func (r *AdvancedStatsRepository) PlayerBaserunning(ctx context.Context, playerID core.PlayerID, season core.SeasonYear, teamID *core.TeamID) (*core.BaserunningStats, error) {
	var teamIDParam sql.NullString
	if teamID != nil {
		teamIDParam = sql.NullString{String: string(*teamID), Valid: true}
	}

	row := r.db.QueryRowContext(ctx, playerBaserunningQuery, string(playerID), int(season), teamIDParam)

	var stats core.BaserunningStats
	var playerIDStr, teamIDStr, leagueStr string

	err := row.Scan(
		&playerIDStr,
		&stats.Season,
		&teamIDStr,
		&leagueStr,
		&stats.SB,
		&stats.CS,
		&stats.BaserunningRuns,
	)
	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("baserunning stats", "")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query baserunning stats: %w", err)
	}

	stats.PlayerID = core.PlayerID(playerIDStr)
	if teamIDStr != "" {
		tid := core.TeamID(teamIDStr)
		stats.TeamID = &tid
	}
	if leagueStr != "" {
		lid := core.LeagueID(leagueStr)
		stats.League = &lid
	}
	return &stats, nil
}

// PlayerFielding calculates fielding runs for a player using range factor.
func (r *AdvancedStatsRepository) PlayerFielding(ctx context.Context, playerID core.PlayerID, season core.SeasonYear, teamID *core.TeamID) (*core.FieldingStats, error) {
	var teamIDParam sql.NullString
	if teamID != nil {
		teamIDParam = sql.NullString{String: string(*teamID), Valid: true}
	}

	row := r.db.QueryRowContext(ctx, playerFieldingQuery, string(playerID), int(season), teamIDParam)

	var stats core.FieldingStats
	var playerIDStr, teamIDStr, leagueStr string

	err := row.Scan(
		&playerIDStr,
		&stats.Season,
		&teamIDStr,
		&leagueStr,
		&stats.Position,
		&stats.Games,
		&stats.Putouts,
		&stats.Assists,
		&stats.Errors,
		&stats.RangeFactor,
		&stats.LeagueAvgRF,
		&stats.FieldingRuns,
	)
	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("fielding stats", "")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query fielding stats: %w", err)
	}

	stats.PlayerID = core.PlayerID(playerIDStr)
	if teamIDStr != "" {
		tid := core.TeamID(teamIDStr)
		stats.TeamID = &tid
	}
	if leagueStr != "" {
		lid := core.LeagueID(leagueStr)
		stats.League = &lid
	}
	return &stats, nil
}

// PlayerAdvancedPitching computes FIP, xFIP, ERA+, etc. for a pitcher.
func (r *AdvancedStatsRepository) PlayerAdvancedPitching(
	ctx context.Context,
	playerID core.PlayerID,
	filter core.AdvancedPitchingFilter,
) (*core.AdvancedPitchingStats, error) {
	season := 2024
	if filter.Season != nil {
		season = int(*filter.Season)
	}

	var teamID sql.NullString
	if filter.TeamID != nil {
		teamID = sql.NullString{String: string(*filter.TeamID), Valid: true}
	}

	row := r.db.QueryRowContext(ctx, playerAdvancedPitchingQuery, string(playerID), season, teamID)

	var stats core.AdvancedPitchingStats
	var teamIDResult sql.NullString

	err := row.Scan(
		&stats.IPOuts,
		&stats.BF,
		&stats.H,
		&stats.R,
		&stats.ER,
		&stats.HR,
		&stats.BB,
		&stats.IBB,
		&stats.HBP,
		&stats.SO,
		&stats.ERA,
		&stats.WHIP,
		&stats.KPer9,
		&stats.BBPer9,
		&stats.HRPer9,
		&stats.FIP,
		&teamIDResult,
	)
	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("pitching stats", "")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query advanced pitching stats: %w", err)
	}

	stats.PlayerID = playerID
	if teamIDResult.Valid {
		tid := core.TeamID(teamIDResult.String)
		stats.TeamID = &tid
	}

	stats.Context = core.StatContext{
		Season:      core.SeasonYear(season),
		Provider:    core.StatProviderInternal,
		ParkNeutral: false,
		RegSeason:   true,
	}
	return &stats, nil
}

// PlayerWAR computes WAR components and total WAR for a player.
func (r *AdvancedStatsRepository) PlayerWAR(ctx context.Context, playerID core.PlayerID, filter core.WARFilter) (*core.PlayerWARSummary, error) {
	season := 2024
	if filter.Season != nil {
		season = int(*filter.Season)
	}

	var teamID sql.NullString
	if filter.TeamID != nil {
		teamID = sql.NullString{String: string(*filter.TeamID), Valid: true}
	}

	row := r.db.QueryRowContext(ctx, playerWARQuery, string(playerID), season, teamID)

	var war core.PlayerWARSummary
	var playerIDStr, teamIDStr, leagueStr, position string
	var pa int
	var battingRuns, baserunningRuns, fieldingRuns, positionalAdj, replacementRuns, runsAboveReplacement, warValue float64

	err := row.Scan(
		&playerIDStr,
		&season,
		&teamIDStr,
		&leagueStr,
		&pa,
		&battingRuns,
		&baserunningRuns,
		&fieldingRuns,
		&positionalAdj,
		&replacementRuns,
		&runsAboveReplacement,
		&warValue,
		&position,
	)
	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("WAR data", "")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query WAR: %w", err)
	}

	war.PlayerID = core.PlayerID(playerIDStr)
	if teamIDStr != "" {
		tid := core.TeamID(teamIDStr)
		war.TeamID = &tid
	}

	var leagueID *core.LeagueID
	if leagueStr != "" {
		lid := core.LeagueID(leagueStr)
		leagueID = &lid
	}

	war.Context = core.StatContext{
		Season:      core.SeasonYear(season),
		League:      leagueID,
		Provider:    core.StatProviderInternal,
		ParkNeutral: false,
		RegSeason:   true,
	}

	war.WAR = warValue
	war.BattingRuns = &battingRuns
	war.BaseRunningRuns = &baserunningRuns
	war.FieldingRuns = &fieldingRuns
	war.PositionalRuns = &positionalAdj
	war.ReplacementRuns = &replacementRuns
	return &war, nil
}

// SeasonBattingLeaders returns top N players by advanced batting stat.
func (r *AdvancedStatsRepository) SeasonBattingLeaders(ctx context.Context, season core.SeasonYear, stat string, limit int, filter core.AdvancedBattingFilter) ([]core.AdvancedBattingStats, error) {
	minPA := 502
	if filter.MinPA != nil {
		minPA = *filter.MinPA
	}

	var teamID, leagueID sql.NullString
	if filter.TeamID != nil {
		teamID = sql.NullString{String: string(*filter.TeamID), Valid: true}
	}
	if filter.League != nil {
		leagueID = sql.NullString{String: string(*filter.League), Valid: true}
	}

	statUpper := stat
	if stat == "" {
		statUpper = "WRC_PLUS"
	}

	rows, err := r.db.QueryContext(ctx, seasonBattingLeadersQuery, int(season), teamID, leagueID, minPA, statUpper, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query batting leaders: %w", err)
	}
	defer rows.Close()

	var leaders []core.AdvancedBattingStats
	for rows.Next() {
		var stats core.AdvancedBattingStats
		var playerIDStr string
		var teamIDResult, leagueIDResult sql.NullString

		err = rows.Scan(
			&playerIDStr,
			&stats.PA,
			&stats.AB,
			&stats.H,
			&stats.Doubles,
			&stats.Triples,
			&stats.HR,
			&stats.BB,
			&stats.IBB,
			&stats.HBP,
			&stats.SF,
			&stats.SH,
			&stats.SO,
			&stats.AVG,
			&stats.OBP,
			&stats.SLG,
			&stats.ISO,
			&stats.BABIP,
			&stats.KRate,
			&stats.BBRate,
			&stats.WOBA,
			&stats.WRAA,
			&stats.WRC,
			&stats.WRCPlus,
			&teamIDResult,
			&leagueIDResult,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batting leader: %w", err)
		}

		stats.PlayerID = core.PlayerID(playerIDStr)
		if teamIDResult.Valid {
			tid := core.TeamID(teamIDResult.String)
			stats.TeamID = &tid
		}

		var league *core.LeagueID
		if leagueIDResult.Valid {
			lid := core.LeagueID(leagueIDResult.String)
			league = &lid
		}

		stats.Context = core.StatContext{
			Season:      season,
			League:      league,
			Provider:    core.StatProviderInternal,
			ParkNeutral: false,
			RegSeason:   true,
		}

		stats.OPS = stats.OBP + stats.SLG
		leaders = append(leaders, stats)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating batting leaders: %w", err)
	}

	return leaders, nil
}

// SeasonPitchingLeaders returns top N pitchers by advanced pitching stat.
func (r *AdvancedStatsRepository) SeasonPitchingLeaders(ctx context.Context, season core.SeasonYear, stat string, limit int, filter core.AdvancedPitchingFilter) ([]core.AdvancedPitchingStats, error) {
	minIPOuts := 162 * 3
	if filter.MinIP != nil {
		minIPOuts = int(*filter.MinIP * 3)
	}

	var teamID sql.NullString
	if filter.TeamID != nil {
		teamID = sql.NullString{String: string(*filter.TeamID), Valid: true}
	}

	statUpper := stat
	if stat == "" {
		statUpper = "ERA"
	}

	rows, err := r.db.QueryContext(ctx, seasonPitchingLeadersQuery, int(season), teamID, minIPOuts, statUpper, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pitching leaders: %w", err)
	}
	defer rows.Close()

	var leaders []core.AdvancedPitchingStats
	for rows.Next() {
		var stats core.AdvancedPitchingStats
		var playerIDStr string
		var teamIDResult sql.NullString

		err = rows.Scan(
			&playerIDStr,
			&stats.IPOuts,
			&stats.BF,
			&stats.H,
			&stats.R,
			&stats.ER,
			&stats.HR,
			&stats.BB,
			&stats.IBB,
			&stats.HBP,
			&stats.SO,
			&stats.ERA,
			&stats.WHIP,
			&stats.KPer9,
			&stats.BBPer9,
			&stats.HRPer9,
			&stats.FIP,
			&teamIDResult,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pitching leader: %w", err)
		}

		stats.PlayerID = core.PlayerID(playerIDStr)
		if teamIDResult.Valid {
			tid := core.TeamID(teamIDResult.String)
			stats.TeamID = &tid
		}

		stats.Context = core.StatContext{
			Season:      season,
			Provider:    core.StatProviderInternal,
			ParkNeutral: false,
			RegSeason:   true,
		}

		leaders = append(leaders, stats)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pitching leaders: %w", err)
	}

	return leaders, nil
}

// SeasonWARLeaders returns top N players by WAR.
func (r *AdvancedStatsRepository) SeasonWARLeaders(ctx context.Context, season core.SeasonYear, limit int, filter core.WARFilter) ([]core.PlayerWARSummary, error) {
	minPA := 502
	if filter.MinPA != nil {
		minPA = *filter.MinPA
	}

	var teamID sql.NullString
	if filter.TeamID != nil {
		teamID = sql.NullString{String: string(*filter.TeamID), Valid: true}
	}

	rows, err := r.db.QueryContext(ctx, seasonWARLeadersQuery, int(season), teamID, minPA, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query WAR leaders: %w", err)
	}
	defer rows.Close()

	var leaders []core.PlayerWARSummary
	for rows.Next() {
		var war core.PlayerWARSummary
		var playerIDStr, teamIDStr, leagueStr string
		var positionNull sql.NullString
		var seasonYear int
		var pa int
		var battingRuns, baserunningRuns, fieldingRuns, positionalAdj, replacementRuns, runsAboveReplacement, warValue float64

		err = rows.Scan(
			&playerIDStr,
			&seasonYear,
			&teamIDStr,
			&leagueStr,
			&pa,
			&battingRuns,
			&baserunningRuns,
			&fieldingRuns,
			&positionalAdj,
			&replacementRuns,
			&runsAboveReplacement,
			&warValue,
			&positionNull,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan WAR leader: %w", err)
		}

		war.PlayerID = core.PlayerID(playerIDStr)
		if teamIDStr != "" {
			tid := core.TeamID(teamIDStr)
			war.TeamID = &tid
		}

		var leagueID *core.LeagueID
		if leagueStr != "" {
			lid := core.LeagueID(leagueStr)
			leagueID = &lid
		}

		war.Context = core.StatContext{
			Season:      season,
			League:      leagueID,
			Provider:    core.StatProviderInternal,
			ParkNeutral: false,
			RegSeason:   true,
		}

		war.WAR = warValue
		war.BattingRuns = &battingRuns
		war.BaseRunningRuns = &baserunningRuns
		war.FieldingRuns = &fieldingRuns
		war.PositionalRuns = &positionalAdj
		war.ReplacementRuns = &replacementRuns

		leaders = append(leaders, war)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating WAR leaders: %w", err)
	}

	return leaders, nil
}
