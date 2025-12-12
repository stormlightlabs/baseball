WITH allstar_games AS (
	SELECT DISTINCT "yearID", "gameNum", "gameID"
	FROM "AllstarFull"
)
SELECT
	ag."yearID",
	ag."gameNum",
	ag."gameID",
	g.date,
	g.park_id,
	g.visiting_team,
	g.home_team,
	g.visiting_team_league,
	g.home_team_league,
	g.visiting_score,
	g.home_score,
	g.game_length_outs,
	g.day_of_week,
	g.attendance,
	g.game_time_minutes
FROM allstar_games ag
JOIN games g
	ON ag."gameID" = g.home_team || g.date || COALESCE(g.game_number::text, '0')
WHERE g.game_type = 'allstar'
