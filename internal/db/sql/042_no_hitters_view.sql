-- Create materialized view for no-hitter achievements
-- A no-hitter is a game where a team allows zero hits to the opposing team
-- Coverage: All games in plays table (1910-2025)

DROP MATERIALIZED VIEW IF EXISTS no_hitters CASCADE;

CREATE MATERIALIZED VIEW no_hitters AS
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
CREATE INDEX idx_no_hitters_game_id ON no_hitters(game_id);
CREATE INDEX idx_no_hitters_team_id ON no_hitters(team_id);
CREATE INDEX idx_no_hitters_season ON no_hitters(season);
CREATE INDEX idx_no_hitters_date ON no_hitters(date);
CREATE INDEX idx_no_hitters_pitcher ON no_hitters(winning_pitcher_id);
