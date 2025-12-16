package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
	"stormlightlabs.org/baseball/internal/seed"
)

//go:embed queries/play_list.sql
var playListQuery string

//go:embed queries/play_list_by_player.sql
var playListByPlayerQuery string

type PlayRepository struct {
	db *sql.DB
}

func NewPlayRepository(db *sql.DB) *PlayRepository {
	return &PlayRepository{db: db}
}

// List retrieves plays based on filter criteria
func (r *PlayRepository) List(ctx context.Context, filter core.PlayFilter) ([]core.Play, error) {
	if filter.DateFrom == nil && filter.DateTo == nil {
		var leagues []string
		if len(filter.Leagues) > 0 {
			leagues = make([]string, len(filter.Leagues))
			for i, l := range filter.Leagues {
				leagues[i] = string(l)
			}
		} else if filter.League != nil {
			leagues = []string{string(*filter.League)}
		}

		if len(leagues) > 0 {
			if dateRange := seed.GetLeagueDateRange(leagues); dateRange != nil {
				filter.DateFrom = &dateRange.From
				filter.DateTo = &dateRange.To
			}
		}
	}

	query := playListQuery

	args := []any{}
	argNum := 1

	if filter.GameID != nil {
		query += fmt.Sprintf(" AND gid = $%d", argNum)
		args = append(args, string(*filter.GameID))
		argNum++
	}

	if filter.Batter != nil {
		query += fmt.Sprintf(" AND batter = $%d", argNum)
		args = append(args, string(*filter.Batter))
		argNum++
	}

	if filter.Pitcher != nil {
		query += fmt.Sprintf(" AND pitcher = $%d", argNum)
		args = append(args, string(*filter.Pitcher))
		argNum++
	}

	if filter.BatTeam != nil {
		query += fmt.Sprintf(" AND batteam = $%d", argNum)
		args = append(args, string(*filter.BatTeam))
		argNum++
	}

	if filter.PitTeam != nil {
		query += fmt.Sprintf(" AND pitteam = $%d", argNum)
		args = append(args, string(*filter.PitTeam))
		argNum++
	}

	if len(filter.Leagues) > 0 {
		query += fmt.Sprintf(" AND (home_team_league = ANY($%d) OR visiting_team_league = ANY($%d))", argNum, argNum+1)
		leagues := make([]string, len(filter.Leagues))
		for i, league := range filter.Leagues {
			leagues[i] = string(league)
		}
		args = append(args, leagues, leagues)
		argNum += 2
	} else if filter.League != nil {
		query += fmt.Sprintf(" AND (home_team_league = $%d OR visiting_team_league = $%d)", argNum, argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	if filter.GameType != nil {
		query += fmt.Sprintf(" AND gametype = $%d", argNum)
		args = append(args, *filter.GameType)
		argNum++
	}

	if filter.Date != nil {
		query += fmt.Sprintf(" AND date = $%d", argNum)
		args = append(args, *filter.Date)
		argNum++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND date >= $%d", argNum)
		args = append(args, *filter.DateFrom)
		argNum++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND date <= $%d", argNum)
		args = append(args, *filter.DateTo)
		argNum++
	}

	if filter.Inning != nil {
		query += fmt.Sprintf(" AND inning = $%d", argNum)
		args = append(args, *filter.Inning)
		argNum++
	}

	if filter.HomeRuns != nil && *filter.HomeRuns {
		query += " AND hr = 1"
	}

	if filter.Walks != nil && *filter.Walks {
		query += " AND walk = 1"
	}

	if filter.K != nil && *filter.K {
		query += " AND k = 1"
	}

	sortBy := "pn"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "ASC"
	if filter.SortOrder == core.SortDesc {
		sortOrder = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query plays: %w", err)
	}
	defer rows.Close()

	var plays []core.Play
	for rows.Next() {
		var p core.Play
		var batterName, pitcherName, batHand, pitHand sql.NullString
		var balls, strikes, pa, ab, single, double, triple, hr, walk, k, hbp, runs, rbi sql.NullInt64
		var pitches sql.NullString
		var br1Pre, br2Pre, br3Pre sql.NullString

		err := rows.Scan(
			&p.GameID, &p.PlayNum, &p.Inning, &p.TopBot, &p.BatTeam, &p.PitTeam, &p.Date, &p.GameType,
			&p.Batter, &batterName,
			&p.Pitcher, &pitcherName,
			&batHand, &pitHand,
			&p.ScoreVis, &p.ScoreHome, &p.OutsPre, &p.OutsPost,
			&balls, &strikes, &pitches,
			&p.Event,
			&pa, &ab, &single, &double, &triple, &hr, &walk, &k, &hbp,
			&br1Pre, &br2Pre, &br3Pre,
			&runs, &rbi,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan play: %w", err)
		}

		if batterName.Valid {
			p.BatterName = &batterName.String
		}
		if pitcherName.Valid {
			p.PitcherName = &pitcherName.String
		}
		if batHand.Valid {
			p.BatHand = &batHand.String
		}
		if pitHand.Valid {
			p.PitHand = &pitHand.String
		}
		if balls.Valid {
			b := int(balls.Int64)
			p.Balls = &b
		}
		if strikes.Valid {
			s := int(strikes.Int64)
			p.Strikes = &s
		}
		if pitches.Valid {
			p.Pitches = &pitches.String
		}
		if pa.Valid {
			v := int(pa.Int64)
			p.PA = &v
		}
		if ab.Valid {
			v := int(ab.Int64)
			p.AB = &v
		}
		if single.Valid {
			v := int(single.Int64)
			p.Single = &v
		}
		if double.Valid {
			v := int(double.Int64)
			p.Double = &v
		}
		if triple.Valid {
			v := int(triple.Int64)
			p.Triple = &v
		}
		if hr.Valid {
			v := int(hr.Int64)
			p.HR = &v
		}
		if walk.Valid {
			v := int(walk.Int64)
			p.Walk = &v
		}
		if k.Valid {
			v := int(k.Int64)
			p.K = &v
		}
		if hbp.Valid {
			v := int(hbp.Int64)
			p.HBP = &v
		}
		if br1Pre.Valid {
			r := core.RetroPlayerID(br1Pre.String)
			p.Runner1Pre = &r
		}
		if br2Pre.Valid {
			r := core.RetroPlayerID(br2Pre.String)
			p.Runner2Pre = &r
		}
		if br3Pre.Valid {
			r := core.RetroPlayerID(br3Pre.String)
			p.Runner3Pre = &r
		}
		if runs.Valid {
			v := int(runs.Int64)
			p.Runs = &v
		}
		if rbi.Valid {
			v := int(rbi.Int64)
			p.RBI = &v
		}

		plays = append(plays, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plays: %w", err)
	}

	return plays, nil
}

// Count returns the total number of plays matching the filter
func (r *PlayRepository) Count(ctx context.Context, filter core.PlayFilter) (int, error) {
	// Apply implicit date filters for league queries to enable partition pruning
	if filter.DateFrom == nil && filter.DateTo == nil {
		var leagues []string
		if len(filter.Leagues) > 0 {
			leagues = make([]string, len(filter.Leagues))
			for i, l := range filter.Leagues {
				leagues[i] = string(l)
			}
		} else if filter.League != nil {
			leagues = []string{string(*filter.League)}
		}

		if len(leagues) > 0 {
			if dateRange := seed.GetLeagueDateRange(leagues); dateRange != nil {
				filter.DateFrom = &dateRange.From
				filter.DateTo = &dateRange.To
			}
		}
	}

	query := `SELECT COUNT(*) FROM plays WHERE 1=1`

	args := []any{}
	argNum := 1

	if filter.GameID != nil {
		query += fmt.Sprintf(" AND gid = $%d", argNum)
		args = append(args, string(*filter.GameID))
		argNum++
	}

	if filter.Batter != nil {
		query += fmt.Sprintf(" AND batter = $%d", argNum)
		args = append(args, string(*filter.Batter))
		argNum++
	}

	if filter.Pitcher != nil {
		query += fmt.Sprintf(" AND pitcher = $%d", argNum)
		args = append(args, string(*filter.Pitcher))
		argNum++
	}

	if filter.BatTeam != nil {
		query += fmt.Sprintf(" AND batteam = $%d", argNum)
		args = append(args, string(*filter.BatTeam))
		argNum++
	}

	if filter.PitTeam != nil {
		query += fmt.Sprintf(" AND pitteam = $%d", argNum)
		args = append(args, string(*filter.PitTeam))
		argNum++
	}

	if len(filter.Leagues) > 0 {
		query += fmt.Sprintf(" AND (home_team_league = ANY($%d) OR visiting_team_league = ANY($%d))", argNum, argNum+1)
		leagues := make([]string, len(filter.Leagues))
		for i, league := range filter.Leagues {
			leagues[i] = string(league)
		}
		args = append(args, leagues, leagues)
		argNum += 2
	} else if filter.League != nil {
		query += fmt.Sprintf(" AND (home_team_league = $%d OR visiting_team_league = $%d)", argNum, argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	if filter.GameType != nil {
		query += fmt.Sprintf(" AND gametype = $%d", argNum)
		args = append(args, *filter.GameType)
		argNum++
	}

	if filter.Date != nil {
		query += fmt.Sprintf(" AND date = $%d", argNum)
		args = append(args, *filter.Date)
		argNum++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND date >= $%d", argNum)
		args = append(args, *filter.DateFrom)
		argNum++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND date <= $%d", argNum)
		args = append(args, *filter.DateTo)
		argNum++
	}

	if filter.Inning != nil {
		query += fmt.Sprintf(" AND inning = $%d", argNum)
		args = append(args, *filter.Inning)
	}

	if filter.HomeRuns != nil && *filter.HomeRuns {
		query += " AND hr = 1"
	}

	if filter.Walks != nil && *filter.Walks {
		query += " AND walk = 1"
	}

	if filter.K != nil && *filter.K {
		query += " AND k = 1"
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count plays: %w", err)
	}

	return count, nil
}

// ListByGame retrieves all plays for a specific game in order
func (r *PlayRepository) ListByGame(ctx context.Context, gameID core.GameID, p core.Pagination) ([]core.Play, error) {
	filter := core.PlayFilter{
		GameID:     &gameID,
		SortBy:     "pn",
		SortOrder:  core.SortAsc,
		Pagination: p,
	}
	return r.List(ctx, filter)
}

// ListByPlayer retrieves plays involving a specific player (as batter or pitcher)
func (r *PlayRepository) ListByPlayer(ctx context.Context, playerID core.RetroPlayerID, p core.Pagination) ([]core.Play, error) {
	query := playListByPlayerQuery

	rows, err := r.db.QueryContext(ctx, query, string(playerID), p.PerPage, (p.Page-1)*p.PerPage)
	if err != nil {
		return nil, fmt.Errorf("failed to query plays for player: %w", err)
	}
	defer rows.Close()

	var plays []core.Play
	for rows.Next() {
		var p core.Play
		var batterName, pitcherName, batHand, pitHand sql.NullString
		var balls, strikes, pa, ab, single, double, triple, hr, walk, k, hbp, runs, rbi sql.NullInt64
		var pitches sql.NullString
		var br1Pre, br2Pre, br3Pre sql.NullString

		err := rows.Scan(
			&p.GameID, &p.PlayNum, &p.Inning, &p.TopBot, &p.BatTeam, &p.PitTeam, &p.Date, &p.GameType,
			&p.Batter, &batterName, &p.Pitcher, &pitcherName, &batHand, &pitHand,
			&p.ScoreVis, &p.ScoreHome, &p.OutsPre, &p.OutsPost,
			&balls, &strikes, &pitches,
			&p.Event,
			&pa, &ab, &single, &double, &triple, &hr, &walk, &k, &hbp,
			&br1Pre, &br2Pre, &br3Pre,
			&runs, &rbi,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan play: %w", err)
		}

		if batterName.Valid {
			p.BatterName = &batterName.String
		}
		if pitcherName.Valid {
			p.PitcherName = &pitcherName.String
		}
		if batHand.Valid {
			p.BatHand = &batHand.String
		}
		if pitHand.Valid {
			p.PitHand = &pitHand.String
		}
		if balls.Valid {
			b := int(balls.Int64)
			p.Balls = &b
		}
		if strikes.Valid {
			s := int(strikes.Int64)
			p.Strikes = &s
		}
		if pitches.Valid {
			p.Pitches = &pitches.String
		}
		if pa.Valid {
			v := int(pa.Int64)
			p.PA = &v
		}
		if ab.Valid {
			v := int(ab.Int64)
			p.AB = &v
		}
		if single.Valid {
			v := int(single.Int64)
			p.Single = &v
		}
		if double.Valid {
			v := int(double.Int64)
			p.Double = &v
		}
		if triple.Valid {
			v := int(triple.Int64)
			p.Triple = &v
		}
		if hr.Valid {
			v := int(hr.Int64)
			p.HR = &v
		}
		if walk.Valid {
			v := int(walk.Int64)
			p.Walk = &v
		}
		if k.Valid {
			v := int(k.Int64)
			p.K = &v
		}
		if hbp.Valid {
			v := int(hbp.Int64)
			p.HBP = &v
		}
		if br1Pre.Valid {
			r := core.RetroPlayerID(br1Pre.String)
			p.Runner1Pre = &r
		}
		if br2Pre.Valid {
			r := core.RetroPlayerID(br2Pre.String)
			p.Runner2Pre = &r
		}
		if br3Pre.Valid {
			r := core.RetroPlayerID(br3Pre.String)
			p.Runner3Pre = &r
		}
		if runs.Valid {
			v := int(runs.Int64)
			p.Runs = &v
		}
		if rbi.Valid {
			v := int(rbi.Int64)
			p.RBI = &v
		}

		plays = append(plays, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plays: %w", err)
	}

	return plays, nil
}

// CountByPlayer returns the total number of plays where the player was batter or pitcher.
func (r *PlayRepository) CountByPlayer(ctx context.Context, playerID core.RetroPlayerID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM plays
		WHERE batter = $1 OR pitcher = $1
	`, string(playerID)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count plays for player: %w", err)
	}
	return count, nil
}
