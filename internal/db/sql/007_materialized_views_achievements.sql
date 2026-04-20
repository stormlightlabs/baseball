-- Fresh migration set: achievements and notable-game materialized views.
-- Views are created WITH NO DATA and refreshed via ETL/db refresh-views.


-- SECTION 027_no_hitters_view.sql

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
  AND g.game_length_outs >= 27 -- At least 9 innings pitched
WITH NO DATA;

COMMENT ON MATERIALIZED VIEW no_hitters IS
'No-hitter achievements: games where a team allowed zero hits. Includes game metadata and pitcher information.';

-- Create indexes for no_hitters materialized view
CREATE INDEX IF NOT EXISTS idx_no_hitters_game_id ON no_hitters(game_id);
CREATE INDEX IF NOT EXISTS idx_no_hitters_team_id ON no_hitters(team_id);
CREATE INDEX IF NOT EXISTS idx_no_hitters_season ON no_hitters(season);
CREATE INDEX IF NOT EXISTS idx_no_hitters_date ON no_hitters(date);
CREATE INDEX IF NOT EXISTS idx_no_hitters_pitcher ON no_hitters(winning_pitcher_id);



-- SECTION 028_cycles_view.sql

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
  AND pgh.home_runs >= 1
WITH NO DATA;

COMMENT ON MATERIALIZED VIEW cycles IS
'Hitting for the cycle achievements: games where a player hit a single, double, triple, and home run.';

-- Create indexes for cycles materialized view
CREATE INDEX IF NOT EXISTS idx_cycles_game_id ON cycles(game_id);
CREATE INDEX IF NOT EXISTS idx_cycles_player_id ON cycles(player_id);
CREATE INDEX IF NOT EXISTS idx_cycles_team_id ON cycles(team_id);
CREATE INDEX IF NOT EXISTS idx_cycles_season ON cycles(season);
CREATE INDEX IF NOT EXISTS idx_cycles_date ON cycles(date);



-- SECTION 029_multi_hr_games_view.sql

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
WHERE pgh.home_runs >= 3
WITH NO DATA;

COMMENT ON MATERIALIZED VIEW multi_hr_games IS
'Multiple home run game achievements: games where a player hit 3 or more home runs.';

CREATE INDEX IF NOT EXISTS idx_multi_hr_games_game_id ON multi_hr_games(game_id);
CREATE INDEX IF NOT EXISTS idx_multi_hr_games_player_id ON multi_hr_games(player_id);
CREATE INDEX IF NOT EXISTS idx_multi_hr_games_team_id ON multi_hr_games(team_id);
CREATE INDEX IF NOT EXISTS idx_multi_hr_games_season ON multi_hr_games(season);
CREATE INDEX IF NOT EXISTS idx_multi_hr_games_date ON multi_hr_games(date);
CREATE INDEX IF NOT EXISTS idx_multi_hr_games_hr_count ON multi_hr_games(home_runs DESC);



-- SECTION 030_triple_plays_view.sql


-- Create materialized view for triple play achievements
DROP MATERIALIZED VIEW IF EXISTS triple_plays CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS triple_plays AS
WITH home_triple_plays AS (
    SELECT
        g.game_id,
        g.home_team as team_id,
        g.visiting_team as opponent_team_id,
        g.date,
        CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) as season,
        g.home_team,
        g.visiting_team,
        g.home_score as team_score,
        g.visiting_score as opponent_score,
        g.home_triple_plays as triple_plays_count,
        'home' as team_location,
        g.park_id
    FROM games g
    WHERE g.home_triple_plays > 0
),
away_triple_plays AS (
    SELECT
        g.game_id,
        g.visiting_team as team_id,
        g.home_team as opponent_team_id,
        g.date,
        CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) as season,
        g.home_team,
        g.visiting_team,
        g.visiting_score as team_score,
        g.home_score as opponent_score,
        g.visiting_triple_plays as triple_plays_count,
        'away' as team_location,
        g.park_id
    FROM games g
    WHERE g.visiting_triple_plays > 0
)
SELECT * FROM home_triple_plays
UNION ALL
SELECT * FROM away_triple_plays
ORDER BY date DESC;

COMMENT ON MATERIALIZED VIEW triple_plays IS
'Triple play achievements: games where a team recorded one or more triple plays.';

CREATE INDEX IF NOT EXISTS idx_triple_plays_game_id ON triple_plays(game_id);
CREATE INDEX IF NOT EXISTS idx_triple_plays_team_id ON triple_plays(team_id);
CREATE INDEX IF NOT EXISTS idx_triple_plays_season ON triple_plays(season);
CREATE INDEX IF NOT EXISTS idx_triple_plays_date ON triple_plays(date);




-- SECTION 031_extra_inning_games_view.sql


-- Create materialized view for extra inning game achievements
-- Tracks games that went 20 or more innings (60+ outs)

DROP MATERIALIZED VIEW IF EXISTS extra_inning_games CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS extra_inning_games AS
SELECT
    g.game_id,
    g.date,
    CAST(SUBSTRING(g.date, 1, 4) AS INTEGER) as season,
    g.home_team,
    g.visiting_team,
    g.home_team_league,
    g.visiting_team_league,
    g.home_score,
    g.visiting_score,
    g.game_length_outs / 3 as innings,
    g.game_length_outs,
    g.game_time_minutes,
    g.park_id,
    CASE
        WHEN g.home_score > g.visiting_score THEN g.home_team
        WHEN g.visiting_score > g.home_score THEN g.visiting_team
        ELSE NULL
    END as winning_team,
    CASE
        WHEN g.home_score = g.visiting_score THEN 'tie'
        WHEN g.home_score > g.visiting_score THEN 'home_win'
        ELSE 'away_win'
    END as result_type
FROM games g
WHERE g.game_length_outs >= 60
ORDER BY g.game_length_outs DESC, g.date DESC;

COMMENT ON MATERIALIZED VIEW extra_inning_games IS
'Extra inning game achievements: games that lasted 20 or more innings.';

CREATE INDEX IF NOT EXISTS idx_extra_inning_games_game_id ON extra_inning_games(game_id);
CREATE INDEX IF NOT EXISTS idx_extra_inning_games_season ON extra_inning_games(season);
CREATE INDEX IF NOT EXISTS idx_extra_inning_games_date ON extra_inning_games(date);
CREATE INDEX IF NOT EXISTS idx_extra_inning_games_innings ON extra_inning_games(innings DESC);
CREATE INDEX IF NOT EXISTS idx_extra_inning_games_home_team ON extra_inning_games(home_team);
CREATE INDEX IF NOT EXISTS idx_extra_inning_games_visiting_team ON extra_inning_games(visiting_team);


