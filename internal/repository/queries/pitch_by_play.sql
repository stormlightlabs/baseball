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
WHERE p.gid = $1 AND p.pn = $2 AND p.pitches IS NOT NULL
