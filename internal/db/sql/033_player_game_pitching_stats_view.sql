-- Create materialized view for per-game pitching statistics
-- This enables fast queries for pitcher game logs and "game finder" functionality
-- Coverage: All games in plays table (1910-2025)

DROP MATERIALIZED VIEW IF EXISTS player_game_pitching_stats CASCADE;

CREATE MATERIALIZED VIEW player_game_pitching_stats AS
SELECT
    p.pitcher as player_id,
    p.gid as game_id,
    g.date,
    CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) as season,
    p.pitteam as team_id,
    -- Innings pitched (from outs recorded)
    ROUND(SUM(GREATEST(p.outs_post - p.outs_pre, 0))::numeric / 3.0, 1) as ip,
    -- Basic counting stats
    SUM(p.pa) as pa,
    SUM(p.ab) as ab,
    SUM(p.single + p.double + p.triple + p.hr) as h,
    SUM(p.runs) as r,
    SUM(p.er) as er,
    SUM(p.walk) as bb,
    SUM(p.k) as so,
    SUM(p.hr) as hr,
    SUM(p.hbp) as hbp,
    SUM(p.iw) as ibb,
    SUM(p.wp) as wp,
    SUM(p.bk) as bk,
    SUM(p.sh) as sh,
    SUM(p.sf) as sf,
    -- Calculated rate stats
    CASE
        WHEN SUM(GREATEST(p.outs_post - p.outs_pre, 0)) > 0
        THEN ROUND((SUM(p.er) * 27.0) / SUM(GREATEST(p.outs_post - p.outs_pre, 0))::numeric, 2)
        ELSE 0
    END as era,
    CASE
        WHEN SUM(GREATEST(p.outs_post - p.outs_pre, 0)) > 0
        THEN ROUND((SUM(p.single + p.double + p.triple + p.hr + p.walk)::numeric * 3.0) /
                   SUM(GREATEST(p.outs_post - p.outs_pre, 0))::numeric, 2)
        ELSE 0
    END as whip,
    CASE
        WHEN SUM(GREATEST(p.outs_post - p.outs_pre, 0)) > 0
        THEN ROUND((SUM(p.k) * 27.0) / SUM(GREATEST(p.outs_post - p.outs_pre, 0))::numeric, 2)
        ELSE 0
    END as k9,
    CASE
        WHEN SUM(GREATEST(p.outs_post - p.outs_pre, 0)) > 0
        THEN ROUND((SUM(p.walk) * 27.0) / SUM(GREATEST(p.outs_post - p.outs_pre, 0))::numeric, 2)
        ELSE 0
    END as bb9
FROM plays p
JOIN games g ON p.gid = g.game_id
WHERE p.pitcher IS NOT NULL
GROUP BY p.pitcher, p.gid, g.date, p.pitteam;

COMMENT ON MATERIALIZED VIEW player_game_pitching_stats IS
'Per-game pitching statistics aggregated from play-by-play data.
Enables fast pitcher game log queries and game finder functionality.
Refresh after loading new plays data: REFRESH MATERIALIZED VIEW CONCURRENTLY player_game_pitching_stats;';
