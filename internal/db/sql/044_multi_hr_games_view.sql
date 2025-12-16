-- Create materialized view for multiple home run game achievements
-- Tracks games where a player hit 3 or more home runs
-- Coverage: All games in plays table (1910-2025)

DROP MATERIALIZED VIEW IF EXISTS multi_hr_games CASCADE;

CREATE MATERIALIZED VIEW multi_hr_games AS
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
    -- Determine if player was home or away
    CASE
        WHEN pgh.team_id = g.home_team THEN 'home'
        ELSE 'away'
    END as team_location
FROM player_game_hrs pgh
JOIN games g ON pgh.game_id = g.game_id
WHERE pgh.home_runs >= 3;

COMMENT ON MATERIALIZED VIEW multi_hr_games IS
'Multiple home run game achievements: games where a player hit 3 or more home runs.';

-- Create indexes for multi_hr_games materialized view
CREATE INDEX idx_multi_hr_games_game_id ON multi_hr_games(game_id);
CREATE INDEX idx_multi_hr_games_player_id ON multi_hr_games(player_id);
CREATE INDEX idx_multi_hr_games_team_id ON multi_hr_games(team_id);
CREATE INDEX idx_multi_hr_games_season ON multi_hr_games(season);
CREATE INDEX idx_multi_hr_games_date ON multi_hr_games(date);
CREATE INDEX idx_multi_hr_games_hr_count ON multi_hr_games(home_runs DESC);
