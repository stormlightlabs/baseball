-- Create materialized view for extra inning game achievements
-- Tracks games that went 20 or more innings (60+ outs)

DROP MATERIALIZED VIEW IF EXISTS extra_inning_games CASCADE;

CREATE MATERIALIZED VIEW extra_inning_games AS
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

CREATE INDEX idx_extra_inning_games_game_id ON extra_inning_games(game_id);
CREATE INDEX idx_extra_inning_games_season ON extra_inning_games(season);
CREATE INDEX idx_extra_inning_games_date ON extra_inning_games(date);
CREATE INDEX idx_extra_inning_games_innings ON extra_inning_games(innings DESC);
CREATE INDEX idx_extra_inning_games_home_team ON extra_inning_games(home_team);
CREATE INDEX idx_extra_inning_games_visiting_team ON extra_inning_games(visiting_team);
