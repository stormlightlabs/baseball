-- Create/rebuild the win_expectancy_historical materialized view.
-- Includes synthetic id and timestamps for API contract compatibility.

DROP TABLE IF EXISTS win_expectancy_historical CASCADE;
DROP MATERIALIZED VIEW IF EXISTS win_expectancy_historical CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS win_expectancy_historical AS
WITH game_outcomes AS (
    SELECT
        date,
        home_team,
        game_number,
        CASE
            WHEN home_score > visiting_score THEN true
            WHEN home_score < visiting_score THEN false
            ELSE NULL
        END as home_won
    FROM games
    WHERE home_score IS NOT NULL
      AND visiting_score IS NOT NULL
      AND home_score != visiting_score
),
game_states AS (
    SELECT
        LEAST(p.inning, 9) as inning,
        (p.top_bot = 1)::boolean as is_bottom,
        p.outs_pre as outs,
        CONCAT(
            CASE WHEN p.br1_pre IS NOT NULL AND p.br1_pre != '' THEN '1' ELSE '_' END,
            CASE WHEN p.br2_pre IS NOT NULL AND p.br2_pre != '' THEN '2' ELSE '_' END,
            CASE WHEN p.br3_pre IS NOT NULL AND p.br3_pre != '' THEN '3' ELSE '_' END
        ) as runners_state,
        LEAST(GREATEST(p.score_h - p.score_v, -11), 11) as score_diff,
        go.home_won,
        SUBSTRING(p.date, 1, 4)::int as year
    FROM plays p
    INNER JOIN game_outcomes go ON
        SUBSTRING(p.gid, 4, 8) = go.date AND
        LEFT(p.gid, 3) = go.home_team AND
        RIGHT(p.gid, 1)::int = go.game_number
    WHERE p.outs_pre IS NOT NULL
      AND p.inning IS NOT NULL
      AND go.home_won IS NOT NULL
),
win_rates AS (
    SELECT
        inning,
        is_bottom,
        outs,
        runners_state,
        score_diff,
        MIN(year) as start_year,
        MAX(year) as end_year,
        AVG(CASE WHEN home_won THEN 1.0 ELSE 0.0 END) as win_probability,
        COUNT(*) as sample_size
    FROM game_states
    GROUP BY inning, is_bottom, outs, runners_state, score_diff
    HAVING COUNT(*) >= 100
)
SELECT
    ROW_NUMBER() OVER (ORDER BY inning, is_bottom, outs, runners_state, score_diff, start_year, end_year)::int as id,
    inning,
    is_bottom,
    outs,
    runners_state,
    score_diff,
    win_probability,
    sample_size,
    start_year,
    end_year,
    NOW() as created_at,
    NOW() as updated_at
FROM win_rates;

CREATE UNIQUE INDEX IF NOT EXISTS idx_win_expectancy_state ON win_expectancy_historical(
    inning, is_bottom, outs, runners_state, score_diff, start_year, end_year
);
CREATE INDEX IF NOT EXISTS idx_win_expectancy_inning ON win_expectancy_historical(inning);
CREATE INDEX IF NOT EXISTS idx_win_expectancy_outs ON win_expectancy_historical(outs);
CREATE INDEX IF NOT EXISTS idx_win_expectancy_runners ON win_expectancy_historical(runners_state);
CREATE INDEX IF NOT EXISTS idx_win_expectancy_score ON win_expectancy_historical(score_diff);
