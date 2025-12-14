package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"math"

	"stormlightlabs.org/baseball/internal/core"
)

//go:embed queries/player_advanced_batting.sql
var playerAdvancedBattingQuery string

//go:embed queries/player_advanced_pitching.sql
var playerAdvancedPitchingQuery string

//go:embed queries/player_baserunning.sql
var playerBaserunningQuery string

//go:embed queries/player_fielding.sql
var playerFieldingQuery string

//go:embed queries/player_war.sql
var playerWARQuery string

//go:embed queries/park_factor.sql
var parkFactorQuery string

//go:embed queries/park_factor_series.sql
var parkFactorSeriesQuery string

//go:embed queries/season_park_factors.sql
var seasonParkFactorsQuery string

//go:embed queries/game_plate_leverages.sql
var gamePlateLeveragesQuery string

//go:embed queries/season_batting_leaders.sql
var seasonBattingLeadersQuery string

//go:embed queries/season_pitching_leaders.sql
var seasonPitchingLeadersQuery string

//go:embed queries/season_war_leaders.sql
var seasonWARLeadersQuery string

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
		return nil, fmt.Errorf("no batting stats found for player %s in season %d", playerID, season)
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
		return nil, fmt.Errorf("no baserunning stats found for player %s in season %d", playerID, season)
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
		return nil, fmt.Errorf("no fielding stats found for player %s in season %d", playerID, season)
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
		return nil, fmt.Errorf("no pitching stats found for player %s in season %d", playerID, season)
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
		return nil, fmt.Errorf("no WAR data found for player %s in season %d", playerID, season)
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
	// Default minimum PA threshold for leaderboards (3.1 PA per team game)
	minPA := 502 // ~3.1 PA * 162 games
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

// LeverageRepository computes leverage index and win probability metrics.
type LeverageRepository struct {
	db *sql.DB
}

// NewLeverageRepository creates a new LeverageRepository.
func NewLeverageRepository(db *sql.DB) *LeverageRepository {
	return &LeverageRepository{db: db}
}

// GamePlateLeverages returns leverage index for each plate appearance in a game.
func (r *LeverageRepository) GamePlateLeverages(ctx context.Context, gameID core.GameID, minLI *float64) ([]core.PlateAppearanceLeverage, error) {
	minLIValue := 0.0
	if minLI != nil {
		minLIValue = *minLI
	}

	rows, err := r.db.QueryContext(ctx, gamePlateLeveragesQuery, string(gameID))
	if err != nil {
		return nil, fmt.Errorf("failed to query plate leverages: %w", err)
	}
	defer rows.Close()

	var leverages []core.PlateAppearanceLeverage
	for rows.Next() {
		var l core.PlateAppearanceLeverage
		var batterID, pitcherID string
		var topBot int

		err = rows.Scan(
			&l.EventID,
			&l.Inning,
			&topBot,
			&l.HomeScoreBefore,
			&l.AwayScoreBefore,
			&l.OutsBefore,
			&l.BasesBefore,
			&batterID,
			&pitcherID,
			&l.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plate leverage: %w", err)
		}

		l.GameID = gameID
		l.TopOfInning = topBot == 0
		l.BatterID = core.PlayerID(batterID)
		l.PitcherID = core.PlayerID(pitcherID)
		l.LeverageIndex = calculateLeverageIndex(l.Inning, l.OutsBefore, l.HomeScoreBefore-l.AwayScoreBefore, l.BasesBefore)
		l.WinExpectancyBefore = 0.5
		l.WinExpectancyAfter = 0.5
		l.WEChange = 0.0

		if l.LeverageIndex >= minLIValue {
			leverages = append(leverages, l)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plate leverages: %w", err)
	}

	return leverages, nil
}

// PlayerLeverageSummary aggregates leverage metrics for a player.
// TODO: implement player leverage summary
func (r *LeverageRepository) PlayerLeverageSummary(ctx context.Context, playerID core.PlayerID, season core.SeasonYear, role string) (*core.PlayerLeverageSummary, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// PlayerHighLeveragePAs returns high-leverage plate appearances for a player.
// TODO: implement high leverage PAs
func (r *LeverageRepository) PlayerHighLeveragePAs(ctx context.Context, playerID core.PlayerID, season core.SeasonYear, minLI float64) ([]core.PlateAppearanceLeverage, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// GameWinProbabilitySummary returns summary stats for a game's win probability.
// TODO: implement win probability summary
func (r *LeverageRepository) GameWinProbabilitySummary(
	ctx context.Context,
	gameID core.GameID,
) (*core.GameWinProbabilitySummary, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// calculateLeverageIndex is a simplified leverage index calculation.
// A proper implementation would use historical win expectancy tables.
func calculateLeverageIndex(inning, outs, scoreDiff int, bases string) float64 {
	li := 1.0

	if inning >= 7 {
		li *= 1.5
	}
	if inning >= 9 {
		li *= 2.0
	}

	absScoreDiff := math.Abs(float64(scoreDiff))
	if absScoreDiff <= 1 {
		li *= 1.5
	} else if absScoreDiff <= 2 {
		li *= 1.2
	} else if absScoreDiff >= 5 {
		li *= 0.3
	}

	runnersOn := 0
	for _, b := range bases {
		if b == '1' {
			runnersOn++
		}
	}
	li *= (1.0 + float64(runnersOn)*0.2)

	if outs == 2 {
		li *= 1.3
	}

	return li
}

// ParkFactorRepository computes and retrieves park factors.
type ParkFactorRepository struct {
	db *sql.DB
}

// NewParkFactorRepository creates a new ParkFactorRepository.
func NewParkFactorRepository(db *sql.DB) *ParkFactorRepository {
	return &ParkFactorRepository{db: db}
}

// ParkFactor returns park factors for a specific park and season.
func (r *ParkFactorRepository) ParkFactor(ctx context.Context, parkID core.ParkID, season core.SeasonYear) (*core.ParkFactor, error) {
	row := r.db.QueryRowContext(ctx, parkFactorQuery, string(parkID), int(season))

	var pf core.ParkFactor
	var homeTeam string

	err := row.Scan(
		&pf.ParkID,
		&homeTeam,
		&pf.Season,
		&pf.GamesSampled,
		&pf.RunsFactor,
		&pf.HRFactor,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no park factor found for park %s in season %d", parkID, season)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query park factor: %w", err)
	}

	pf.Provider = "internal"
	pf.MultiYear = false

	return &pf, nil
}

// ParkFactorSeries returns park factors over a range of seasons.
func (r *ParkFactorRepository) ParkFactorSeries(ctx context.Context, parkID core.ParkID, fromSeason, toSeason core.SeasonYear) ([]core.ParkFactor, error) {
	rows, err := r.db.QueryContext(ctx, parkFactorSeriesQuery, string(parkID), int(fromSeason), int(toSeason))
	if err != nil {
		return nil, fmt.Errorf("failed to query park factor series: %w", err)
	}
	defer rows.Close()

	var factors []core.ParkFactor
	for rows.Next() {
		var pf core.ParkFactor
		var homeTeam string

		err = rows.Scan(
			&pf.ParkID,
			&homeTeam,
			&pf.Season,
			&pf.GamesSampled,
			&pf.RunsFactor,
			&pf.HRFactor,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan park factor: %w", err)
		}

		pf.Provider = "internal"
		pf.MultiYear = false

		factors = append(factors, pf)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating park factors: %w", err)
	}

	return factors, nil
}

// SeasonParkFactors returns all park factors for a given season.
func (r *ParkFactorRepository) SeasonParkFactors(ctx context.Context, season core.SeasonYear, factorType *string) ([]core.ParkFactor, error) {
	rows, err := r.db.QueryContext(ctx, seasonParkFactorsQuery, int(season))
	if err != nil {
		return nil, fmt.Errorf("failed to query season park factors: %w", err)
	}
	defer rows.Close()

	var factors []core.ParkFactor
	for rows.Next() {
		var pf core.ParkFactor
		var homeTeam string

		err = rows.Scan(
			&pf.ParkID,
			&homeTeam,
			&pf.Season,
			&pf.GamesSampled,
			&pf.RunsFactor,
			&pf.HRFactor,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan park factor: %w", err)
		}

		pf.Provider = "internal"
		pf.MultiYear = false

		factors = append(factors, pf)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating park factors: %w", err)
	}

	return factors, nil
}

// MultiYearParkFactor returns a park factor averaged over multiple seasons.
func (r *ParkFactorRepository) MultiYearParkFactor(ctx context.Context, parkID core.ParkID, fromSeason, toSeason core.SeasonYear) (*core.ParkFactor, error) {
	series, err := r.ParkFactorSeries(ctx, parkID, fromSeason, toSeason)
	if err != nil {
		return nil, err
	}

	if len(series) == 0 {
		return nil, fmt.Errorf("no park factors found for park %s between %d and %d", parkID, fromSeason, toSeason)
	}

	var totalGames int
	var totalRunsFactor, totalHRFactor float64

	for _, pf := range series {
		totalGames += pf.GamesSampled
		totalRunsFactor += pf.RunsFactor * float64(pf.GamesSampled)
		totalHRFactor += pf.HRFactor * float64(pf.GamesSampled)
	}

	avgPF := &core.ParkFactor{
		ParkID:       string(parkID),
		Season:       int(toSeason),
		RunsFactor:   totalRunsFactor / float64(totalGames),
		HRFactor:     totalHRFactor / float64(totalGames),
		GamesSampled: totalGames,
		Provider:     "internal",
		MultiYear:    true,
	}

	return avgPF, nil
}
