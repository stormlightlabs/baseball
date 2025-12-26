package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"stormlightlabs.org/baseball/internal/cache"
	"stormlightlabs.org/baseball/internal/core"
)

//go:embed queries/award_results_list.sql
var awardResultsListQuery string

//go:embed queries/allstar_games_list.sql
var allstarGamesListQuery string

//go:embed queries/allstar_game_get.sql
var allstarGameGetQuery string

type AwardRepository struct {
	db    *sql.DB
	cache *cache.CachedRepository
}

func NewAwardRepository(db *sql.DB, cacheClient *cache.Client) *AwardRepository {
	return &AwardRepository{
		db:    db,
		cache: cache.NewCachedRepository(cacheClient, "award"),
	}
}

func (r *AwardRepository) ListAwards(ctx context.Context) ([]core.Award, error) {
	query := `
		SELECT DISTINCT "awardID"
		FROM "AwardsPlayers"
		ORDER BY "awardID"
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list awards: %w", err)
	}
	defer rows.Close()

	var awards []core.Award
	for rows.Next() {
		var awardID string
		if err := rows.Scan(&awardID); err != nil {
			return nil, fmt.Errorf("failed to scan award: %w", err)
		}

		name := awardID
		description := ""
		switch awardID {
		case "MVP":
			name = "Most Valuable Player"
			description = "Awarded to the best player in each league"
		case "Cy Young Award":
			name = "Cy Young Award"
			description = "Awarded to the best pitcher in each league"
		case "Rookie of the Year":
			name = "Rookie of the Year"
			description = "Awarded to the best rookie in each league"
		case "Gold Glove":
			name = "Gold Glove Award"
			description = "Awarded to the best defensive player at each position"
		case "Silver Slugger":
			name = "Silver Slugger Award"
			description = "Awarded to the best offensive player at each position"
		case "TSN All-Star":
			name = "The Sporting News All-Star"
			description = "The Sporting News All-Star selection"
		case "World Series MVP":
			name = "World Series MVP"
			description = "Most Valuable Player of the World Series"
		case "All-Star Game MVP":
			name = "All-Star Game MVP"
			description = "Most Valuable Player of the All-Star Game"
		case "Babe Ruth Award":
			name = "Babe Ruth Award"
			description = "Awarded to the best postseason performer"
		case "Hank Aaron Award":
			name = "Hank Aaron Award"
			description = "Awarded to the best overall offensive performer"
		case "Roberto Clemente Award":
			name = "Roberto Clemente Award"
			description = "Awarded for sportsmanship and community involvement"
		}

		awards = append(awards, core.Award{
			ID:          core.AwardID(awardID),
			Name:        name,
			Description: description,
		})
	}

	return awards, nil
}

func (r *AwardRepository) ListAwardResults(ctx context.Context, filter core.AwardFilter) ([]core.AwardResult, error) {
	params := awardFilterToParams(filter)
	var cached []core.AwardResult
	if r.cache.List.Get(ctx, params, &cached) {
		return cached, nil
	}

	query := awardResultsListQuery

	args := []any{}
	argNum := 1

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND ap.\"playerID\" = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}

	if filter.AwardID != nil {
		query += fmt.Sprintf(" AND ap.\"awardID\" = $%d", argNum)
		args = append(args, string(*filter.AwardID))
		argNum++
	}

	if filter.Year != nil {
		query += fmt.Sprintf(" AND ap.\"yearID\" = $%d", argNum)
		args = append(args, int(*filter.Year))
		argNum++
	}

	if filter.League != nil {
		query += fmt.Sprintf(" AND ap.\"lgID\" = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	query += " ORDER BY ap.\"yearID\" DESC"

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list award results: %w", err)
	}
	defer rows.Close()

	var results []core.AwardResult
	for rows.Next() {
		var ar core.AwardResult
		var league sql.NullString
		var votesFirst, points sql.NullFloat64

		err := rows.Scan(
			&ar.AwardID, &ar.PlayerID, &ar.Year, &league,
			&votesFirst, &points,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan award result: %w", err)
		}

		if league.Valid {
			lg := core.LeagueID(league.String)
			ar.League = &lg
		}

		if votesFirst.Valid {
			vf := int(votesFirst.Float64)
			ar.VotesFirst = &vf
		}

		if points.Valid {
			p := int(points.Float64)
			ar.Points = &p
		}

		results = append(results, ar)
	}

	_ = r.cache.List.Set(ctx, params, results)
	return results, nil
}

func (r *AwardRepository) CountAwardResults(ctx context.Context, filter core.AwardFilter) (int, error) {
	query := `SELECT COUNT(*) FROM "AwardsPlayers" ap WHERE 1=1`
	args := []any{}
	argNum := 1

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND ap.\"playerID\" = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}

	if filter.AwardID != nil {
		query += fmt.Sprintf(" AND ap.\"awardID\" = $%d", argNum)
		args = append(args, string(*filter.AwardID))
		argNum++
	}

	if filter.Year != nil {
		query += fmt.Sprintf(" AND ap.\"yearID\" = $%d", argNum)
		args = append(args, int(*filter.Year))
		argNum++
	}

	if filter.League != nil {
		query += fmt.Sprintf(" AND ap.\"lgID\" = $%d", argNum)
		args = append(args, string(*filter.League))
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *AwardRepository) HallOfFameByPlayer(ctx context.Context, id core.PlayerID) ([]core.HallOfFameRecord, error) {
	var cached []core.HallOfFameRecord
	if r.cache.Entity.Get(ctx, string(id)+":hof", &cached) {
		return cached, nil
	}

	query := `
		SELECT
			"playerID", "yearid", "votedBy", "ballots", "needed", "votes", "inducted"
		FROM "HallOfFame"
		WHERE "playerID" = $1
		ORDER BY "yearid" DESC
	`

	rows, err := r.db.QueryContext(ctx, query, string(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get hall of fame records: %w", err)
	}
	defer rows.Close()

	var records []core.HallOfFameRecord
	for rows.Next() {
		var hof core.HallOfFameRecord
		var ballots, needed, votes sql.NullInt64
		var inducted sql.NullString

		err := rows.Scan(
			&hof.PlayerID, &hof.Year, &hof.VotedBy,
			&ballots, &needed, &votes, &inducted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan hall of fame record: %w", err)
		}

		if ballots.Valid {
			b := int(ballots.Int64)
			hof.Ballots = &b
		}

		if needed.Valid {
			n := int(needed.Int64)
			hof.Needed = &n
		}

		if votes.Valid {
			v := int(votes.Int64)
			hof.Votes = &v
		}

		if inducted.Valid {
			hof.Inducted = inducted.String == "Y"
		}

		records = append(records, hof)
	}

	_ = r.cache.Entity.Set(ctx, string(id)+":hof", records)
	return records, nil
}

// ListAllStarGames returns all all-star games, optionally filtered by year.
func (r *AwardRepository) ListAllStarGames(ctx context.Context, year *core.SeasonYear) ([]core.AllStarGame, error) {
	query := allstarGamesListQuery
	args := []any{}
	argNum := 1

	if year != nil {
		query += fmt.Sprintf(" AND ag.\"yearID\" = $%d", argNum)
		args = append(args, int(*year))
		argNum++
	}

	query += " ORDER BY ag.\"yearID\" DESC, ag.\"gameNum\""

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list all-star games: %w", err)
	}
	defer rows.Close()

	var games []core.AllStarGame
	for rows.Next() {
		var game core.AllStarGame
		var gameID sql.NullString
		var gameNum sql.NullInt64
		var date sql.NullString
		var park sql.NullString
		var visitingTeam, homeTeam sql.NullString
		var visitingLeague, homeLeague sql.NullString
		var visitingScore, homeScore sql.NullInt64
		var gameLengthOuts sql.NullInt64
		var dayOfWeek sql.NullString
		var attendance, duration sql.NullInt64

		err := rows.Scan(
			&game.Year,
			&gameNum,
			&gameID,
			&date,
			&park,
			&visitingTeam,
			&homeTeam,
			&visitingLeague,
			&homeLeague,
			&visitingScore,
			&homeScore,
			&gameLengthOuts,
			&dayOfWeek,
			&attendance,
			&duration,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan all-star game: %w", err)
		}

		if gameNum.Valid {
			game.GameNum = int(gameNum.Int64)
		}

		if gameID.Valid {
			game.GameID = gameID.String
		}

		if err := attachRetrosheetDataToAllStarGame(
			&game,
			date,
			park,
			visitingTeam,
			homeTeam,
			visitingLeague,
			homeLeague,
			visitingScore,
			homeScore,
			gameLengthOuts,
			dayOfWeek,
			attendance,
			duration,
		); err != nil {
			return nil, err
		}

		games = append(games, game)
	}

	return games, nil
}

// GetAllStarGame returns details for a specific all-star game.
func (r *AwardRepository) GetAllStarGame(ctx context.Context, gameID string) (*core.AllStarGame, error) {
	var cached core.AllStarGame
	if r.cache.Entity.Get(ctx, gameID, &cached) {
		return &cached, nil
	}

	query := allstarGameGetQuery

	var game core.AllStarGame
	var dbGameID sql.NullString
	var gameNum sql.NullInt64
	var date sql.NullString
	var park sql.NullString
	var visitingTeam, homeTeam sql.NullString
	var visitingLeague, homeLeague sql.NullString
	var visitingScore, homeScore sql.NullInt64
	var gameLengthOuts sql.NullInt64
	var dayOfWeek sql.NullString
	var attendance, duration sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, gameID).Scan(
		&game.Year,
		&gameNum,
		&dbGameID,
		&date,
		&park,
		&visitingTeam,
		&homeTeam,
		&visitingLeague,
		&homeLeague,
		&visitingScore,
		&homeScore,
		&gameLengthOuts,
		&dayOfWeek,
		&attendance,
		&duration,
	)
	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("all-star game", "")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get all-star game: %w", err)
	}

	if gameNum.Valid {
		game.GameNum = int(gameNum.Int64)
	}

	if dbGameID.Valid {
		game.GameID = dbGameID.String
	}

	if err := attachRetrosheetDataToAllStarGame(
		&game,
		date,
		park,
		visitingTeam,
		homeTeam,
		visitingLeague,
		homeLeague,
		visitingScore,
		homeScore,
		gameLengthOuts,
		dayOfWeek,
		attendance,
		duration,
	); err != nil {
		return nil, err
	}

	participantsQuery := `
		SELECT
			"playerID", "yearID", "gameNum", "gameID",
			"teamID", "lgID", "GP", "startingPos"
		FROM "AllstarFull"
		WHERE "gameID" = $1
		ORDER BY "startingPos" NULLS LAST, "playerID"
	`

	rows, err := r.db.QueryContext(ctx, participantsQuery, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all-star participants: %w", err)
	}
	defer rows.Close()

	var participants []core.AllStarAppearance
	for rows.Next() {
		var p core.AllStarAppearance
		var teamID, league sql.NullString
		var gameNum, gp, startingPos sql.NullInt64

		err := rows.Scan(
			&p.PlayerID, &p.Year, &gameNum, &p.GameID,
			&teamID, &league, &gp, &startingPos,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		if gameNum.Valid {
			p.GameNum = int(gameNum.Int64)
		}

		if teamID.Valid {
			tid := core.TeamID(teamID.String)
			p.TeamID = &tid
		}

		if league.Valid {
			lg := core.LeagueID(league.String)
			p.League = &lg
		}

		if gp.Valid {
			gpInt := int(gp.Int64)
			p.GP = &gpInt
		}

		if startingPos.Valid {
			spInt := int(startingPos.Int64)
			p.StartingPos = &spInt
		}

		participants = append(participants, p)
	}

	game.Participants = participants
	_ = r.cache.Entity.Set(ctx, gameID, &game)
	return &game, nil
}

func attachRetrosheetDataToAllStarGame(
	game *core.AllStarGame,
	date sql.NullString,
	park sql.NullString,
	visitingTeam sql.NullString,
	homeTeam sql.NullString,
	visitingLeague sql.NullString,
	homeLeague sql.NullString,
	visitingScore sql.NullInt64,
	homeScore sql.NullInt64,
	gameLengthOuts sql.NullInt64,
	dayOfWeek sql.NullString,
	attendance sql.NullInt64,
	duration sql.NullInt64,
) error {
	var retro core.Game

	if game.GameID != "" {
		retro.ID = core.GameID(game.GameID)
	}

	if date.Valid {
		parsed, err := time.Parse("20060102", date.String)
		if err != nil {
			return fmt.Errorf("failed to parse all-star game date: %w", err)
		}
		game.Date = &parsed
		retro.Date = parsed
		retro.Season = core.SeasonYear(parsed.Year())
	} else if game.Year != 0 {
		retro.Season = game.Year
	}

	if park.Valid && park.String != "" {
		pid := core.ParkID(park.String)
		game.Venue = &pid
		retro.ParkID = pid
	}

	if visitingTeam.Valid {
		retro.AwayTeam = core.TeamID(visitingTeam.String)
	}
	if homeTeam.Valid {
		retro.HomeTeam = core.TeamID(homeTeam.String)
	}
	if visitingLeague.Valid {
		retro.AwayLeague = core.LeagueID(visitingLeague.String)
	}
	if homeLeague.Valid {
		retro.HomeLeague = core.LeagueID(homeLeague.String)
	}
	if visitingScore.Valid {
		retro.AwayScore = int(visitingScore.Int64)
	}
	if homeScore.Valid {
		retro.HomeScore = int(homeScore.Int64)
	}
	if gameLengthOuts.Valid {
		retro.Innings = int(gameLengthOuts.Int64) / 3
	}
	if dayOfWeek.Valid {
		retro.DayOfWeek = dayOfWeek.String
	}
	if attendance.Valid {
		a := int(attendance.Int64)
		retro.Attendance = &a
	}
	if duration.Valid {
		d := int(duration.Int64)
		retro.DurationMin = &d
	}

	retro.IsPostseason = false
	game.RetrosheetGame = &retro

	return nil
}

// awardFilterToParams converts AwardFilter to cache param map
func awardFilterToParams(filter core.AwardFilter) map[string]string {
	params := make(map[string]string)

	if filter.PlayerID != nil {
		params["player_id"] = string(*filter.PlayerID)
	}
	if filter.AwardID != nil {
		params["award_id"] = string(*filter.AwardID)
	}
	if filter.Year != nil {
		params["year"] = strconv.Itoa(int(*filter.Year))
	}
	if filter.League != nil {
		params["league"] = string(*filter.League)
	}
	params["page"] = strconv.Itoa(filter.Pagination.Page)
	params["per_page"] = strconv.Itoa(filter.Pagination.PerPage)

	return params
}
