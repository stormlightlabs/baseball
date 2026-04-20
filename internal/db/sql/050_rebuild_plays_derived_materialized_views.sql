-- Rebuild plays-derived materialized views after plays table partition swap.
--
-- Migration 033 renamed 'plays' -> 'plays_old' and swapped in partitioned 'plays'.
-- Any materialized views created before that swap remained bound to the old relation OID
-- (now named plays_old). Reapplying these view definitions rebonds them to current public.plays.

-- BEGIN 016_player_game_batting_stats_view.sql
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
GROUP BY p.batter, p.gid, g.date, p.batteam;

COMMENT ON MATERIALIZED VIEW player_game_batting_stats IS
'Per-game batting statistics aggregated from play-by-play data.
Enables fast player game log queries and game finder functionality.
Refresh after loading new plays data: REFRESH MATERIALIZED VIEW CONCURRENTLY player_game_batting_stats;';

-- END 016_player_game_batting_stats_view.sql

-- BEGIN 017_player_game_batting_stats_indexes.sql
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

-- END 017_player_game_batting_stats_indexes.sql

-- BEGIN 018_player_game_pitching_stats_view.sql
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
GROUP BY p.pitcher, p.gid, g.date, p.pitteam;

COMMENT ON MATERIALIZED VIEW player_game_pitching_stats IS
'Per-game pitching statistics aggregated from play-by-play data.
Enables fast pitcher game log queries and game finder functionality.
Refresh after loading new plays data: REFRESH MATERIALIZED VIEW CONCURRENTLY player_game_pitching_stats;';

-- END 018_player_game_pitching_stats_view.sql

-- BEGIN 019_player_game_pitching_stats_indexes.sql
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

-- END 019_player_game_pitching_stats_indexes.sql

-- BEGIN 020_player_game_fielding_stats_view.sql
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
WHERE fp.fielder IS NOT NULL;

COMMENT ON MATERIALIZED VIEW player_game_fielding_stats IS
'Per-game fielding statistics by position aggregated from play-by-play data.
Each row represents a player''s fielding performance at a specific position in a game.
Enables fast fielding game log queries and game finder functionality.
Refresh after loading new plays data: REFRESH MATERIALIZED VIEW CONCURRENTLY player_game_fielding_stats;';

-- END 020_player_game_fielding_stats_view.sql

-- BEGIN 021_player_game_fielding_stats_indexes.sql
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

-- END 021_player_game_fielding_stats_indexes.sql

-- BEGIN 022_team_game_stats_view.sql
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
LEFT JOIN pitching_stats p ON b.game_id = p.game_id AND b.team_id = p.team_id;

COMMENT ON MATERIALIZED VIEW team_game_stats IS
'Per-game team statistics aggregated from play-by-play data.
Includes both offensive (batting) and defensive (pitching/fielding) stats.
Enables daily performance tracking and rolling aggregate queries.
Refresh after loading new plays data: REFRESH MATERIALIZED VIEW CONCURRENTLY team_game_stats;';

-- END 022_team_game_stats_view.sql

-- BEGIN 023_team_game_stats_indexes.sql
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

-- END 023_team_game_stats_indexes.sql

-- BEGIN 027_no_hitters_view.sql
-- Create materialized view for no-hitter achievements
-- A no-hitter is a game where a team allows zero hits to the opposing team
-- Coverage: All games in plays table (1910-2025)

DROP MATERIALIZED VIEW IF EXISTS no_hitters CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS no_hitters AS
WITH game_hits AS (
    SELECT
        p.gid as game_id,
        p.batteam as batting_team,
        p.pitteam as pitching_team,
        SUM(p.single + p.double + p.triple + p.hr) as hits_allowed
    FROM plays p
    WHERE p.gid IS NOT NULL
      AND p.batteam IS NOT NULL
      AND p.pitteam IS NOT NULL
    GROUP BY p.gid, p.batteam, p.pitteam
)
SELECT
    gh.game_id,
    gh.pitching_team as team_id,
    gh.batting_team as opponent_team_id,
    g.date,
    CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) as season,
    g.home_team,
    g.visiting_team,
    g.home_score,
    g.visiting_score,
    g.game_length_outs / 3 as innings,
    g.park_id,
    -- Determine if home or away team threw the no-hitter
    CASE
        WHEN gh.pitching_team = g.home_team THEN 'home'
        ELSE 'away'
    END as team_location,
    -- Get winning pitcher info
    g.winning_pitcher_id,
    g.winning_pitcher_name
FROM game_hits gh
JOIN games g ON gh.game_id = g.game_id
WHERE gh.hits_allowed = 0
  AND g.game_length_outs >= 27; -- At least 9 innings pitched

COMMENT ON MATERIALIZED VIEW no_hitters IS
'No-hitter achievements: games where a team allowed zero hits. Includes game metadata and pitcher information.';

-- Create indexes for no_hitters materialized view
CREATE INDEX IF NOT EXISTS idx_no_hitters_game_id ON no_hitters(game_id);
CREATE INDEX IF NOT EXISTS idx_no_hitters_team_id ON no_hitters(team_id);
CREATE INDEX IF NOT EXISTS idx_no_hitters_season ON no_hitters(season);
CREATE INDEX IF NOT EXISTS idx_no_hitters_date ON no_hitters(date);
CREATE INDEX IF NOT EXISTS idx_no_hitters_pitcher ON no_hitters(winning_pitcher_id);

-- END 027_no_hitters_view.sql

-- BEGIN 028_cycles_view.sql
-- Create materialized view for hitting for the cycle achievements
-- A cycle is when a player hits a single, double, triple, and home run in the same game
-- Coverage: All games in plays table (1910-2025)

DROP MATERIALIZED VIEW IF EXISTS cycles CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS cycles AS
WITH player_game_hits AS (
    SELECT
        p.gid as game_id,
        p.batter as player_id,
        p.batteam as team_id,
        SUM(p.single) as singles,
        SUM(p.double) as doubles,
        SUM(p.triple) as triples,
        SUM(p.hr) as home_runs,
        SUM(p.single + p.double + p.triple + p.hr) as total_hits
    FROM plays p
    WHERE p.batter IS NOT NULL
      AND p.batter != ''
      AND p.gid IS NOT NULL
    GROUP BY p.gid, p.batter, p.batteam
)
SELECT
    pgh.game_id,
    pgh.player_id,
    pgh.team_id,
    g.date,
    CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) as season,
    g.home_team,
    g.visiting_team,
    pgh.singles,
    pgh.doubles,
    pgh.triples,
    pgh.home_runs,
    pgh.total_hits,
    g.park_id,
    -- Determine if player was home or away
    CASE
        WHEN pgh.team_id = g.home_team THEN 'home'
        ELSE 'away'
    END as team_location
FROM player_game_hits pgh
JOIN games g ON pgh.game_id = g.game_id
WHERE pgh.singles >= 1
  AND pgh.doubles >= 1
  AND pgh.triples >= 1
  AND pgh.home_runs >= 1;

COMMENT ON MATERIALIZED VIEW cycles IS
'Hitting for the cycle achievements: games where a player hit a single, double, triple, and home run.';

-- Create indexes for cycles materialized view
CREATE INDEX IF NOT EXISTS idx_cycles_game_id ON cycles(game_id);
CREATE INDEX IF NOT EXISTS idx_cycles_player_id ON cycles(player_id);
CREATE INDEX IF NOT EXISTS idx_cycles_team_id ON cycles(team_id);
CREATE INDEX IF NOT EXISTS idx_cycles_season ON cycles(season);
CREATE INDEX IF NOT EXISTS idx_cycles_date ON cycles(date);

-- END 028_cycles_view.sql

-- BEGIN 029_multi_hr_games_view.sql
-- Create materialized view for multiple home run game achievements
DROP MATERIALIZED VIEW IF EXISTS multi_hr_games CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS multi_hr_games AS
WITH player_game_hrs AS (
    SELECT
        p.gid as game_id,
        p.batter as player_id,
        p.batteam as team_id,
        SUM(p.hr) as home_runs,
        SUM(p.single + p.double + p.triple + p.hr) as total_hits,
        SUM(p.ab) as at_bats
    FROM plays p
    WHERE p.batter IS NOT NULL
      AND p.batter != ''
      AND p.gid IS NOT NULL
    GROUP BY p.gid, p.batter, p.batteam
)
SELECT
    pgh.game_id,
    pgh.player_id,
    pgh.team_id,
    g.date,
    CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) as season,
    g.home_team,
    g.visiting_team,
    pgh.home_runs,
    pgh.total_hits,
    pgh.at_bats,
    g.park_id,
    CASE
        WHEN pgh.team_id = g.home_team THEN 'home'
        ELSE 'away'
    END as team_location
FROM player_game_hrs pgh
JOIN games g ON pgh.game_id = g.game_id
WHERE pgh.home_runs >= 3;

COMMENT ON MATERIALIZED VIEW multi_hr_games IS
'Multiple home run game achievements: games where a player hit 3 or more home runs.';

CREATE INDEX IF NOT EXISTS idx_multi_hr_games_game_id ON multi_hr_games(game_id);
CREATE INDEX IF NOT EXISTS idx_multi_hr_games_player_id ON multi_hr_games(player_id);
CREATE INDEX IF NOT EXISTS idx_multi_hr_games_team_id ON multi_hr_games(team_id);
CREATE INDEX IF NOT EXISTS idx_multi_hr_games_season ON multi_hr_games(season);
CREATE INDEX IF NOT EXISTS idx_multi_hr_games_date ON multi_hr_games(date);
CREATE INDEX IF NOT EXISTS idx_multi_hr_games_hr_count ON multi_hr_games(home_runs DESC);

-- END 029_multi_hr_games_view.sql

-- BEGIN 036_season_batting_leaders_view.sql
-- Create materialized view for season batting leaders
-- Combines Retrosheet per-game stats (1903-2025) with Lahman pre-1903 data (1871-1902)
-- Pre-aggregates all stats including advanced metrics (wOBA, wRC+)

CREATE MATERIALIZED VIEW IF NOT EXISTS season_batting_leaders AS
WITH retrosheet_batting AS (
    SELECT
        player_id,
        season,
        SUM(pa) as pa,
        SUM(ab) as ab,
        SUM(h) as h,
        SUM(doubles) as doubles,
        SUM(triples) as triples,
        SUM(hr) as hr,
        SUM(rbi) as rbi,
        SUM(sb) as sb,
        SUM(cs) as cs,
        SUM(bb) as bb,
        SUM(ibb) as ibb,
        SUM(so) as so,
        SUM(hbp) as hbp,
        SUM(sf) as sf,
        SUM(sh) as sh,
        SUM(gdp) as gdp,
        MAX(team_id) as team_id
    FROM player_game_batting_stats
    GROUP BY player_id, season
),
retrosheet_with_league AS (
    -- Add league information based on team_id
    SELECT
        rb.*,
        COALESCE(
            (SELECT DISTINCT home_team_league
             FROM games
             WHERE CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) = rb.season
               AND home_team = rb.team_id
             LIMIT 1),
            (SELECT DISTINCT visiting_team_league
             FROM games
             WHERE CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) = rb.season
               AND visiting_team = rb.team_id
             LIMIT 1)
        ) as league
    FROM retrosheet_batting rb
),
lahman_batting AS (
    SELECT
        "playerID" as player_id,
        "yearID" as season,
        SUM("AB" + "BB" + COALESCE("HBP", 0) + COALESCE("SF", 0)) as pa,
        SUM("AB") as ab,
        SUM("H") as h,
        SUM("2B") as doubles,
        SUM("3B") as triples,
        SUM("HR") as hr,
        SUM("RBI") as rbi,
        SUM(COALESCE("SB", 0)) as sb,
        SUM(COALESCE("CS", 0)) as cs,
        SUM("BB") as bb,
        SUM(COALESCE("IBB", 0)) as ibb,
        SUM(COALESCE("SO", 0)) as so,
        SUM(COALESCE("HBP", 0)) as hbp,
        SUM(COALESCE("SF", 0)) as sf,
        SUM(COALESCE("SH", 0)) as sh,
        SUM(COALESCE("GIDP", 0)) as gdp,
        MAX("teamID") as team_id,
        MAX("lgID") as league
    FROM "Batting"
    WHERE "yearID" < 1903
    GROUP BY "playerID", "yearID"
),
all_batting AS (
    SELECT * FROM retrosheet_with_league
    UNION ALL
    SELECT * FROM lahman_batting
),
stats_with_advanced AS (
    SELECT
        ab.*,
        CASE WHEN ab.ab > 0 THEN ROUND((ab.h::numeric / ab.ab), 3) ELSE 0 END as avg,
        CASE WHEN ab.pa > 0 THEN ROUND(((ab.h + ab.bb + ab.hbp)::numeric / ab.pa), 3) ELSE 0 END as obp,
        CASE WHEN ab.ab > 0 THEN ROUND(((ab.h + ab.doubles + 2*ab.triples + 3*ab.hr)::numeric / ab.ab), 3) ELSE 0 END as slg,
        CASE WHEN ab.ab > 0 THEN ROUND(((ab.h + ab.doubles + 2*ab.triples + 3*ab.hr)::numeric / ab.ab - ab.h::numeric / ab.ab), 3) ELSE 0 END as iso,
        CASE WHEN (ab.ab - ab.so - ab.hr + ab.sf) > 0 THEN ROUND(((ab.h - ab.hr)::numeric / (ab.ab - ab.so - ab.hr + ab.sf)), 3) ELSE 0 END as babip,
        CASE WHEN ab.pa > 0 THEN ROUND((ab.so::numeric / ab.pa), 3) ELSE 0 END as k_rate,
        CASE WHEN ab.pa > 0 THEN ROUND((ab.bb::numeric / ab.pa), 3) ELSE 0 END as bb_rate,
        CASE WHEN wc.season IS NOT NULL AND (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) > 0 THEN
            ROUND(
                (wc.w_bb * (ab.bb - ab.ibb) + wc.w_hbp * ab.hbp +
                 wc.w_1b * (ab.h - ab.doubles - ab.triples - ab.hr) +
                 wc.w_2b * ab.doubles + wc.w_3b * ab.triples + wc.w_hr * ab.hr)::numeric /
                (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp),
                3
            )
        ELSE NULL END as woba,
        CASE WHEN wc.season IS NOT NULL AND (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) > 0 THEN
            ROUND(
                (((wc.w_bb * (ab.bb - ab.ibb) + wc.w_hbp * ab.hbp +
                   wc.w_1b * (ab.h - ab.doubles - ab.triples - ab.hr) +
                   wc.w_2b * ab.doubles + wc.w_3b * ab.triples + wc.w_hr * ab.hr)::numeric /
                  (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) - wc.woba) /
                 wc.woba_scale * ab.pa)::numeric,
                2
            )
        ELSE NULL END as wraa,
        CASE WHEN wc.season IS NOT NULL AND lc.season IS NOT NULL AND (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) > 0 THEN
            ROUND(
                (((wc.w_bb * (ab.bb - ab.ibb) + wc.w_hbp * ab.hbp +
                   wc.w_1b * (ab.h - ab.doubles - ab.triples - ab.hr) +
                   wc.w_2b * ab.doubles + wc.w_3b * ab.triples + wc.w_hr * ab.hr)::numeric /
                  (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) - wc.woba) /
                 wc.woba_scale * ab.pa + (lc.wrc_per_pa * ab.pa))::numeric,
                2
            )
        ELSE NULL END as wrc,
        -- wRC+ (park adjusted)
        CASE WHEN lc.season IS NOT NULL AND lc.wrc_per_pa > 0 AND (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) > 0 THEN
            ROUND(
                ((((wc.w_bb * (ab.bb - ab.ibb) + wc.w_hbp * ab.hbp +
                    wc.w_1b * (ab.h - ab.doubles - ab.triples - ab.hr) +
                    wc.w_2b * ab.doubles + wc.w_3b * ab.triples + wc.w_hr * ab.hr)::numeric /
                   (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) - wc.woba) /
                  wc.woba_scale + lc.wrc_per_pa) +
                  (lc.wrc_per_pa - COALESCE(pf.basic_5yr, 100) / 100.0 * lc.wrc_per_pa)) /
                 lc.wrc_per_pa * 100,
                0
            )::int
        ELSE NULL END as wrc_plus
    FROM all_batting ab
    LEFT JOIN woba_constants wc ON wc.season = ab.season
    LEFT JOIN league_constants lc ON lc.season = ab.season AND lc.league = ab.league
    LEFT JOIN fangraphs_team_park_map pm ON pm.retrosheet_team_id = ab.team_id
        AND ab.season BETWEEN COALESCE(pm.start_year, 1871) AND COALESCE(pm.end_year, 2100)
    LEFT JOIN park_factors pf ON pf.park_id = pm.primary_park_id AND pf.season = ab.season
)
SELECT
    player_id,
    season,
    team_id,
    league,
    pa, ab, h, doubles, triples, hr, rbi, sb, cs, bb, ibb, so, hbp, sf, sh, gdp,
    avg, obp, slg, iso, babip, k_rate, bb_rate,
    woba, wraa, wrc, wrc_plus,
    -- Calculated OPS
    CASE WHEN obp IS NOT NULL AND slg IS NOT NULL THEN ROUND((obp + slg)::numeric, 3) ELSE NULL END as ops
FROM stats_with_advanced;

CREATE UNIQUE INDEX IF NOT EXISTS idx_season_batting_leaders_pk ON season_batting_leaders(player_id, season);

CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_hr ON season_batting_leaders(season, hr DESC, h DESC) WHERE ab >= 300;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_avg ON season_batting_leaders(season, avg DESC) WHERE ab >= 300;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_rbi ON season_batting_leaders(season, rbi DESC) WHERE ab >= 300;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_sb ON season_batting_leaders(season, sb DESC) WHERE ab >= 300;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_h ON season_batting_leaders(season, h DESC) WHERE ab >= 300;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_wrc_plus ON season_batting_leaders(season, wrc_plus DESC) WHERE pa >= 502;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_woba ON season_batting_leaders(season, woba DESC) WHERE pa >= 502;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_league_hr ON season_batting_leaders(season, league, hr DESC) WHERE ab >= 300;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_season ON season_batting_leaders(season);
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_player ON season_batting_leaders(player_id, season DESC);

ANALYZE season_batting_leaders;

-- END 036_season_batting_leaders_view.sql

-- BEGIN 037_season_pitching_leaders_view.sql
-- Create materialized view for season pitching leaders
-- Combines Retrosheet per-game stats (1903-2025) with W/L/SV from games table
-- Includes Lahman pre-1903 data (1871-1902)
-- Pre-aggregates all stats including advanced metrics (FIP, WHIP, K/9)

CREATE MATERIALIZED VIEW IF NOT EXISTS season_pitching_leaders AS
WITH retrosheet_pitching AS (
    SELECT
        player_id,
        season,
        COUNT(*) as g,  -- games appeared
        SUM(ip * 3) as ipouts,  -- convert IP back to outs
        SUM(h) as h,
        SUM(er) as er,
        SUM(hr) as hr,
        SUM(bb) as bb,
        SUM(so) as so,
        SUM(ibb) as ibb,
        SUM(hbp) as hbp,
        SUM(wp) as wp,
        SUM(bk) as bk,
        SUM(pa) as bfp,
        MAX(team_id) as team_id
    FROM player_game_pitching_stats
    GROUP BY player_id, season
),
retrosheet_with_league AS (
    SELECT
        rp.*,
        (
            SELECT DISTINCT home_team_league
            FROM games
            WHERE CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) = rp.season
              AND (home_team = rp.team_id OR visiting_team = rp.team_id)
            LIMIT 1
        ) as league
    FROM retrosheet_pitching rp
),
pitcher_decisions AS (
    SELECT
        season,
        player_id,
        COUNT(*) FILTER (WHERE decision = 'W') as w,
        COUNT(*) FILTER (WHERE decision = 'L') as l,
        COUNT(*) FILTER (WHERE decision = 'SV') as sv,
        COUNT(*) FILTER (WHERE decision = 'GS') as gs
    FROM (
        SELECT
            CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) as season,
            winning_pitcher_id as player_id,
            'W' as decision
        FROM games
        WHERE winning_pitcher_id IS NOT NULL AND winning_pitcher_id != ''

        UNION ALL

        SELECT
            CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) as season,
            losing_pitcher_id as player_id,
            'L' as decision
        FROM games
        WHERE losing_pitcher_id IS NOT NULL AND losing_pitcher_id != ''

        UNION ALL

        SELECT
            CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) as season,
            saving_pitcher_id as player_id,
            'SV' as decision
        FROM games
        WHERE saving_pitcher_id IS NOT NULL AND saving_pitcher_id != ''

        UNION ALL

        SELECT
            CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) as season,
            h_starting_pitcher_id as player_id,
            'GS' as decision
        FROM games
        WHERE h_starting_pitcher_id IS NOT NULL AND h_starting_pitcher_id != ''

        UNION ALL

        SELECT
            CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) as season,
            v_starting_pitcher_id as player_id,
            'GS' as decision
        FROM games
        WHERE v_starting_pitcher_id IS NOT NULL AND v_starting_pitcher_id != ''
    ) decisions
    GROUP BY season, player_id
),
retrosheet_combined AS (
    SELECT
        rp.player_id,
        rp.season,
        COALESCE(pd.w, 0) as w,
        COALESCE(pd.l, 0) as l,
        COALESCE(pd.sv, 0) as sv,
        COALESCE(pd.gs, 0) as gs,
        0 as cg,
        0 as sho,
        rp.g,
        rp.ipouts,
        rp.h,
        rp.er,
        rp.hr,
        rp.bb,
        rp.so,
        rp.ibb,
        rp.hbp,
        rp.wp,
        rp.bk,
        rp.bfp,
        rp.team_id,
        rp.league
    FROM retrosheet_with_league rp
    LEFT JOIN pitcher_decisions pd ON pd.player_id = rp.player_id AND pd.season = rp.season
),
lahman_pitching AS (
    SELECT
        "playerID" as player_id,
        "yearID" as season,
        SUM("W") as w,
        SUM("L") as l,
        SUM("SV") as sv,
        SUM("GS") as gs,
        SUM("CG") as cg,
        SUM("SHO") as sho,
        SUM("G") as g,
        SUM("IPouts") as ipouts,
        SUM("H") as h,
        SUM("ER") as er,
        SUM("HR") as hr,
        SUM("BB") as bb,
        SUM("SO") as so,
        SUM(COALESCE("IBB", 0)) as ibb,
        SUM(COALESCE("HBP", 0)) as hbp,
        SUM(COALESCE("WP", 0)) as wp,
        SUM(COALESCE("BK", 0)) as bk,
        SUM(COALESCE("BFP", 0)) as bfp,
        MAX("teamID") as team_id,
        MAX("lgID") as league
    FROM "Pitching"
    WHERE "yearID" < 1903
    GROUP BY "playerID", "yearID"
),
all_pitching AS (
    SELECT * FROM retrosheet_combined
    UNION ALL
    SELECT * FROM lahman_pitching
)
SELECT
    ap.*,
    ROUND((ap.ipouts::numeric / 3), 1) as ip,
    CASE WHEN ap.ipouts > 0 THEN ROUND((ap.er::numeric * 27.0 / ap.ipouts), 2) ELSE 0 END as era,
    CASE WHEN ap.ipouts > 0 THEN ROUND(((ap.h + ap.bb)::numeric * 3 / ap.ipouts), 2) ELSE 0 END as whip,
    CASE WHEN ap.ipouts > 0 THEN ROUND((ap.so::numeric * 27.0 / ap.ipouts), 2) ELSE 0 END as k_per_9,
    CASE WHEN ap.ipouts > 0 THEN ROUND((ap.bb::numeric * 27.0 / ap.ipouts), 2) ELSE 0 END as bb_per_9,
    CASE WHEN ap.ipouts > 0 THEN ROUND((ap.hr::numeric * 27.0 / ap.ipouts), 2) ELSE 0 END as hr_per_9,
    CASE WHEN ap.ipouts > 0 THEN
        ROUND(
            (((13 * ap.hr + 3 * (ap.bb + ap.hbp) - 2 * ap.so)::numeric / (ap.ipouts / 3.0)) +
             COALESCE(wc.c_fip, 3.2))::numeric,
            2
        )
    ELSE NULL END as fip
FROM all_pitching ap
LEFT JOIN woba_constants wc ON wc.season = ap.season;

CREATE UNIQUE INDEX IF NOT EXISTS idx_season_pitching_leaders_pk ON season_pitching_leaders(player_id, season);

CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_era ON season_pitching_leaders(season, era ASC) WHERE ipouts >= 450;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_so ON season_pitching_leaders(season, so DESC) WHERE ipouts >= 450;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_w ON season_pitching_leaders(season, w DESC) WHERE ipouts >= 450;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_sv ON season_pitching_leaders(season, sv DESC) WHERE g >= 20;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_fip ON season_pitching_leaders(season, fip ASC) WHERE ipouts >= 450;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_whip ON season_pitching_leaders(season, whip ASC) WHERE ipouts >= 450;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_league_era ON season_pitching_leaders(season, league, era ASC) WHERE ipouts >= 450;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_season ON season_pitching_leaders(season);
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_player ON season_pitching_leaders(player_id, season DESC);

ANALYZE season_pitching_leaders;

-- END 037_season_pitching_leaders_view.sql

-- BEGIN 038_career_batting_leaders_view.sql
-- Create materialized view for career batting leaders
-- Aggregates from season_batting_leaders to get career totals
-- Pre-calculates career rate stats

CREATE MATERIALIZED VIEW IF NOT EXISTS career_batting_leaders AS
SELECT
    player_id,
    MAX(season) as last_season,
    SUM(pa) as total_pa,
    SUM(ab) as total_ab,
    SUM(h) as total_h,
    SUM(doubles) as total_doubles,
    SUM(triples) as total_triples,
    SUM(hr) as total_hr,
    SUM(rbi) as total_rbi,
    SUM(sb) as total_sb,
    SUM(cs) as total_cs,
    SUM(bb) as total_bb,
    SUM(ibb) as total_ibb,
    SUM(so) as total_so,
    SUM(hbp) as total_hbp,
    SUM(sf) as total_sf,
    SUM(sh) as total_sh,
    SUM(gdp) as total_gdp,
    CASE WHEN SUM(ab) > 0 THEN ROUND((SUM(h)::numeric / SUM(ab)), 3) ELSE 0 END as career_avg,
    CASE WHEN SUM(pa) > 0 THEN ROUND(((SUM(h) + SUM(bb) + SUM(hbp))::numeric / SUM(pa)), 3) ELSE 0 END as career_obp,
    CASE WHEN SUM(ab) > 0 THEN ROUND(((SUM(h) + SUM(doubles) + 2*SUM(triples) + 3*SUM(hr))::numeric / SUM(ab)), 3) ELSE 0 END as career_slg,
    -- Career OPS
    CASE WHEN SUM(pa) > 0 AND SUM(ab) > 0 THEN
        ROUND((
            ((SUM(h) + SUM(bb) + SUM(hbp))::numeric / SUM(pa)) +
            ((SUM(h) + SUM(doubles) + 2*SUM(triples) + 3*SUM(hr))::numeric / SUM(ab))
        ), 3)
    ELSE 0 END as career_ops,
    COUNT(*) as seasons_played
FROM season_batting_leaders
GROUP BY player_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_career_batting_leaders_pk ON career_batting_leaders(player_id);

CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_hr ON career_batting_leaders(total_hr DESC, total_h DESC) WHERE total_ab >= 1000;
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_h ON career_batting_leaders(total_h DESC) WHERE total_ab >= 1000;
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_rbi ON career_batting_leaders(total_rbi DESC) WHERE total_ab >= 1000;
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_avg ON career_batting_leaders(career_avg DESC) WHERE total_ab >= 1000;
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_ops ON career_batting_leaders(career_ops DESC) WHERE total_pa >= 3000;

ANALYZE career_batting_leaders;

-- END 038_career_batting_leaders_view.sql

-- BEGIN 039_career_pitching_leaders_view.sql
-- Create materialized view for career pitching leaders
-- Aggregates from season_pitching_leaders to get career totals
-- Pre-calculates career rate stats

CREATE MATERIALIZED VIEW IF NOT EXISTS career_pitching_leaders AS
SELECT
    player_id,
    MAX(season) as last_season,
    SUM(w) as total_w,
    SUM(l) as total_l,
    SUM(sv) as total_sv,
    SUM(gs) as total_gs,
    SUM(cg) as total_cg,
    SUM(sho) as total_sho,
    SUM(g) as total_g,
    SUM(ipouts) as total_ipouts,
    SUM(h) as total_h,
    SUM(er) as total_er,
    SUM(hr) as total_hr,
    SUM(bb) as total_bb,
    SUM(so) as total_so,
    SUM(ibb) as total_ibb,
    SUM(hbp) as total_hbp,
    SUM(wp) as total_wp,
    SUM(bk) as total_bk,
    SUM(bfp) as total_bfp,
    ROUND((SUM(ipouts)::numeric / 3), 1) as career_ip,
    CASE WHEN SUM(ipouts) > 0 THEN ROUND((SUM(er)::numeric * 27.0 / SUM(ipouts)), 2) ELSE 0 END as career_era,
    CASE WHEN SUM(ipouts) > 0 THEN ROUND(((SUM(h) + SUM(bb))::numeric * 3 / SUM(ipouts)), 2) ELSE 0 END as career_whip,
    CASE WHEN SUM(ipouts) > 0 THEN ROUND((SUM(so)::numeric * 27.0 / SUM(ipouts)), 2) ELSE 0 END as career_k_per_9,
    CASE WHEN SUM(ipouts) > 0 THEN ROUND((SUM(bb)::numeric * 27.0 / SUM(ipouts)), 2) ELSE 0 END as career_bb_per_9,
    CASE WHEN SUM(ipouts) > 0 THEN ROUND((SUM(hr)::numeric * 27.0 / SUM(ipouts)), 2) ELSE 0 END as career_hr_per_9,
    -- TODO: use a computed constant
    CASE WHEN SUM(ipouts) > 0 THEN
        ROUND(
            (((13 * SUM(hr) + 3 * (SUM(bb) + SUM(hbp)) - 2 * SUM(so))::numeric / (SUM(ipouts) / 3.0)) + 3.2)::numeric,
            2
        )
    ELSE NULL END as career_fip,
    COUNT(*) as seasons_pitched
FROM season_pitching_leaders
GROUP BY player_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_career_pitching_leaders_pk ON career_pitching_leaders(player_id);

CREATE INDEX IF NOT EXISTS idx_career_pitching_leaders_w ON career_pitching_leaders(total_w DESC) WHERE total_ipouts >= 1500;
CREATE INDEX IF NOT EXISTS idx_career_pitching_leaders_so ON career_pitching_leaders(total_so DESC) WHERE total_ipouts >= 1500;
CREATE INDEX IF NOT EXISTS idx_career_pitching_leaders_sv ON career_pitching_leaders(total_sv DESC) WHERE total_g >= 100;
CREATE INDEX IF NOT EXISTS idx_career_pitching_leaders_era ON career_pitching_leaders(career_era ASC) WHERE total_ipouts >= 1500;
CREATE INDEX IF NOT EXISTS idx_career_pitching_leaders_whip ON career_pitching_leaders(career_whip ASC) WHERE total_ipouts >= 1500;

ANALYZE career_pitching_leaders;

-- END 039_career_pitching_leaders_view.sql

