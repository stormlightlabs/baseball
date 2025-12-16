SELECT
	p.gid, p.pn, p.inning, p.top_bot, p.batteam, p.pitteam, p.date, p.gametype,
	p.batter,
	(SELECT first_name || ' ' || last_name FROM retrosheet_players WHERE player_id = p.batter LIMIT 1) as batter_name,
	p.pitcher,
	(SELECT first_name || ' ' || last_name FROM retrosheet_players WHERE player_id = p.pitcher LIMIT 1) as pitcher_name,
	p.bathand, p.pithand,
	p.score_v, p.score_h, p.outs_pre, p.outs_post,
	p.balls, p.strikes, p.pitches,
	p.event,
	p.pa, p.ab, p.single, p.double, p.triple, p.hr, p.walk, p.k, p.hbp,
	p.br1_pre, p.br2_pre, p.br3_pre,
	p.runs, p.rbi
FROM plays p
WHERE 1=1
