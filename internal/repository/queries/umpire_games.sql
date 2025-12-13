SELECT
	date,
	game_number,
	visiting_team,
	home_team,
	visiting_team_league,
	home_team_league,
	visiting_score,
	home_score,
	game_length_outs,
	day_of_week,
	attendance,
	game_time_minutes,
	park_id,
	hp_ump_id,
	b1_ump_id,
	b2_ump_id,
	b3_ump_id
FROM games
WHERE (hp_ump_id = $1 OR b1_ump_id = $1 OR b2_ump_id = $1 OR b3_ump_id = $1 OR lf_ump_id = $1 OR rf_ump_id = $1)
