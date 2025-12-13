SELECT
	gid, pn, inning, top_bot, batteam, pitteam,
	SUBSTRING(gid, 4, 8) as date,
	CASE
		WHEN SUBSTRING(gid, 12, 1) = '0' THEN 'regular'
		WHEN SUBSTRING(gid, 12, 1) = '1' THEN 'postseason'
		ELSE 'other'
	END as game_type,
	batter, pitcher, bathand, pithand,
	score_v, score_h, outs_pre, outs_post,
	balls, strikes, pitches,
	event
FROM plays
WHERE gid = $1 AND pn = $2 AND pitches IS NOT NULL
