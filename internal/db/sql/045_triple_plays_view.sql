-- Create materialized view for triple play achievements
DROP MATERIALIZED VIEW IF EXISTS triple_plays CASCADE;

CREATE MATERIALIZED VIEW triple_plays AS
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

CREATE INDEX idx_triple_plays_game_id ON triple_plays(game_id);
CREATE INDEX idx_triple_plays_team_id ON triple_plays(team_id);
CREATE INDEX idx_triple_plays_season ON triple_plays(season);
CREATE INDEX idx_triple_plays_date ON triple_plays(date);
