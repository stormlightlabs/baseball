package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

// StandingsRepository serves historical and current-season standings.
type StandingsRepository struct {
	db *sql.DB
}

func NewStandingsRepository(db *sql.DB) *StandingsRepository {
	return &StandingsRepository{db: db}
}

func (r *StandingsRepository) SeasonStandings(ctx context.Context, season core.SeasonYear) ([]core.SeasonStanding, *time.Time, error) {
	useCurrentSeason, err := r.shouldUseCurrentSeasonStandings(ctx, season)
	if err != nil {
		return nil, nil, err
	}
	if useCurrentSeason {
		standings, lastUpdated, err := r.currentSeasonStandings(ctx, season)
		if err != nil {
			return nil, nil, err
		}
		if len(standings) == 0 {
			return nil, nil, core.NewNotFoundError("standings", fmt.Sprintf("%d", season))
		}
		return standings, lastUpdated, nil
	}

	standings, err := r.historicalStandings(ctx, season)
	if err != nil {
		return nil, nil, err
	}
	if len(standings) == 0 {
		return nil, nil, core.NewNotFoundError("standings", fmt.Sprintf("%d", season))
	}
	return standings, nil, nil
}

func (r *StandingsRepository) shouldUseCurrentSeasonStandings(ctx context.Context, season core.SeasonYear) (bool, error) {
	var hasCurrent bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM current_season.standings WHERE season = $1 LIMIT 1)`, int(season)).Scan(&hasCurrent); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "current_season") {
			return false, nil
		}
		return false, fmt.Errorf("failed checking current-season standings season=%d: %w", season, err)
	}
	if !hasCurrent {
		return false, nil
	}
	if int(season) >= currentYear() {
		return true, nil
	}

	var hasHistorical bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM "Teams" WHERE "yearID" = $1 LIMIT 1)`, int(season)).Scan(&hasHistorical); err != nil {
		return false, fmt.Errorf("failed checking historical standings season=%d: %w", season, err)
	}
	return !hasHistorical, nil
}

func (r *StandingsRepository) currentSeasonStandings(ctx context.Context, season core.SeasonYear) ([]core.SeasonStanding, *time.Time, error) {
	query := `
		SELECT
			cs.season,
			cs.division_id,
			cs.division_name,
			COALESCE(tm.local_league, ''),
			COALESCE(cs.team_id, tm.local_team_id),
			COALESCE(cs.franchise_id, tm.local_franchise_id),
			tm.local_team_name,
			cs.team_mlb_id,
			COALESCE(cs.w, 0),
			COALESCE(cs.l, 0),
			cs.pct,
			cs.gb,
			cs.wc_gb,
			cs.streak,
			cs.l10,
			cs.run_diff,
			cs.rs,
			cs.ra,
			cs.fetched_at
		FROM current_season.standings cs
		LEFT JOIN team_mlbam_map tm
			ON tm.season = cs.season
			AND tm.mlbam_team_id = cs.team_mlb_id
		WHERE cs.season = $1
	`

	rows, err := r.db.QueryContext(ctx, query, int(season))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query current-season standings: %w", err)
	}
	defer rows.Close()

	standings := []core.SeasonStanding{}
	var lastUpdated *time.Time

	for rows.Next() {
		var s core.SeasonStanding
		var league sql.NullString
		var teamID sql.NullString
		var franchiseID sql.NullString
		var teamName sql.NullString
		var teamMLBID int
		var pct sql.NullFloat64
		var gb sql.NullString
		var wcgb sql.NullString
		var streak sql.NullString
		var l10 sql.NullString
		var runDiff sql.NullInt64
		var rs sql.NullInt64
		var ra sql.NullInt64
		var fetchedAt time.Time

		if err := rows.Scan(
			&s.Season,
			&s.DivisionID,
			&s.DivisionName,
			&league,
			&teamID,
			&franchiseID,
			&teamName,
			&teamMLBID,
			&s.W,
			&s.L,
			&pct,
			&gb,
			&wcgb,
			&streak,
			&l10,
			&runDiff,
			&rs,
			&ra,
			&fetchedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("failed to scan current-season standings row: %w", err)
		}

		if league.Valid && league.String != "" {
			lg := core.LeagueID(league.String)
			s.League = &lg
		}
		if teamID.Valid && teamID.String != "" {
			tid := core.TeamID(teamID.String)
			s.TeamID = &tid
		}
		if franchiseID.Valid && franchiseID.String != "" {
			fid := core.FranchiseID(franchiseID.String)
			s.FranchiseID = &fid
		}
		if teamName.Valid && teamName.String != "" {
			tn := teamName.String
			s.TeamName = &tn
		}
		mlbID := teamMLBID
		s.TeamMLBID = &mlbID
		if pct.Valid {
			value := pct.Float64
			s.PCT = &value
		}
		if gb.Valid {
			value := gb.String
			s.GB = &value
		}
		if wcgb.Valid {
			value := wcgb.String
			s.WCGB = &value
		}
		if streak.Valid {
			value := streak.String
			s.Streak = &value
		}
		if l10.Valid {
			value := l10.String
			s.L10 = &value
		}
		if runDiff.Valid {
			value := int(runDiff.Int64)
			s.RunDiff = &value
		}
		if rs.Valid {
			value := int(rs.Int64)
			s.RS = &value
		}
		if ra.Valid {
			value := int(ra.Int64)
			s.RA = &value
		}
		s.Source = "current_season"
		s.FetchedAt = &fetchedAt
		if lastUpdated == nil || fetchedAt.After(*lastUpdated) {
			copy := fetchedAt
			lastUpdated = &copy
		}
		standings = append(standings, s)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed iterating current-season standings rows: %w", err)
	}

	sort.SliceStable(standings, func(i, j int) bool {
		if standings[i].DivisionName == standings[j].DivisionName {
			if standings[i].W == standings[j].W {
				return standings[i].L < standings[j].L
			}
			return standings[i].W > standings[j].W
		}
		return standings[i].DivisionName < standings[j].DivisionName
	})

	return standings, lastUpdated, nil
}

func (r *StandingsRepository) historicalStandings(ctx context.Context, season core.SeasonYear) ([]core.SeasonStanding, error) {
	query := `
		SELECT
			"yearID",
			"lgID",
			"divID",
			"Rank",
			"teamID",
			"franchID",
			"name",
			COALESCE("W", 0),
			COALESCE("L", 0),
			COALESCE("R", 0),
			COALESCE("RA", 0)
		FROM "Teams"
		WHERE "yearID" = $1
	`

	rows, err := r.db.QueryContext(ctx, query, int(season))
	if err != nil {
		return nil, fmt.Errorf("failed to query historical standings: %w", err)
	}
	defer rows.Close()

	standings := []core.SeasonStanding{}
	for rows.Next() {
		var year int
		var league string
		var division sql.NullString
		var rank sql.NullInt64
		var teamID string
		var franchiseID string
		var teamName string
		var w int
		var l int
		var rs int
		var ra int

		if err := rows.Scan(
			&year,
			&league,
			&division,
			&rank,
			&teamID,
			&franchiseID,
			&teamName,
			&w,
			&l,
			&rs,
			&ra,
		); err != nil {
			return nil, fmt.Errorf("failed to scan historical standings row: %w", err)
		}

		divisionName := strings.TrimSpace(fmt.Sprintf("%s %s", league, division.String))
		if divisionName == "" {
			divisionName = league
		}
		divisionID := 0
		if rank.Valid {
			divisionID = int(rank.Int64)
		}
		lg := core.LeagueID(league)
		tid := core.TeamID(teamID)
		fid := core.FranchiseID(franchiseID)
		tn := teamName
		standing := core.SeasonStanding{
			Season:       core.SeasonYear(year),
			DivisionID:   divisionID,
			DivisionName: divisionName,
			League:       &lg,
			TeamID:       &tid,
			FranchiseID:  &fid,
			TeamName:     &tn,
			W:            w,
			L:            l,
			Source:       "lahman",
		}
		games := w + l
		if games > 0 {
			pct := float64(w) / float64(games)
			standing.PCT = &pct
		}
		rd := rs - ra
		standing.RunDiff = &rd
		standing.RS = &rs
		standing.RA = &ra
		standings = append(standings, standing)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating historical standings rows: %w", err)
	}

	sort.SliceStable(standings, func(i, j int) bool {
		li := ""
		if standings[i].League != nil {
			li = string(*standings[i].League)
		}
		lj := ""
		if standings[j].League != nil {
			lj = string(*standings[j].League)
		}
		if li == lj {
			if standings[i].DivisionName == standings[j].DivisionName {
				if standings[i].W == standings[j].W {
					return standings[i].L < standings[j].L
				}
				return standings[i].W > standings[j].W
			}
			return standings[i].DivisionName < standings[j].DivisionName
		}
		return li < lj
	})

	return standings, nil
}
