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

//go:embed queries/player_high_leverage_pas.sql
var playerHighLeveragePAsQuery string

//go:embed queries/player_leverage_summary.sql
var playerLeverageSummaryQuery string

//go:embed queries/game_win_probability_summary.sql
var gameWinProbabilitySummaryQuery string

//go:embed queries/season_batting_leaders.sql
var seasonBattingLeadersQuery string

//go:embed queries/season_pitching_leaders.sql
var seasonPitchingLeadersQuery string

//go:embed queries/season_war_leaders.sql
var seasonWARLeadersQuery string

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
// The role parameter filters PAs to "batter", "pitcher", or "" (both).
// This computes average leverage index and categorizes PAs by leverage tier using SQL aggregation.
func (r *LeverageRepository) PlayerLeverageSummary(ctx context.Context, playerID core.PlayerID, season core.SeasonYear, role string) (*core.PlayerLeverageSummary, error) {
	if role != "batter" && role != "pitcher" {
		role = ""
	}

	row := r.db.QueryRowContext(ctx, playerLeverageSummaryQuery, string(playerID), int(season), role)

	var teamIDStr, leagueStr sql.NullString
	var avgLI, totalWPA float64
	var lowPA, mediumPA, highPA int
	var lowIPOuts, mediumIPOuts, highIPOuts int

	err := row.Scan(
		&teamIDStr,
		&leagueStr,
		&avgLI,
		&lowPA,
		&mediumPA,
		&highPA,
		&totalWPA,
		&lowIPOuts,
		&mediumIPOuts,
		&highIPOuts,
	)
	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("plate appearances", "")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query player leverage summary: %w", err)
	}

	summary := &core.PlayerLeverageSummary{
		PlayerID:             playerID,
		AvgLeverageIndex:     avgLI,
		LowLeveragePA:        lowPA,
		MediumLeveragePA:     mediumPA,
		HighLeveragePA:       highPA,
		LowLeverageIPOuts:    lowIPOuts,
		MediumLeverageIPOuts: mediumIPOuts,
		HighLeverageIPOuts:   highIPOuts,
		Context: core.StatContext{
			Season:      season,
			Provider:    core.StatProviderInternal,
			ParkNeutral: false,
			RegSeason:   true,
		},
	}

	if teamIDStr.Valid {
		tid := core.TeamID(teamIDStr.String)
		summary.TeamID = &tid
	}

	if leagueStr.Valid {
		lid := core.LeagueID(leagueStr.String)
		summary.Context.League = &lid
	}

	summary.WinProbabilityAdded = &totalWPA

	return summary, nil
}

// PlayerHighLeveragePAs returns high-leverage plate appearances for a player.
// This queries all plate appearances for a player (as batter or pitcher) and filters
// by leverage index threshold. The LI calculation is done in SQL.
func (r *LeverageRepository) PlayerHighLeveragePAs(ctx context.Context, playerID core.PlayerID, season core.SeasonYear, minLI float64) ([]core.PlateAppearanceLeverage, error) {
	rows, err := r.db.QueryContext(ctx, playerHighLeveragePAsQuery, string(playerID), int(season), minLI)
	if err != nil {
		return nil, fmt.Errorf("failed to query high leverage PAs: %w", err)
	}
	defer rows.Close()

	var leverages []core.PlateAppearanceLeverage
	for rows.Next() {
		var l core.PlateAppearanceLeverage
		var gameIDStr, batterID, pitcherID string
		var topBot int

		err = rows.Scan(
			&gameIDStr,
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
			&l.LeverageIndex,
			&l.WinExpectancyBefore,
			&l.WinExpectancyAfter,
			&l.WEChange,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan high leverage PA: %w", err)
		}

		l.GameID = core.GameID(gameIDStr)
		l.TopOfInning = topBot == 0
		l.BatterID = core.PlayerID(batterID)
		l.PitcherID = core.PlayerID(pitcherID)

		leverages = append(leverages, l)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating high leverage PAs: %w", err)
	}

	return leverages, nil
}

// GameWinProbabilitySummary returns summary stats for a game's win probability.
// This queries game information from SQL and returns a summary including biggest WE swings.
func (r *LeverageRepository) GameWinProbabilitySummary(
	ctx context.Context,
	gameID core.GameID,
) (*core.GameWinProbabilitySummary, error) {
	row := r.db.QueryRowContext(ctx, gameWinProbabilitySummaryQuery, string(gameID))

	var summary core.GameWinProbabilitySummary
	var gameIDStr, homeTeamStr, awayTeamStr string

	var posEventID sql.NullInt64
	var posInning, posTopBot sql.NullInt64
	var posHomeScore, posAwayScore, posOuts sql.NullInt64
	var posBases, posBatterID, posPitcherID, posDesc sql.NullString
	var posWEBefore, posWEAfter, posWEChange sql.NullFloat64

	var negEventID sql.NullInt64
	var negInning, negTopBot sql.NullInt64
	var negHomeScore, negAwayScore, negOuts sql.NullInt64
	var negBases, negBatterID, negPitcherID, negDesc sql.NullString
	var negWEBefore, negWEAfter, negWEChange sql.NullFloat64

	err := row.Scan(
		&gameIDStr, &summary.Season, &homeTeamStr, &awayTeamStr,
		&summary.HomeWinProbStart, &summary.HomeWinProbEnd,

		&posEventID, &posInning, &posTopBot, &posHomeScore, &posAwayScore,
		&posOuts, &posBases, &posBatterID, &posPitcherID,
		&posDesc, &posWEBefore, &posWEAfter, &posWEChange,

		&negEventID, &negInning, &negTopBot,
		&negHomeScore, &negAwayScore, &negOuts, &negBases, &negBatterID,
		&negPitcherID, &negDesc, &negWEBefore, &negWEAfter, &negWEChange,
	)
	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("game", string(gameID))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query game win probability summary: %w", err)
	}

	summary.GameID = core.GameID(gameIDStr)
	summary.HomeTeam = core.TeamID(homeTeamStr)
	summary.AwayTeam = core.TeamID(awayTeamStr)

	if posEventID.Valid {
		summary.BiggestPositiveSwing = &core.PlateAppearanceLeverage{
			GameID:              gameID,
			EventID:             int(posEventID.Int64),
			Inning:              int(posInning.Int64),
			TopOfInning:         posTopBot.Int64 == 0,
			HomeScoreBefore:     int(posHomeScore.Int64),
			AwayScoreBefore:     int(posAwayScore.Int64),
			OutsBefore:          int(posOuts.Int64),
			BasesBefore:         posBases.String,
			BatterID:            core.PlayerID(posBatterID.String),
			PitcherID:           core.PlayerID(posPitcherID.String),
			Description:         posDesc.String,
			WinExpectancyBefore: posWEBefore.Float64,
			WinExpectancyAfter:  posWEAfter.Float64,
			WEChange:            posWEChange.Float64,
		}
	}

	if negEventID.Valid {
		summary.BiggestNegativeSwing = &core.PlateAppearanceLeverage{
			GameID:              gameID,
			EventID:             int(negEventID.Int64),
			Inning:              int(negInning.Int64),
			TopOfInning:         negTopBot.Int64 == 0,
			HomeScoreBefore:     int(negHomeScore.Int64),
			AwayScoreBefore:     int(negAwayScore.Int64),
			OutsBefore:          int(negOuts.Int64),
			BasesBefore:         negBases.String,
			BatterID:            core.PlayerID(negBatterID.String),
			PitcherID:           core.PlayerID(negPitcherID.String),
			Description:         negDesc.String,
			WinExpectancyBefore: negWEBefore.Float64,
			WinExpectancyAfter:  negWEAfter.Float64,
			WEChange:            negWEChange.Float64,
		}
	}

	return &summary, nil
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
