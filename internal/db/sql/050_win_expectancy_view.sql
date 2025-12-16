-- Create materialized view for win expectancy from historical play-by-play data

DROP TABLE IF EXISTS win_expectancy_historical CASCADE;

DROP MATERIALIZED VIEW IF EXISTS win_expectancy_historical CASCADE;

CREATE MATERIALIZED VIEW win_expectancy_historical AS
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
    inning,
    is_bottom,
    outs,
    runners_state,
    score_diff,
    win_probability,
    sample_size,
    start_year,
    end_year
FROM win_rates
ORDER BY inning, is_bottom, outs, runners_state, score_diff;

CREATE UNIQUE INDEX idx_win_expectancy_state ON win_expectancy_historical(
    inning, is_bottom, outs, runners_state, score_diff, start_year, end_year
);
CREATE INDEX idx_win_expectancy_inning ON win_expectancy_historical(inning);
CREATE INDEX idx_win_expectancy_outs ON win_expectancy_historical(outs);
CREATE INDEX idx_win_expectancy_runners ON win_expectancy_historical(runners_state);
CREATE INDEX idx_win_expectancy_score ON win_expectancy_historical(score_diff);

CREATE OR REPLACE FUNCTION get_win_expectancy(
    p_inning INT,
    p_is_bottom BOOLEAN,
    p_outs INT,
    p_runners_state VARCHAR,
    p_score_diff INT
)
RETURNS TABLE(
    win_probability NUMERIC,
    sample_size BIGINT
) AS $$
    SELECT win_probability, sample_size
    FROM win_expectancy_historical
    WHERE inning = LEAST(p_inning, 9)
      AND is_bottom = p_is_bottom
      AND outs = p_outs
      AND runners_state = p_runners_state
      AND score_diff = LEAST(GREATEST(p_score_diff, -11), 11)
    ORDER BY end_year DESC
    LIMIT 1;
$$ LANGUAGE SQL STABLE;

COMMENT ON MATERIALIZED VIEW win_expectancy_historical IS
'Win expectancy probabilities derived from historical play-by-play data.
Each row represents the probability of the home team winning given a specific game state.
Minimum sample size: 100 plays per state for statistical reliability.
Coverage: All games with play-by-play data (1910-2025).
Refresh after loading new play-by-play data: REFRESH MATERIALIZED VIEW CONCURRENTLY win_expectancy_historical;';

COMMENT ON FUNCTION get_win_expectancy IS
'Get win expectancy for a specific game state (inning, is_bottom, outs, runners, score_diff).
Returns the most recent win probability and sample size for the given state.';
