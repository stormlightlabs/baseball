SELECT
	gid, pn, inning, top_bot, batteam, pitteam, date, gametype,
	batter, pitcher, bathand, pithand,
	score_v, score_h, outs_pre, outs_post,
	balls, strikes, pitches,
	event,
	pa, ab, single, double, triple, hr, walk, k, hbp,
	br1_pre, br2_pre, br3_pre,
	runs, rbi
FROM plays
WHERE 1=1
