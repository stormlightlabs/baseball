-- game log and game-stat materialized views.
-- Views are created WITH NO DATA and refreshed via ETL/db refresh-views.


-- SECTION 016_player_game_batting_stats_view.sql

-- Create materialized view for per-game batting statistics
-- This enables fast queries for player game logs and "game finder" functionality
-- Coverage: All games in plays table (1910-2025)

DROP MATERIALIZED VIEW IF EXISTS player_game_batting_stats CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS player_game_batting_stats AS
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
GROUP BY p.batter, p.gid, g.date, p.batteam
WITH NO DATA;

COMMENT ON MATERIALIZED VIEW player_game_batting_stats IS
'Per-game batting statistics aggregated from play-by-play data.
Enables fast player game log queries and game finder functionality.
Refresh after loading new plays data: REFRESH MATERIALIZED VIEW CONCURRENTLY player_game_batting_stats;';



-- SECTION 017_player_game_batting_stats_indexes.sql

-- Create indexes for player_game_batting_stats materialized view
-- Optimizes common query patterns: player lookups, game lookups, date ranges, season filters

-- Primary lookup indexes
CREATE INDEX IF NOT EXISTS idx_player_game_batting_player_id ON player_game_batting_stats(player_id);
CREATE INDEX IF NOT EXISTS idx_player_game_batting_game_id ON player_game_batting_stats(game_id);
CREATE INDEX IF NOT EXISTS idx_player_game_batting_date ON player_game_batting_stats(date);
CREATE INDEX IF NOT EXISTS idx_player_game_batting_season ON player_game_batting_stats(season);
CREATE INDEX IF NOT EXISTS idx_player_game_batting_team ON player_game_batting_stats(team_id);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_player_game_batting_player_season ON player_game_batting_stats(player_id, season);
CREATE INDEX IF NOT EXISTS idx_player_game_batting_player_date ON player_game_batting_stats(player_id, date);
CREATE INDEX IF NOT EXISTS idx_player_game_batting_team_date ON player_game_batting_stats(team_id, date);
CREATE INDEX IF NOT EXISTS idx_player_game_batting_season_date ON player_game_batting_stats(season, date);

-- Game finder indexes (for queries like "games with 2+ HR")
CREATE INDEX IF NOT EXISTS idx_player_game_batting_hr ON player_game_batting_stats(hr) WHERE hr > 0;
CREATE INDEX IF NOT EXISTS idx_player_game_batting_h ON player_game_batting_stats(h) WHERE h >= 3;
CREATE INDEX IF NOT EXISTS idx_player_game_batting_rbi ON player_game_batting_stats(rbi) WHERE rbi > 0;
CREATE INDEX IF NOT EXISTS idx_player_game_batting_sb ON player_game_batting_stats(sb) WHERE sb > 0;

COMMENT ON INDEX idx_player_game_batting_player_id IS 'Fast lookup of all games for a specific player';
COMMENT ON INDEX idx_player_game_batting_player_season IS 'Fast lookup of player season game logs';
COMMENT ON INDEX idx_player_game_batting_hr IS 'Game finder: games with home runs';



-- SECTION 018_player_game_pitching_stats_view.sql

-- Create materialized view for per-game pitching statistics
-- This enables fast queries for pitcher game logs and "game finder" functionality
-- Coverage: All games in plays table (1910-2025)

DROP MATERIALIZED VIEW IF EXISTS player_game_pitching_stats CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS player_game_pitching_stats AS
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
GROUP BY p.pitcher, p.gid, g.date, p.pitteam
WITH NO DATA;

COMMENT ON MATERIALIZED VIEW player_game_pitching_stats IS
'Per-game pitching statistics aggregated from play-by-play data.
Enables fast pitcher game log queries and game finder functionality.
Refresh after loading new plays data: REFRESH MATERIALIZED VIEW CONCURRENTLY player_game_pitching_stats;';



-- SECTION 019_player_game_pitching_stats_indexes.sql

-- Create indexes for player_game_pitching_stats materialized view
-- Optimizes common query patterns: pitcher lookups, game lookups, date ranges, season filters

-- Primary lookup indexes
CREATE INDEX IF NOT EXISTS idx_player_game_pitching_player_id ON player_game_pitching_stats(player_id);
CREATE INDEX IF NOT EXISTS idx_player_game_pitching_game_id ON player_game_pitching_stats(game_id);
CREATE INDEX IF NOT EXISTS idx_player_game_pitching_date ON player_game_pitching_stats(date);
CREATE INDEX IF NOT EXISTS idx_player_game_pitching_season ON player_game_pitching_stats(season);
CREATE INDEX IF NOT EXISTS idx_player_game_pitching_team ON player_game_pitching_stats(team_id);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_player_game_pitching_player_season ON player_game_pitching_stats(player_id, season);
CREATE INDEX IF NOT EXISTS idx_player_game_pitching_player_date ON player_game_pitching_stats(player_id, date);
CREATE INDEX IF NOT EXISTS idx_player_game_pitching_team_date ON player_game_pitching_stats(team_id, date);
CREATE INDEX IF NOT EXISTS idx_player_game_pitching_season_date ON player_game_pitching_stats(season, date);

-- Game finder indexes (for queries like "games with 10+ strikeouts")
CREATE INDEX IF NOT EXISTS idx_player_game_pitching_so ON player_game_pitching_stats(so) WHERE so >= 10;
CREATE INDEX IF NOT EXISTS idx_player_game_pitching_era ON player_game_pitching_stats(era) WHERE era > 0;
CREATE INDEX IF NOT EXISTS idx_player_game_pitching_ip ON player_game_pitching_stats(ip) WHERE ip >= 5;

COMMENT ON INDEX idx_player_game_pitching_player_id IS 'Fast lookup of all games for a specific pitcher';
COMMENT ON INDEX idx_player_game_pitching_player_season IS 'Fast lookup of pitcher season game logs';
COMMENT ON INDEX idx_player_game_pitching_so IS 'Game finder: games with 10+ strikeouts';



-- SECTION 020_player_game_fielding_stats_view.sql

-- Create materialized view for per-game fielding statistics by position
-- This enables fast queries for player fielding logs and "game finder" functionality
-- Coverage: All games in plays table (1910-2025)
-- Position codes: 1=P, 2=C, 3=1B, 4=2B, 5=3B, 6=SS, 7=LF, 8=CF, 9=RF

DROP MATERIALIZED VIEW IF EXISTS player_game_fielding_stats CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS player_game_fielding_stats AS
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
WHERE fp.fielder IS NOT NULL
WITH NO DATA;

COMMENT ON MATERIALIZED VIEW player_game_fielding_stats IS
'Per-game fielding statistics by position aggregated from play-by-play data.
Each row represents a player''s fielding performance at a specific position in a game.
Enables fast fielding game log queries and game finder functionality.
Refresh after loading new plays data: REFRESH MATERIALIZED VIEW CONCURRENTLY player_game_fielding_stats;';



-- SECTION 021_player_game_fielding_stats_indexes.sql

-- Create indexes for player_game_fielding_stats materialized view
-- Optimizes common query patterns: player lookups, game lookups, date ranges, season filters, position filters

-- Primary lookup indexes
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_player_id ON player_game_fielding_stats(player_id);
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_game_id ON player_game_fielding_stats(game_id);
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_date ON player_game_fielding_stats(date);
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_season ON player_game_fielding_stats(season);
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_team ON player_game_fielding_stats(team_id);
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_position ON player_game_fielding_stats(position);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_player_season ON player_game_fielding_stats(player_id, season);
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_player_position ON player_game_fielding_stats(player_id, position);
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_player_date ON player_game_fielding_stats(player_id, date);
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_team_date ON player_game_fielding_stats(team_id, date);
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_season_position ON player_game_fielding_stats(season, position);

-- Game finder indexes (for queries like "games with 3+ errors")
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_errors ON player_game_fielding_stats(e) WHERE e > 0;
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_assists ON player_game_fielding_stats(a) WHERE a >= 5;
CREATE INDEX IF NOT EXISTS idx_player_game_fielding_putouts ON player_game_fielding_stats(po) WHERE po >= 10;

COMMENT ON INDEX idx_player_game_fielding_player_id IS 'Fast lookup of all games for a specific player';
COMMENT ON INDEX idx_player_game_fielding_player_position IS 'Fast lookup of player games at specific position';
COMMENT ON INDEX idx_player_game_fielding_errors IS 'Game finder: games with errors';



-- SECTION 022_team_game_stats_view.sql

-- Create materialized view for per-game team statistics
-- This enables fast queries for team game logs and daily performance tracking
-- Coverage: All games in plays table (1910-2025)

DROP MATERIALIZED VIEW IF EXISTS team_game_stats CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS team_game_stats AS
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
LEFT JOIN pitching_stats p ON b.game_id = p.game_id AND b.team_id = p.team_id
WITH NO DATA;

COMMENT ON MATERIALIZED VIEW team_game_stats IS
'Per-game team statistics aggregated from play-by-play data.
Includes both offensive (batting) and defensive (pitching/fielding) stats.
Enables daily performance tracking and rolling aggregate queries.
Refresh after loading new plays data: REFRESH MATERIALIZED VIEW CONCURRENTLY team_game_stats;';



-- SECTION 023_team_game_stats_indexes.sql

-- Create indexes for team_game_stats materialized view
-- These indexes support common query patterns for team daily stats API

-- Primary lookup: team + date range queries
CREATE INDEX IF NOT EXISTS idx_team_game_stats_team_date
ON team_game_stats(team_id, date DESC);

-- Season-based queries
CREATE INDEX IF NOT EXISTS idx_team_game_stats_team_season
ON team_game_stats(team_id, season DESC);

-- Game lookup for joining with games table
CREATE INDEX IF NOT EXISTS idx_team_game_stats_game_id
ON team_game_stats(game_id);

-- Date-based queries for league-wide daily stats
CREATE INDEX IF NOT EXISTS idx_team_game_stats_date
ON team_game_stats(date DESC);

-- Season queries for leaderboards
CREATE INDEX IF NOT EXISTS idx_team_game_stats_season
ON team_game_stats(season DESC);

-- Support concurrent refresh
CREATE UNIQUE INDEX IF NOT EXISTS idx_team_game_stats_unique
ON team_game_stats(game_id, team_id);

COMMENT ON INDEX idx_team_game_stats_team_date IS
'Primary index for team daily stats queries by team and date range';

COMMENT ON INDEX idx_team_game_stats_team_season IS
'Supports season-based team performance queries';

COMMENT ON INDEX idx_team_game_stats_unique IS
'Unique constraint to support REFRESH MATERIALIZED VIEW CONCURRENTLY';

