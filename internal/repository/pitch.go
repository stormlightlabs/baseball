package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

//go:embed queries/pitch_by_play.sql
var pitchByPlayQuery string

type PitchRepository struct {
	db *sql.DB
}

func NewPitchRepository(db *sql.DB) *PitchRepository {
	return &PitchRepository{db: db}
}

// parsePitchSequence converts a Retrosheet pitch sequence string into individual Pitch records.
// The sequence uses single characters to represent pitch types and events.
func (r *PitchRepository) parsePitchSequence(play core.Play) []core.Pitch {
	if play.Pitches == nil || *play.Pitches == "" {
		return nil
	}

	sequence := *play.Pitches
	pitches := []core.Pitch{}
	balls := 0
	strikes := 0

	for i, ch := range sequence {
		pitchChar := string(ch)

		if pitchChar == "." || pitchChar == ">" || pitchChar == "*" ||
			pitchChar == "+" || pitchChar == "1" || pitchChar == "2" || pitchChar == "3" {
			continue
		}

		pitch := core.Pitch{
			GameID:      play.GameID,
			PlayNum:     play.PlayNum,
			Inning:      play.Inning,
			TopBot:      play.TopBot,
			BatTeam:     play.BatTeam,
			PitTeam:     play.PitTeam,
			Date:        play.Date,
			Batter:      play.Batter,
			BatterName:  play.BatterName,
			Pitcher:     play.Pitcher,
			PitcherName: play.PitcherName,
			BatHand:     play.BatHand,
			PitHand:     play.PitHand,
			OutsPre:     play.OutsPre,
			SeqNum:      len(pitches) + 1,
			PitchType:   pitchChar,
			BallCount:   balls,
			StrikeCount: strikes,
		}

		switch pitchChar {
		case "B": // Ball
			pitch.IsBall = true
			pitch.Description = "Ball"
			balls++
		case "C": // Called strike
			pitch.IsStrike = true
			pitch.Description = "Called strike"
			strikes++
		case "F": // Foul ball
			pitch.IsStrike = true
			pitch.Description = "Foul ball"
			if strikes < 2 {
				strikes++
			}
		case "S": // Swinging strike
			pitch.IsStrike = true
			pitch.Description = "Swinging strike"
			strikes++
		case "X": // Ball in play
			pitch.IsInPlay = true
			pitch.Description = "Ball in play"
		case "L": // Foul bunt
			pitch.IsStrike = true
			pitch.Description = "Foul bunt"
			if strikes < 2 {
				strikes++
			}
		case "M": // Missed bunt attempt
			pitch.IsStrike = true
			pitch.Description = "Missed bunt attempt"
			strikes++
		case "O": // Foul ball on bunt
			pitch.IsStrike = true
			pitch.Description = "Foul ball on bunt"
			if strikes < 2 {
				strikes++
			}
		case "P": // Pitchout
			pitch.IsBall = true
			pitch.Description = "Pitchout"
			balls++
		case "T": // Foul tip
			pitch.IsStrike = true
			pitch.Description = "Foul tip"
			strikes++
		case "V": // Called strike (on appeal)
			pitch.IsStrike = true
			pitch.Description = "Called strike (on appeal)"
			strikes++
		case "H": // Hit by pitch
			pitch.IsBall = true
			pitch.Description = "Hit by pitch"
		case "I": // Intentional ball
			pitch.IsBall = true
			pitch.Description = "Intentional ball"
			balls++
		case "N": // No pitch
			pitch.Description = "No pitch (balk, interference, etc.)"
			// Don't increment counts for no pitch
			continue
		case "A": // Automatic ball
			pitch.IsBall = true
			pitch.Description = "Automatic ball"
			balls++
		default:
			pitch.Description = fmt.Sprintf("Unknown pitch type: %s", pitchChar)
		}

		if i == len(sequence)-1 || (i < len(sequence)-1 &&
			(sequence[i+1] == '.' || sequence[i+1] == '>' || sequence[i+1] == '*' ||
				sequence[i+1] == '+' || sequence[i+1] == '1' || sequence[i+1] == '2' ||
				sequence[i+1] == '3')) {

			nextPitchExists := false
			for j := i + 1; j < len(sequence); j++ {
				ch := sequence[j]
				if ch != '.' && ch != '>' && ch != '*' && ch != '+' &&
					ch != '1' && ch != '2' && ch != '3' {
					nextPitchExists = true
					break
				}
			}
			if !nextPitchExists {
				pitch.Event = &play.Event
			}
		}

		pitches = append(pitches, pitch)
	}

	return pitches
}

// List retrieves pitches based on filter criteria
func (r *PitchRepository) List(ctx context.Context, filter core.PitchFilter) ([]core.Pitch, error) {
	query := `
		SELECT
			p.gid, p.pn, p.inning, p.top_bot, p.batteam, p.pitteam,
			SUBSTRING(p.gid, 4, 8) as date,
			CASE
				WHEN SUBSTRING(p.gid, 12, 1) = '0' THEN 'regular'
				WHEN SUBSTRING(p.gid, 12, 1) = '1' THEN 'postseason'
				ELSE 'other'
			END as game_type,
			p.batter,
			(SELECT first_name || ' ' || last_name FROM retrosheet_players WHERE player_id = p.batter LIMIT 1) as batter_name,
			p.pitcher,
			(SELECT first_name || ' ' || last_name FROM retrosheet_players WHERE player_id = p.pitcher LIMIT 1) as pitcher_name,
			p.bathand, p.pithand,
			p.score_v, p.score_h, p.outs_pre, p.outs_post,
			p.balls, p.strikes, p.pitches,
			p.event
		FROM plays p
		WHERE p.pitches IS NOT NULL AND p.pitches != ''
	`

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

	if filter.Date != nil {
		query += fmt.Sprintf(" AND SUBSTRING(gid, 4, 8) = $%d", argNum)
		args = append(args, *filter.Date)
		argNum++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND SUBSTRING(gid, 4, 8) >= $%d", argNum)
		args = append(args, *filter.DateFrom)
		argNum++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND SUBSTRING(gid, 4, 8) <= $%d", argNum)
		args = append(args, *filter.DateTo)
		argNum++
	}

	if filter.Inning != nil {
		query += fmt.Sprintf(" AND inning = $%d", argNum)
		args = append(args, *filter.Inning)
		argNum++
	}

	if filter.TopBot != nil {
		query += fmt.Sprintf(" AND top_bot = $%d", argNum)
		args = append(args, *filter.TopBot)
		argNum++
	}

	if filter.PitchType != nil {
		query += fmt.Sprintf(" AND pitches LIKE $%d", argNum)
		args = append(args, "%"+*filter.PitchType+"%")
		argNum++
	}

	query += " ORDER BY gid, pn"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query plays for pitches: %w", err)
	}
	defer rows.Close()

	var allPitches []core.Pitch

	for rows.Next() {
		var play core.Play
		var batterName, pitcherName, batHand, pitHand sql.NullString
		var balls, strikes sql.NullInt64
		var pitches sql.NullString

		err := rows.Scan(
			&play.GameID, &play.PlayNum, &play.Inning, &play.TopBot, &play.BatTeam, &play.PitTeam,
			&play.Date, &play.GameType,
			&play.Batter, &batterName,
			&play.Pitcher, &pitcherName,
			&batHand, &pitHand,
			&play.ScoreVis, &play.ScoreHome, &play.OutsPre, &play.OutsPost,
			&balls, &strikes, &pitches,
			&play.Event,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan play: %w", err)
		}

		if batterName.Valid {
			play.BatterName = &batterName.String
		}
		if pitcherName.Valid {
			play.PitcherName = &pitcherName.String
		}
		if batHand.Valid {
			play.BatHand = &batHand.String
		}
		if pitHand.Valid {
			play.PitHand = &pitHand.String
		}
		if balls.Valid {
			b := int(balls.Int64)
			play.Balls = &b
		}
		if strikes.Valid {
			s := int(strikes.Int64)
			play.Strikes = &s
		}
		if pitches.Valid {
			play.Pitches = &pitches.String
		}

		playPitches := r.parsePitchSequence(play)

		for _, pitch := range playPitches {
			if !r.matchesPitchFilter(pitch, filter) {
				continue
			}
			allPitches = append(allPitches, pitch)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plays: %w", err)
	}

	start := (filter.Pagination.Page - 1) * filter.Pagination.PerPage
	end := start + filter.Pagination.PerPage
	if start >= len(allPitches) {
		return []core.Pitch{}, nil
	}
	if end > len(allPitches) {
		end = len(allPitches)
	}

	return allPitches[start:end], nil
}

// matchesPitchFilter checks if a pitch matches the filter criteria
func (r *PitchRepository) matchesPitchFilter(pitch core.Pitch, filter core.PitchFilter) bool {
	if filter.TopBot != nil && pitch.TopBot != *filter.TopBot {
		return false
	}
	if filter.PitchType != nil && pitch.PitchType != *filter.PitchType {
		return false
	}
	if filter.BallCount != nil && pitch.BallCount != *filter.BallCount {
		return false
	}
	if filter.StrikeCount != nil && pitch.StrikeCount != *filter.StrikeCount {
		return false
	}
	if filter.IsInPlay != nil && pitch.IsInPlay != *filter.IsInPlay {
		return false
	}
	if filter.IsStrike != nil && pitch.IsStrike != *filter.IsStrike {
		return false
	}
	if filter.IsBall != nil && pitch.IsBall != *filter.IsBall {
		return false
	}
	return true
}

// Count returns the count of pitches matching the filter
func (r *PitchRepository) Count(ctx context.Context, filter core.PitchFilter) (int, error) {
	pitches, err := r.List(ctx, filter)
	if err != nil {
		return 0, err
	}
	return len(pitches), nil
}

// ListByGame retrieves all pitches for a specific game
func (r *PitchRepository) ListByGame(ctx context.Context, gameID core.GameID, p core.Pagination) ([]core.Pitch, error) {
	filter := core.PitchFilter{
		GameID:     &gameID,
		Pagination: p,
	}
	return r.List(ctx, filter)
}

// ListByPlay retrieves all pitches from a specific plate appearance
func (r *PitchRepository) ListByPlay(ctx context.Context, gameID core.GameID, playNum int) ([]core.Pitch, error) {
	var play core.Play
	var batterName, pitcherName, batHand, pitHand sql.NullString
	var balls, strikes sql.NullInt64
	var pitches sql.NullString

	err := r.db.QueryRowContext(ctx, pitchByPlayQuery, string(gameID), playNum).Scan(
		&play.GameID, &play.PlayNum, &play.Inning, &play.TopBot, &play.BatTeam, &play.PitTeam,
		&play.Date, &play.GameType,
		&play.Batter, &batterName,
		&play.Pitcher, &pitcherName,
		&batHand, &pitHand,
		&play.ScoreVis, &play.ScoreHome, &play.OutsPre, &play.OutsPost,
		&balls, &strikes, &pitches,
		&play.Event,
	)
	if err == sql.ErrNoRows {
		return []core.Pitch{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query play: %w", err)
	}

	if batterName.Valid {
		play.BatterName = &batterName.String
	}
	if pitcherName.Valid {
		play.PitcherName = &pitcherName.String
	}
	if batHand.Valid {
		play.BatHand = &batHand.String
	}
	if pitHand.Valid {
		play.PitHand = &pitHand.String
	}
	if balls.Valid {
		b := int(balls.Int64)
		play.Balls = &b
	}
	if strikes.Valid {
		s := int(strikes.Int64)
		play.Strikes = &s
	}
	if pitches.Valid {
		play.Pitches = &pitches.String
	}

	return r.parsePitchSequence(play), nil
}
