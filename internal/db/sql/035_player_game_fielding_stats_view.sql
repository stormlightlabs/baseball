-- Create materialized view for per-game fielding statistics by position
-- This enables fast queries for player fielding logs and "game finder" functionality
-- Coverage: All games in plays table (1910-2025)
-- Position codes: 1=P, 2=C, 3=1B, 4=2B, 5=3B, 6=SS, 7=LF, 8=CF, 9=RF

DROP MATERIALIZED VIEW IF EXISTS player_game_fielding_stats CASCADE;

CREATE MATERIALIZED VIEW player_game_fielding_stats AS
WITH fielding_plays AS (
    -- Pitcher (position 1)
    SELECT
        gid,
        pitcher as fielder,
        pitteam as team_id,
        1 as position,
        SUM(COALESCE(po1, 0)) as po,
        SUM(COALESCE(a1, 0)) as a,
        SUM(COALESCE(e1, 0)) as e
    FROM plays
    WHERE pitcher IS NOT NULL
    GROUP BY gid, pitcher, pitteam
    UNION ALL
    -- Catcher (position 2)
    SELECT gid, f2, pitteam, 2, SUM(COALESCE(po2, 0)), SUM(COALESCE(a2, 0)), SUM(COALESCE(e2, 0))
    FROM plays WHERE f2 IS NOT NULL GROUP BY gid, f2, pitteam
    UNION ALL
    -- First Base (position 3)
    SELECT gid, f3, pitteam, 3, SUM(COALESCE(po3, 0)), SUM(COALESCE(a3, 0)), SUM(COALESCE(e3, 0))
    FROM plays WHERE f3 IS NOT NULL GROUP BY gid, f3, pitteam
    UNION ALL
    -- Second Base (position 4)
    SELECT gid, f4, pitteam, 4, SUM(COALESCE(po4, 0)), SUM(COALESCE(a4, 0)), SUM(COALESCE(e4, 0))
    FROM plays WHERE f4 IS NOT NULL GROUP BY gid, f4, pitteam
    UNION ALL
    -- Third Base (position 5)
    SELECT gid, f5, pitteam, 5, SUM(COALESCE(po5, 0)), SUM(COALESCE(a5, 0)), SUM(COALESCE(e5, 0))
    FROM plays WHERE f5 IS NOT NULL GROUP BY gid, f5, pitteam
    UNION ALL
    -- Shortstop (position 6)
    SELECT gid, f6, pitteam, 6, SUM(COALESCE(po6, 0)), SUM(COALESCE(a6, 0)), SUM(COALESCE(e6, 0))
    FROM plays WHERE f6 IS NOT NULL GROUP BY gid, f6, pitteam
    UNION ALL
    -- Left Field (position 7)
    SELECT gid, f7, pitteam, 7, SUM(COALESCE(po7, 0)), SUM(COALESCE(a7, 0)), SUM(COALESCE(e7, 0))
    FROM plays WHERE f7 IS NOT NULL GROUP BY gid, f7, pitteam
    UNION ALL
    -- Center Field (position 8)
    SELECT gid, f8, pitteam, 8, SUM(COALESCE(po8, 0)), SUM(COALESCE(a8, 0)), SUM(COALESCE(e8, 0))
    FROM plays WHERE f8 IS NOT NULL GROUP BY gid, f8, pitteam
    UNION ALL
    -- Right Field (position 9)
    SELECT gid, f9, pitteam, 9, SUM(COALESCE(po9, 0)), SUM(COALESCE(a9, 0)), SUM(COALESCE(e9, 0))
    FROM plays WHERE f9 IS NOT NULL GROUP BY gid, f9, pitteam
)
SELECT
    fp.fielder as player_id,
    fp.gid as game_id,
    g.date,
    CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) as season,
    fp.team_id,
    fp.position,
    fp.po,
    fp.a,
    fp.e,
    (fp.po + fp.a + fp.e) as tc,
    CASE
        WHEN (fp.po + fp.a + fp.e) > 0
        THEN ROUND((fp.po + fp.a)::numeric / (fp.po + fp.a + fp.e)::numeric, 3)
        ELSE 1.000
    END as fpct
FROM fielding_plays fp
JOIN games g ON fp.gid = g.game_id
WHERE fp.fielder IS NOT NULL;

COMMENT ON MATERIALIZED VIEW player_game_fielding_stats IS
'Per-game fielding statistics by position aggregated from play-by-play data.
Each row represents a player''s fielding performance at a specific position in a game.
Enables fast fielding game log queries and game finder functionality.
Refresh after loading new plays data: REFRESH MATERIALIZED VIEW CONCURRENTLY player_game_fielding_stats;';
