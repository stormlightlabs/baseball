package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"math"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

//go:embed queries/player_hitting_streaks.sql
var playerHittingStreaksQuery string

//go:embed queries/player_scoreless_innings.sql
var playerScorelessInningsQuery string

//go:embed queries/team_run_differential.sql
var teamRunDifferentialQuery string

//go:embed queries/game_win_probability.sql
var gameWinProbabilityQuery string

//go:embed queries/player_splits_home_away.sql
var playerSplitsHomeAwayQuery string

//go:embed queries/player_splits_pitcher_handed.sql
var playerSplitsPitcherHandedQuery string

//go:embed queries/player_splits_month.sql
var playerSplitsMonthQuery string

type DerivedStatsRepository struct {
	db *sql.DB
}

func NewDerivedStatsRepository(db *sql.DB) *DerivedStatsRepository {
	return &DerivedStatsRepository{db: db}
}

// PlayerStreaks retrieves hitting or scoreless innings streaks for a player.
func (r *DerivedStatsRepository) PlayerStreaks(ctx context.Context, playerID core.PlayerID, kind core.StreakKind, season core.SeasonYear, minLength int) ([]core.Streak, error) {
	var rows *sql.Rows
	var err error

	seasonStr := fmt.Sprintf("%d", season)

	switch kind {
	case core.StreakKindHitting:
		rows, err = r.db.QueryContext(ctx, playerHittingStreaksQuery, string(playerID), seasonStr, minLength)
	case core.StreakKindScorelessInnings:
		rows, err = r.db.QueryContext(ctx, playerScorelessInningsQuery, string(playerID), seasonStr, float64(minLength))
	default:
		return nil, fmt.Errorf("unsupported streak kind: %s", kind)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query %s streaks: %w", kind, err)
	}
	defer rows.Close()

	var streaks []core.Streak

	for rows.Next() {
		var s core.Streak
		var startDate, endDate string
		var startGameID, endGameID string

		if kind == core.StreakKindHitting {
			var totalAB, totalHits int
			err = rows.Scan(
				&s.EntityID,
				&startDate,
				&endDate,
				&s.Length,
				&startGameID,
				&endGameID,
				&totalAB,
				&totalHits,
			)
		} else {
			var gamesInStreak int
			var lengthFloat float64
			err = rows.Scan(
				&s.EntityID,
				&startDate,
				&endDate,
				&gamesInStreak,
				&lengthFloat,
				&startGameID,
				&endGameID,
			)
			s.Length = int(lengthFloat)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to scan streak: %w", err)
		}

		s.ID = fmt.Sprintf("%s-%s-%s-%s", s.EntityID, kind, startDate, endDate)
		s.Kind = kind
		s.EntityType = core.StreakEntityPlayer
		s.Season = int(season)
		s.StartGameID = core.GameID(startGameID)
		s.EndGameID = core.GameID(endGameID)
		s.StartDate = formatDate(startDate)
		s.EndDate = formatDate(endDate)
		s.Label = fmt.Sprintf("%d-game %s streak", s.Length, kind)

		streaks = append(streaks, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating streaks: %w", err)
	}

	return streaks, nil
}

// TeamRunDifferential calculates season run differential with rolling windows.
func (r *DerivedStatsRepository) TeamRunDifferential(ctx context.Context, teamID core.TeamID, season core.SeasonYear, windows []int) (*core.RunDifferentialSeries, error) {
	rows, err := r.db.QueryContext(ctx, teamRunDifferentialQuery, string(teamID), int(season))
	if err != nil {
		return nil, fmt.Errorf("failed to query run differential: %w", err)
	}
	defer rows.Close()

	var games []core.RunDifferentialGamePoint
	totalRS, totalRA := 0, 0

	for rows.Next() {
		var g core.RunDifferentialGamePoint
		var gameID, date, opponentID string
		var isHome int
		var gameNum int

		err = rows.Scan(
			&gameID,
			&date,
			&opponentID,
			&isHome,
			&g.RunsScored,
			&g.RunsAllowed,
			&g.Differential,
			&g.CumulativeDiff,
			&gameNum,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}

		g.GameID = core.GameID(gameID)
		g.Date = formatDate(date)
		g.OpponentID = core.TeamID(opponentID)
		g.Home = isHome == 1

		totalRS += g.RunsScored
		totalRA += g.RunsAllowed

		games = append(games, g)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating games: %w", err)
	}

	var rollingWindows []core.RunDifferentialWindow
	for _, windowSize := range windows {
		window := r.calculateRollingWindow(games, windowSize)
		rollingWindows = append(rollingWindows, window)
	}

	return &core.RunDifferentialSeries{
		EntityType:      "team",
		EntityID:        string(teamID),
		Season:          int(season),
		GamesPlayed:     len(games),
		RunsScored:      totalRS,
		RunsAllowed:     totalRA,
		RunDifferential: totalRS - totalRA,
		Games:           games,
		Rolling:         rollingWindows,
	}, nil
}

// calculateRollingWindow computes rolling window stats for a given window size.
func (r *DerivedStatsRepository) calculateRollingWindow(games []core.RunDifferentialGamePoint, windowSize int) core.RunDifferentialWindow {
	window := core.RunDifferentialWindow{
		WindowSize: windowSize,
		Label:      fmt.Sprintf("last_%d", windowSize),
		Points:     []core.RunDifferentialWindowPoint{},
	}

	for i := windowSize - 1; i < len(games); i++ {
		rs, ra := 0, 0
		for j := i - windowSize + 1; j <= i; j++ {
			rs += games[j].RunsScored
			ra += games[j].RunsAllowed
		}

		point := core.RunDifferentialWindowPoint{
			EndGameID:       games[i].GameID,
			EndDate:         games[i].Date,
			GamesInWindow:   windowSize,
			RunsScored:      rs,
			RunsAllowed:     ra,
			RunDifferential: rs - ra,
		}
		window.Points = append(window.Points, point)
	}

	return window
}

// GameWinProbability returns win probability curve for a game.
func (r *DerivedStatsRepository) GameWinProbability(ctx context.Context, gameID core.GameID) (*core.WinProbabilityCurve, error) {
	rows, err := r.db.QueryContext(ctx, gameWinProbabilityQuery, string(gameID))
	if err != nil {
		return nil, fmt.Errorf("failed to query win probability: %w", err)
	}
	defer rows.Close()

	curve := &core.WinProbabilityCurve{
		GameID: gameID,
		Points: []core.WinProbabilityPoint{},
	}

	if len(gameID) >= 11 {
		curve.HomeTeam = core.TeamID(string(gameID)[8:11])
	}

	for rows.Next() {
		var p core.WinProbabilityPoint
		var topOfInning bool

		err = rows.Scan(
			&p.EventIndex,
			&p.Inning,
			&topOfInning,
			&p.HomeScore,
			&p.AwayScore,
			&p.Outs,
			&p.Bases,
			&p.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan play: %w", err)
		}

		p.TopOfInning = topOfInning

		p.HomeWinProb, p.AwayWinProb = r.calculateWinProbability(
			p.Inning,
			p.TopOfInning,
			p.HomeScore,
			p.AwayScore,
			p.Outs,
			p.Bases,
		)

		curve.Points = append(curve.Points, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plays: %w", err)
	}

	return curve, nil
}

// calculateWinProbability is a simplified win probability model.
// Algorithm
//   - we use a simplified model based on score differential and inning
//   - Base probability from comes from score differential, starting at 0.5, where each run is worth about 15% win probability
//   - If it's late in the game, score differential matters more
//
// TODO: historical win expectancy tables based on game state.
func (r *DerivedStatsRepository) calculateWinProbability(inning int, topOfInning bool, homeScore, awayScore, outs int, _ string) (float64, float64) {
	scoreDiff := homeScore - awayScore
	baseProb := 0.5
	if scoreDiff != 0 {
		baseProb = 0.5 + (float64(scoreDiff) * 0.15)
	}

	inningsRemaining := 9.0 - float64(inning)
	if topOfInning {
		inningsRemaining += 0.5
	}

	if inningsRemaining <= 1 {
		if scoreDiff > 0 {
			baseProb = 0.5 + (float64(scoreDiff) * 0.25)
		} else if scoreDiff < 0 {
			baseProb = 0.5 + (float64(scoreDiff) * 0.25)
		}
	}

	if inning >= 9 && !topOfInning && homeScore > awayScore {
		baseProb = 1.0
	} else if inning >= 9 && topOfInning && outs == 3 && homeScore < awayScore {
		baseProb = 0.0
	}

	homeWinProb := math.Max(0.0, math.Min(1.0, baseProb))
	return homeWinProb, 1.0 - homeWinProb
}

// formatDate converts YYYYMMDD to YYYY-MM-DD.
func formatDate(dateStr string) string {
	if len(dateStr) != 8 {
		return dateStr
	}
	t, err := time.Parse("20060102", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format(time.DateOnly)
}

// PlayerSplits calculates batting splits for a player by dimension.
func (r *DerivedStatsRepository) PlayerSplits(
	ctx context.Context,
	playerID core.PlayerID,
	dimension core.SplitDimension,
	season core.SeasonYear,
) (*core.SplitResult, error) {
	var query string

	switch dimension {
	case core.SplitDimHomeAway:
		query = playerSplitsHomeAwayQuery
	case core.SplitDimPitcherHanded:
		query = playerSplitsPitcherHandedQuery
	case core.SplitDimMonth:
		query = playerSplitsMonthQuery
	default:
		return nil, fmt.Errorf("unsupported split dimension: %s", dimension)
	}

	seasonStr := fmt.Sprintf("%d", season)
	rows, err := r.db.QueryContext(ctx, query, string(playerID), seasonStr)
	if err != nil {
		return nil, fmt.Errorf("failed to query %s splits: %w", dimension, err)
	}
	defer rows.Close()

	result := &core.SplitResult{
		EntityType: core.SplitEntityPlayer,
		EntityID:   string(playerID),
		Season:     int(season),
		Dimension:  dimension,
		Groups:     []core.SplitGroup{},
	}

	for rows.Next() {
		var group core.SplitGroup
		var avg, obp, slg float64
		var metaValue sql.NullString

		err = rows.Scan(
			&group.Key,
			&group.Label,
			&metaValue,
			&group.Games,
			&group.PA,
			&group.AB,
			&group.H,
			&group.HR,
			&group.BB,
			&group.SO,
			&avg,
			&obp,
			&slg,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan split group: %w", err)
		}

		group.AVG = roundFloat(avg, 3)
		group.OBP = roundFloat(obp, 3)
		group.SLG = roundFloat(slg, 3)
		group.OPS = roundFloat(obp+slg, 3)

		if metaValue.Valid {
			group.Meta = map[string]string{}
			switch dimension {
			case core.SplitDimPitcherHanded:
				group.Meta["handedness"] = metaValue.String
			case core.SplitDimMonth:
				group.Meta["month_number"] = metaValue.String
			}
		}

		result.Groups = append(result.Groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating split groups: %w", err)
	}

	return result, nil
}

// roundFloat rounds a float to n decimal places.
func roundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
