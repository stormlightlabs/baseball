-- Create materialized view for hitting for the cycle achievements
-- A cycle is when a player hits a single, double, triple, and home run in the same game
-- Coverage: All games in plays table (1910-2025)

DROP MATERIALIZED VIEW IF EXISTS cycles CASCADE;

CREATE MATERIALIZED VIEW cycles AS
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
CREATE INDEX idx_cycles_game_id ON cycles(game_id);
CREATE INDEX idx_cycles_player_id ON cycles(player_id);
CREATE INDEX idx_cycles_team_id ON cycles(team_id);
CREATE INDEX idx_cycles_season ON cycles(season);
CREATE INDEX idx_cycles_date ON cycles(date);
