package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

// CrosswalkRepository reads local and MLBAM identifier mappings.
type CrosswalkRepository struct {
	db *sql.DB
}

func NewCrosswalkRepository(db *sql.DB) *CrosswalkRepository {
	return &CrosswalkRepository{db: db}
}

func (r *CrosswalkRepository) ListTeamCrosswalk(ctx context.Context, filter core.TeamCrosswalkFilter) ([]core.TeamCrosswalkRow, error) {
	args := []any{}
	clauses := []string{"1=1"}
	argNum := 1

	if !filter.AllSeasons {
		season := filter.Season
		if season == nil {
			current := core.SeasonYear(currentYear())
			season = &current
		}
		clauses = append(clauses, fmt.Sprintf("base.season = $%d", argNum))
		args = append(args, int(*season))
		argNum++
	}

	if filter.TeamID != nil {
		clauses = append(clauses, fmt.Sprintf("base.team_id = $%d", argNum))
		args = append(args, string(*filter.TeamID))
		argNum++
	}
	if filter.FranchiseID != nil {
		clauses = append(clauses, fmt.Sprintf("base.franchise_id = $%d", argNum))
		args = append(args, string(*filter.FranchiseID))
		argNum++
	}
	if filter.MLBAMTeamID != nil {
		clauses = append(clauses, fmt.Sprintf("base.mlbam_team_id = $%d", argNum))
		args = append(args, *filter.MLBAMTeamID)
		argNum++
	}

	query := fmt.Sprintf(`
		WITH local_rows AS (
			SELECT
				tfm.season,
				tfm.team_id,
				tfm.franchise_id,
				tfm.team_name,
				tfm.league,
				NULL::int as mlbam_team_id,
				NULL::text as mlb_abbreviation,
				NULL::text as mlb_team_code,
				NULL::text as mlb_file_code,
				NULL::text as mlb_team_name,
				NULL::text as mlb_franchise_name,
				NULL::text as mlb_club_name,
				NULL::text as match_method,
				'local'::text as confidence,
				'local'::text as source
			FROM team_franchise_map tfm
		),
		mlbam_rows AS (
			SELECT
				tmm.season,
				tmm.local_team_id as team_id,
				tmm.local_franchise_id as franchise_id,
				tmm.local_team_name as team_name,
				tmm.local_league as league,
				tmm.mlbam_team_id,
				tmm.mlb_abbreviation,
				tmm.mlb_team_code,
				tmm.mlb_file_code,
				tmm.mlb_team_name,
				tmm.mlb_franchise_name,
				tmm.mlb_club_name,
				tmm.match_method,
				tmm.confidence,
				tmm.source
			FROM team_mlbam_map tmm
		),
		base AS (
			SELECT * FROM local_rows
			UNION ALL
			SELECT * FROM mlbam_rows
		)
		SELECT DISTINCT ON (
			base.season,
			COALESCE(base.team_id, ''),
			COALESCE(base.franchise_id, ''),
			COALESCE(base.mlbam_team_id, 0)
		)
			base.season,
			base.team_id,
			base.franchise_id,
			base.team_name,
			base.league,
			base.mlbam_team_id,
			base.mlb_abbreviation,
			base.mlb_team_code,
			base.mlb_file_code,
			base.mlb_team_name,
			base.mlb_franchise_name,
			base.mlb_club_name,
			base.match_method,
			base.confidence,
			base.source
		FROM base
		WHERE %s
		ORDER BY
			base.season DESC,
			COALESCE(base.team_id, ''),
			COALESCE(base.franchise_id, ''),
			COALESCE(base.mlbam_team_id, 0),
			CASE WHEN base.mlbam_team_id IS NULL THEN 1 ELSE 0 END,
			CASE base.confidence WHEN 'high' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END
	`, strings.Join(clauses, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list team crosswalk rows: %w", err)
	}
	defer rows.Close()

	result := []core.TeamCrosswalkRow{}
	for rows.Next() {
		var row core.TeamCrosswalkRow
		var teamID sql.NullString
		var franchiseID sql.NullString
		var teamName sql.NullString
		var league sql.NullString
		var mlbamTeamID sql.NullInt64
		var mlbAbbreviation sql.NullString
		var mlbTeamCode sql.NullString
		var mlbFileCode sql.NullString
		var mlbTeamName sql.NullString
		var mlbFranchiseName sql.NullString
		var mlbClubName sql.NullString
		var matchMethod sql.NullString
		var confidence sql.NullString
		var source sql.NullString
		if err := rows.Scan(
			&row.Season,
			&teamID,
			&franchiseID,
			&teamName,
			&league,
			&mlbamTeamID,
			&mlbAbbreviation,
			&mlbTeamCode,
			&mlbFileCode,
			&mlbTeamName,
			&mlbFranchiseName,
			&mlbClubName,
			&matchMethod,
			&confidence,
			&source,
		); err != nil {
			return nil, fmt.Errorf("failed to scan team crosswalk row: %w", err)
		}
		row.TeamName = teamName.String
		row.MLBAbbreviation = mlbAbbreviation.String
		row.MLBTeamCode = mlbTeamCode.String
		row.MLBFileCode = mlbFileCode.String
		row.MLBTeamName = mlbTeamName.String
		row.MLBFranchise = mlbFranchiseName.String
		row.MLBClubName = mlbClubName.String
		row.MatchMethod = matchMethod.String
		row.Confidence = confidence.String
		row.Source = source.String
		if teamID.Valid {
			t := core.TeamID(teamID.String)
			row.TeamID = &t
		}
		if franchiseID.Valid {
			f := core.FranchiseID(franchiseID.String)
			row.FranchiseID = &f
		}
		if league.Valid {
			lg := core.LeagueID(league.String)
			row.League = &lg
		}
		if mlbamTeamID.Valid {
			id := int(mlbamTeamID.Int64)
			row.MLBAMTeamID = &id
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating team crosswalk rows: %w", err)
	}
	return result, nil
}

func (r *CrosswalkRepository) ListPlayerCrosswalk(ctx context.Context, filter core.PlayerCrosswalkFilter) ([]core.PlayerCrosswalkRow, error) {
	args := []any{}
	clauses := []string{"1=1"}
	argNum := 1

	if filter.PlayerID != nil {
		clauses = append(clauses, fmt.Sprintf("base.player_id = $%d", argNum))
		args = append(args, string(*filter.PlayerID))
		argNum++
	}
	if filter.RetroID != nil {
		clauses = append(clauses, fmt.Sprintf("base.retro_id = $%d", argNum))
		args = append(args, string(*filter.RetroID))
		argNum++
	}
	if filter.MLBAMID != nil {
		clauses = append(clauses, fmt.Sprintf("base.mlbam_id = $%d", argNum))
		args = append(args, *filter.MLBAMID)
		argNum++
	}

	query := fmt.Sprintf(`
		WITH local_rows AS (
			SELECT
				pim.lahman_id as player_id,
				pim.retro_id,
				NULL::int as mlbam_id,
				NULL::text as bbref_id,
				pim.given_name as full_name,
				pim.first_name,
				pim.last_name,
				'local'::text as source,
				'local'::text as confidence
			FROM player_id_map pim
		),
		mlbam_rows AS (
			SELECT
				pmm.lahman_id as player_id,
				pmm.retro_id,
				pmm.mlbam_id,
				pmm.bbref_id,
				pmm.full_name,
				NULL::text as first_name,
				NULL::text as last_name,
				pmm.source,
				pmm.confidence
			FROM player_mlbam_map pmm
		),
		base AS (
			SELECT * FROM local_rows
			UNION ALL
			SELECT * FROM mlbam_rows
		)
		SELECT DISTINCT ON (
			COALESCE(base.player_id, ''),
			COALESCE(base.retro_id, ''),
			COALESCE(base.mlbam_id, 0)
		)
			base.player_id,
			base.retro_id,
			base.mlbam_id,
			base.bbref_id,
			base.full_name,
			base.first_name,
			base.last_name,
			base.source,
			base.confidence
		FROM base
		WHERE %s
		ORDER BY
			COALESCE(base.player_id, ''),
			COALESCE(base.retro_id, ''),
			COALESCE(base.mlbam_id, 0),
			CASE WHEN base.mlbam_id IS NULL THEN 1 ELSE 0 END,
			CASE base.confidence WHEN 'high' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END
	`, strings.Join(clauses, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list player crosswalk rows: %w", err)
	}
	defer rows.Close()

	result := []core.PlayerCrosswalkRow{}
	for rows.Next() {
		var row core.PlayerCrosswalkRow
		var playerID sql.NullString
		var retroID sql.NullString
		var mlbamID sql.NullInt64
		var bbrefID sql.NullString
		var fullName sql.NullString
		var firstName sql.NullString
		var lastName sql.NullString
		var source sql.NullString
		var confidence sql.NullString
		if err := rows.Scan(
			&playerID,
			&retroID,
			&mlbamID,
			&bbrefID,
			&fullName,
			&firstName,
			&lastName,
			&source,
			&confidence,
		); err != nil {
			return nil, fmt.Errorf("failed to scan player crosswalk row: %w", err)
		}
		row.BBRefID = bbrefID.String
		row.FullName = fullName.String
		row.FirstName = firstName.String
		row.LastName = lastName.String
		row.Source = source.String
		row.Confidence = confidence.String
		if playerID.Valid {
			p := core.PlayerID(playerID.String)
			row.PlayerID = &p
		}
		if retroID.Valid {
			rid := core.RetroPlayerID(retroID.String)
			row.RetroID = &rid
		}
		if mlbamID.Valid {
			id := int(mlbamID.Int64)
			row.MLBAMID = &id
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating player crosswalk rows: %w", err)
	}

	return result, nil
}

func currentYear() int {
	return time.Now().Year()
}
