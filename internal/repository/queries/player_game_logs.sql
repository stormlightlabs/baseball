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
WHERE (
	v_player_1_id = $1 OR v_player_2_id = $1 OR v_player_3_id = $1 OR
	v_player_4_id = $1 OR v_player_5_id = $1 OR v_player_6_id = $1 OR
	v_player_7_id = $1 OR v_player_8_id = $1 OR v_player_9_id = $1 OR
	h_player_1_id = $1 OR h_player_2_id = $1 OR h_player_3_id = $1 OR
	h_player_4_id = $1 OR h_player_5_id = $1 OR h_player_6_id = $1 OR
	h_player_7_id = $1 OR h_player_8_id = $1 OR h_player_9_id = $1
)
