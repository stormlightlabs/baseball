-- Create materialized view for per-game batting statistics
-- This enables fast queries for player game logs and "game finder" functionality
-- Coverage: All games in plays table (1910-2025)

DROP MATERIALIZED VIEW IF EXISTS player_game_batting_stats CASCADE;

CREATE MATERIALIZED VIEW player_game_batting_stats AS
SELECT
    p.batter as player_id,
    p.gid as game_id,
    g.date,
    CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) as season,
    p.batteam as team_id,
    -- Basic counting stats
    SUM(p.pa) as pa,
    SUM(p.ab) as ab,
    SUM(p.single + p.double + p.triple + p.hr) as h,
    SUM(p.single) as singles,
    SUM(p.double) as doubles,
    SUM(p.triple) as triples,
    SUM(p.hr) as hr,
    SUM(p.runs) as r,
    SUM(p.rbi) as rbi,
    SUM(p.walk) as bb,
    SUM(p.k) as so,
    SUM(p.hbp) as hbp,
    SUM(p.sf) as sf,
    SUM(p.sh) as sh,
    SUM(p.sb2 + p.sb3 + COALESCE(p.sbh, 0)) as sb,
    SUM(p.cs2 + p.cs3 + COALESCE(p.csh, 0)) as cs,
    SUM(p.iw) as ibb,
    SUM(p.gdp) as gdp,
    -- Calculated rate stats
    CASE
        WHEN SUM(p.ab) > 0
        THEN ROUND(SUM(p.single + p.double + p.triple + p.hr)::numeric / SUM(p.ab)::numeric, 3)
        ELSE 0
    END as avg,
    CASE
        WHEN SUM(p.ab + p.walk + p.hbp + p.sf) > 0
        THEN ROUND((SUM(p.single + p.double + p.triple + p.hr + p.walk + p.hbp)::numeric) /
                   (SUM(p.ab + p.walk + p.hbp + p.sf)::numeric), 3)
        ELSE 0
    END as obp,
    CASE
        WHEN SUM(p.ab) > 0
        THEN ROUND((SUM(p.single + 2*p.double + 3*p.triple + 4*p.hr)::numeric) / SUM(p.ab)::numeric, 3)
        ELSE 0
    END as slg
FROM plays p
JOIN games g ON p.gid = g.game_id
WHERE p.batter IS NOT NULL
GROUP BY p.batter, p.gid, g.date, p.batteam;

COMMENT ON MATERIALIZED VIEW player_game_batting_stats IS
'Per-game batting statistics aggregated from play-by-play data.
Enables fast player game log queries and game finder functionality.
Refresh after loading new plays data: REFRESH MATERIALIZED VIEW CONCURRENTLY player_game_batting_stats;';
