-- Create materialized view for per-game team statistics
-- This enables fast queries for team game logs and daily performance tracking
-- Coverage: All games in plays table (1910-2025)

DROP MATERIALIZED VIEW IF EXISTS team_game_stats CASCADE;

CREATE MATERIALIZED VIEW team_game_stats AS
WITH batting_stats AS (
    -- Aggregate offensive stats when team is batting
    SELECT
        p.gid as game_id,
        p.batteam as team_id,
        g.date,
        CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) as season,
        -- Batting stats
        SUM(p.pa) as pa,
        SUM(p.ab) as ab,
        SUM(p.single + p.double + p.triple + p.hr) as h,
        SUM(p.single) as singles,
        SUM(p.double) as doubles,
        SUM(p.triple) as triples,
        SUM(p.hr) as hr,
        SUM(CASE WHEN p.run_b IS NOT NULL AND p.run_b != '' THEN 1 ELSE 0 END +
            CASE WHEN p.run1 IS NOT NULL AND p.run1 != '' THEN 1 ELSE 0 END +
            CASE WHEN p.run2 IS NOT NULL AND p.run2 != '' THEN 1 ELSE 0 END +
            CASE WHEN p.run3 IS NOT NULL AND p.run3 != '' THEN 1 ELSE 0 END) as runs_scored,
        SUM(CASE
            WHEN p.ab > 0 AND (p.single + p.double + p.triple + p.hr) = 0
                 AND p.sf = 0 AND p.sh = 0 AND p.walk = 0
            THEN 1
            ELSE 0
        END) as outs_made,
        SUM(p.walk) as bb,
        SUM(p.k) as so,
        SUM(p.hbp) as hbp,
        SUM(p.sf) as sf,
        SUM(p.sh) as sh,
        SUM(p.sb2 + p.sb3 + COALESCE(p.sbh, 0)) as sb,
        SUM(p.cs2 + p.cs3 + COALESCE(p.csh, 0)) as cs,
        SUM(p.iw) as ibb,
        SUM(p.gdp) as gdp
    FROM plays p
    JOIN games g ON p.gid = g.game_id
    WHERE p.batteam IS NOT NULL AND p.batteam != ''
    GROUP BY p.gid, p.batteam, g.date
),
pitching_stats AS (
    -- Aggregate pitching/defensive stats when team is pitching
    SELECT
        p.gid as game_id,
        p.pitteam as team_id,
        -- Pitching stats (opponent batting stats)
        SUM(p.ab) as ab_against,
        SUM(p.single + p.double + p.triple + p.hr) as h_against,
        SUM(p.hr) as hr_against,
        SUM(p.walk) as bb_against,
        SUM(p.k) as so_against,
        SUM(CASE WHEN p.run_b IS NOT NULL AND p.run_b != '' THEN 1 ELSE 0 END +
            CASE WHEN p.run1 IS NOT NULL AND p.run1 != '' THEN 1 ELSE 0 END +
            CASE WHEN p.run2 IS NOT NULL AND p.run2 != '' THEN 1 ELSE 0 END +
            CASE WHEN p.run3 IS NOT NULL AND p.run3 != '' THEN 1 ELSE 0 END) as runs_allowed,
        -- Fielding stats
        SUM(p.e1 + p.e2 + p.e3 + p.e4 + p.e5 + p.e6 + p.e7 + p.e8 + p.e9) as errors,
        SUM(CASE
            WHEN p.ab > 0 AND (p.single + p.double + p.triple + p.hr) = 0
                 AND p.sf = 0 AND p.sh = 0 AND p.walk = 0
            THEN 1
            ELSE 0
        END) as outs_recorded
    FROM plays p
    WHERE p.pitteam IS NOT NULL AND p.pitteam != ''
    GROUP BY p.gid, p.pitteam
)
SELECT
    b.game_id,
    b.team_id,
    b.date,
    b.season,
    -- Batting stats
    b.pa,
    b.ab,
    b.h,
    b.singles,
    b.doubles,
    b.triples,
    b.hr,
    b.runs_scored,
    b.bb,
    b.so,
    b.hbp,
    b.sf,
    b.sh,
    b.sb,
    b.cs,
    b.ibb,
    b.gdp,
    -- Pitching/defensive stats
    COALESCE(p.ab_against, 0) as ab_against,
    COALESCE(p.h_against, 0) as h_against,
    COALESCE(p.hr_against, 0) as hr_against,
    COALESCE(p.bb_against, 0) as bb_against,
    COALESCE(p.so_against, 0) as so_against,
    COALESCE(p.runs_allowed, 0) as runs_allowed,
    COALESCE(p.errors, 0) as errors,
    -- Calculated fields
    CASE
        WHEN b.ab > 0
        THEN ROUND(b.h::numeric / b.ab::numeric, 3)
        ELSE 0
    END as avg,
    CASE
        WHEN b.ab + b.bb + b.hbp + b.sf > 0
        THEN ROUND((b.h + b.bb + b.hbp)::numeric / (b.ab + b.bb + b.hbp + b.sf)::numeric, 3)
        ELSE 0
    END as obp,
    CASE
        WHEN b.ab > 0
        THEN ROUND((b.singles + 2*b.doubles + 3*b.triples + 4*b.hr)::numeric / b.ab::numeric, 3)
        ELSE 0
    END as slg,
    -- Game result
    CASE
        WHEN b.runs_scored > COALESCE(p.runs_allowed, 0) THEN 'W'
        WHEN b.runs_scored < COALESCE(p.runs_allowed, 0) THEN 'L'
        ELSE 'T'
    END as result
FROM batting_stats b
LEFT JOIN pitching_stats p ON b.game_id = p.game_id AND b.team_id = p.team_id;

COMMENT ON MATERIALIZED VIEW team_game_stats IS
'Per-game team statistics aggregated from play-by-play data.
Includes both offensive (batting) and defensive (pitching/fielding) stats.
Enables daily performance tracking and rolling aggregate queries.
Refresh after loading new plays data: REFRESH MATERIALIZED VIEW CONCURRENTLY team_game_stats;';
