package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/cache"
	"stormlightlabs.org/baseball/internal/core"
)

type AchievementRepository struct {
	db    *sql.DB
	cache *cache.CachedRepository
}

func NewAchievementRepository(db *sql.DB, cacheClient *cache.Client) *AchievementRepository {
	return &AchievementRepository{
		db:    db,
		cache: cache.NewCachedRepository(cacheClient, "achievement"),
	}
}

// ListNoHitters retrieves no-hitter achievements from the materialized view.
func (r *AchievementRepository) ListNoHitters(ctx context.Context, filter core.AchievementFilter) ([]core.NoHitter, error) {
	query, args := buildNoHittersQuery(filter)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query no-hitters: %w", err)
	}
	defer rows.Close()

	var results []core.NoHitter
	for rows.Next() {
		var nh core.NoHitter
		var date string
		var winningPitcherID, winningPitcherName sql.NullString

		err := rows.Scan(
			&nh.GameID,
			&nh.TeamID,
			&nh.OpponentTeamID,
			&date,
			&nh.Season,
			&nh.HomeTeam,
			&nh.VisitingTeam,
			&nh.HomeScore,
			&nh.VisitingScore,
			&nh.Innings,
			&nh.ParkID,
			&nh.TeamLocation,
			&winningPitcherID,
			&winningPitcherName,
		)
		if err != nil {
			return nil, fmt.Errorf("scan no-hitter: %w", err)
		}

		parsedDate, err := time.Parse("20060102", date)
		if err != nil {
			return nil, fmt.Errorf("parse date %s: %w", date, err)
		}
		nh.Date = parsedDate

		if winningPitcherID.Valid {
			nh.WinningPitcherID = &winningPitcherID.String
		}
		if winningPitcherName.Valid {
			nh.WinningPitcherName = &winningPitcherName.String
		}

		results = append(results, nh)
	}

	return results, rows.Err()
}

// CountNoHitters returns the total number of no-hitters matching the filter.
func (r *AchievementRepository) CountNoHitters(ctx context.Context, filter core.AchievementFilter) (int, error) {
	where, args := buildAchievementWhere(filter)
	query := "SELECT COUNT(*) FROM no_hitters"
	if where != "" {
		query += " WHERE " + where
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// ListCycles retrieves hitting for the cycle achievements.
func (r *AchievementRepository) ListCycles(ctx context.Context, filter core.AchievementFilter) ([]core.Cycle, error) {
	query, args := buildCyclesQuery(filter)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query cycles: %w", err)
	}
	defer rows.Close()

	var results []core.Cycle
	for rows.Next() {
		var c core.Cycle
		var date string

		err := rows.Scan(
			&c.GameID,
			&c.PlayerID,
			&c.TeamID,
			&date,
			&c.Season,
			&c.HomeTeam,
			&c.VisitingTeam,
			&c.Singles,
			&c.Doubles,
			&c.Triples,
			&c.HomeRuns,
			&c.TotalHits,
			&c.ParkID,
			&c.TeamLocation,
		)
		if err != nil {
			return nil, fmt.Errorf("scan cycle: %w", err)
		}

		parsedDate, err := time.Parse("20060102", date)
		if err != nil {
			return nil, fmt.Errorf("parse date %s: %w", date, err)
		}
		c.Date = parsedDate

		results = append(results, c)
	}

	return results, rows.Err()
}

// CountCycles returns the total number of cycles matching the filter.
func (r *AchievementRepository) CountCycles(ctx context.Context, filter core.AchievementFilter) (int, error) {
	where, args := buildAchievementWhere(filter)
	query := "SELECT COUNT(*) FROM cycles"
	if where != "" {
		query += " WHERE " + where
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// ListMultiHRGames retrieves games where a player hit 3+ home runs.
func (r *AchievementRepository) ListMultiHRGames(ctx context.Context, filter core.AchievementFilter) ([]core.MultiHRGame, error) {
	query, args := buildMultiHRGamesQuery(filter)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query multi-HR games: %w", err)
	}
	defer rows.Close()

	var results []core.MultiHRGame
	for rows.Next() {
		var mh core.MultiHRGame
		var date string

		err := rows.Scan(
			&mh.GameID,
			&mh.PlayerID,
			&mh.TeamID,
			&date,
			&mh.Season,
			&mh.HomeTeam,
			&mh.VisitingTeam,
			&mh.HomeRuns,
			&mh.TotalHits,
			&mh.AtBats,
			&mh.ParkID,
			&mh.TeamLocation,
		)
		if err != nil {
			return nil, fmt.Errorf("scan multi-HR game: %w", err)
		}

		parsedDate, err := time.Parse("20060102", date)
		if err != nil {
			return nil, fmt.Errorf("parse date %s: %w", date, err)
		}
		mh.Date = parsedDate

		results = append(results, mh)
	}

	return results, rows.Err()
}

// CountMultiHRGames returns the total number of multi-HR games matching the filter.
func (r *AchievementRepository) CountMultiHRGames(ctx context.Context, filter core.AchievementFilter) (int, error) {
	where, args := buildAchievementWhere(filter)
	query := "SELECT COUNT(*) FROM multi_hr_games"
	if where != "" {
		query += " WHERE " + where
	}

	if filter.MinHR != nil {
		if where != "" {
			query += " AND home_runs >= $" + fmt.Sprintf("%d", len(args)+1)
		} else {
			query += " WHERE home_runs >= $1"
		}
		args = append(args, *filter.MinHR)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// ListTriplePlays retrieves games where a team recorded one or more triple plays.
func (r *AchievementRepository) ListTriplePlays(ctx context.Context, filter core.AchievementFilter) ([]core.TriplePlay, error) {
	query, args := buildTriplePlaysQuery(filter)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query triple plays: %w", err)
	}
	defer rows.Close()

	var results []core.TriplePlay
	for rows.Next() {
		var tp core.TriplePlay
		var date string

		err := rows.Scan(
			&tp.GameID,
			&tp.TeamID,
			&tp.OpponentTeamID,
			&date,
			&tp.Season,
			&tp.HomeTeam,
			&tp.VisitingTeam,
			&tp.TeamScore,
			&tp.OpponentScore,
			&tp.TriplePlaysCount,
			&tp.TeamLocation,
			&tp.ParkID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan triple play: %w", err)
		}

		parsedDate, err := time.Parse("20060102", date)
		if err != nil {
			return nil, fmt.Errorf("parse date %s: %w", date, err)
		}
		tp.Date = parsedDate

		results = append(results, tp)
	}

	return results, rows.Err()
}

// CountTriplePlays returns the total number of triple plays matching the filter.
func (r *AchievementRepository) CountTriplePlays(ctx context.Context, filter core.AchievementFilter) (int, error) {
	where, args := buildAchievementWhere(filter)
	query := "SELECT COUNT(*) FROM triple_plays"
	if where != "" {
		query += " WHERE " + where
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// ListExtraInningGames retrieves games that lasted 20+ innings.
func (r *AchievementRepository) ListExtraInningGames(ctx context.Context, filter core.AchievementFilter) ([]core.ExtraInningGame, error) {
	query, args := buildExtraInningGamesQuery(filter)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query extra inning games: %w", err)
	}
	defer rows.Close()

	var results []core.ExtraInningGame
	for rows.Next() {
		var eg core.ExtraInningGame
		var date string
		var homeTeamLeague, visitingTeamLeague, gameTimeMinutes sql.NullString
		var winningTeam sql.NullString

		err := rows.Scan(
			&eg.GameID,
			&date,
			&eg.Season,
			&eg.HomeTeam,
			&eg.VisitingTeam,
			&homeTeamLeague,
			&visitingTeamLeague,
			&eg.HomeScore,
			&eg.VisitingScore,
			&eg.Innings,
			&eg.GameLengthOuts,
			&gameTimeMinutes,
			&eg.ParkID,
			&winningTeam,
			&eg.ResultType,
		)
		if err != nil {
			return nil, fmt.Errorf("scan extra inning game: %w", err)
		}

		parsedDate, err := time.Parse("20060102", date)
		if err != nil {
			return nil, fmt.Errorf("parse date %s: %w", date, err)
		}
		eg.Date = parsedDate

		if homeTeamLeague.Valid {
			eg.HomeTeamLeague = &homeTeamLeague.String
		}
		if visitingTeamLeague.Valid {
			eg.VisitingTeamLeague = &visitingTeamLeague.String
		}
		if gameTimeMinutes.Valid {
			var minutes int
			fmt.Sscanf(gameTimeMinutes.String, "%d", &minutes)
			eg.GameTimeMinutes = &minutes
		}
		if winningTeam.Valid {
			wt := core.TeamID(winningTeam.String)
			eg.WinningTeam = &wt
		}

		results = append(results, eg)
	}

	return results, rows.Err()
}

// CountExtraInningGames returns the total number of extra inning games matching the filter.
func (r *AchievementRepository) CountExtraInningGames(ctx context.Context, filter core.AchievementFilter) (int, error) {
	where, args := buildAchievementWhere(filter)
	query := "SELECT COUNT(*) FROM extra_inning_games"
	if where != "" {
		query += " WHERE " + where
	}

	if filter.MinInnings != nil {
		if where != "" {
			query += " AND innings >= $" + fmt.Sprintf("%d", len(args)+1)
		} else {
			query += " WHERE innings >= $1"
		}
		args = append(args, *filter.MinInnings)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func buildAchievementWhere(filter core.AchievementFilter) (string, []any) {
	var conditions []string
	var args []any
	argNum := 1

	if filter.Season != nil {
		conditions = append(conditions, fmt.Sprintf("season = $%d", argNum))
		args = append(args, *filter.Season)
		argNum++
	}

	if filter.SeasonFrom != nil {
		conditions = append(conditions, fmt.Sprintf("season >= $%d", argNum))
		args = append(args, *filter.SeasonFrom)
		argNum++
	}

	if filter.SeasonTo != nil {
		conditions = append(conditions, fmt.Sprintf("season <= $%d", argNum))
		args = append(args, *filter.SeasonTo)
		argNum++
	}

	if filter.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("team_id = $%d", argNum))
		args = append(args, *filter.TeamID)
		argNum++
	}

	if filter.PlayerID != nil {
		conditions = append(conditions, fmt.Sprintf("player_id = $%d", argNum))
		args = append(args, *filter.PlayerID)
		argNum++
	}

	if filter.DateFrom != nil {
		dateStr := filter.DateFrom.Format("20060102")
		conditions = append(conditions, fmt.Sprintf("date >= $%d", argNum))
		args = append(args, dateStr)
		argNum++
	}

	if filter.DateTo != nil {
		dateStr := filter.DateTo.Format("20060102")
		conditions = append(conditions, fmt.Sprintf("date <= $%d", argNum))
		args = append(args, dateStr)
		argNum++
	}

	if filter.ParkID != nil {
		conditions = append(conditions, fmt.Sprintf("park_id = $%d", argNum))
		args = append(args, *filter.ParkID)
		argNum++
	}

	return strings.Join(conditions, " AND "), args
}

func buildNoHittersQuery(filter core.AchievementFilter) (string, []any) {
	where, args := buildAchievementWhere(filter)

	query := `
		SELECT
			game_id,
			team_id,
			opponent_team_id,
			date,
			season,
			home_team,
			visiting_team,
			home_score,
			visiting_score,
			innings,
			park_id,
			team_location,
			winning_pitcher_id,
			winning_pitcher_name
		FROM no_hitters`

	if where != "" {
		query += " WHERE " + where
	}

	query += " ORDER BY date DESC"

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d",
			filter.Pagination.PerPage,
			(filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	return query, args
}

func buildCyclesQuery(filter core.AchievementFilter) (string, []any) {
	where, args := buildAchievementWhere(filter)

	query := `
		SELECT
			game_id,
			player_id,
			team_id,
			date,
			season,
			home_team,
			visiting_team,
			singles,
			doubles,
			triples,
			home_runs,
			total_hits,
			park_id,
			team_location
		FROM cycles`

	if where != "" {
		query += " WHERE " + where
	}

	query += " ORDER BY date DESC"

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d",
			filter.Pagination.PerPage,
			(filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	return query, args
}

func buildMultiHRGamesQuery(filter core.AchievementFilter) (string, []any) {
	where, args := buildAchievementWhere(filter)

	query := `
		SELECT
			game_id,
			player_id,
			team_id,
			date,
			season,
			home_team,
			visiting_team,
			home_runs,
			total_hits,
			at_bats,
			park_id,
			team_location
		FROM multi_hr_games`

	if where != "" {
		query += " WHERE " + where
	}

	if filter.MinHR != nil {
		if where != "" {
			query += fmt.Sprintf(" AND home_runs >= $%d", len(args)+1)
		} else {
			query += fmt.Sprintf(" WHERE home_runs >= $%d", len(args)+1)
		}
		args = append(args, *filter.MinHR)
	}

	query += " ORDER BY home_runs DESC, date DESC"

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d",
			filter.Pagination.PerPage,
			(filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	return query, args
}

func buildTriplePlaysQuery(filter core.AchievementFilter) (string, []any) {
	where, args := buildAchievementWhere(filter)

	query := `
		SELECT
			game_id,
			team_id,
			opponent_team_id,
			date,
			season,
			home_team,
			visiting_team,
			team_score,
			opponent_score,
			triple_plays_count,
			team_location,
			park_id
		FROM triple_plays`

	if where != "" {
		query += " WHERE " + where
	}

	query += " ORDER BY date DESC"

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d",
			filter.Pagination.PerPage,
			(filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	return query, args
}

func buildExtraInningGamesQuery(filter core.AchievementFilter) (string, []any) {
	where, args := buildAchievementWhere(filter)

	query := `
		SELECT
			game_id,
			date,
			season,
			home_team,
			visiting_team,
			home_team_league,
			visiting_team_league,
			home_score,
			visiting_score,
			innings,
			game_length_outs,
			game_time_minutes,
			park_id,
			winning_team,
			result_type
		FROM extra_inning_games`

	if where != "" {
		query += " WHERE " + where
	}

	if filter.MinInnings != nil {
		if where != "" {
			query += fmt.Sprintf(" AND innings >= $%d", len(args)+1)
		} else {
			query += fmt.Sprintf(" WHERE innings >= $%d", len(args)+1)
		}
		args = append(args, *filter.MinInnings)
	}

	query += " ORDER BY innings DESC, date DESC"

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d",
			filter.Pagination.PerPage,
			(filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	return query, args
}
