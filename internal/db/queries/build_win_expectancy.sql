-- Build Win Expectancy from Historical Play-by-Play Data
--
-- This query analyzes all plays in the plays table, joins with game outcomes,
-- and calculates win probabilities for each unique game state (inning, outs,
-- runners, score differential).
--
-- The query:
-- 1. Determines game outcomes (who won)
-- 2. Extracts game state for each play
-- 3. Calculates win probability by averaging outcomes
-- 4. Filters by minimum sample size for statistical reliability
-- 5. Upserts into win_expectancy_historical table
--
-- Parameters:
--   $1: minSampleSize - minimum number of plays required for a state to be included

WITH game_outcomes AS (
    -- Determine who won each game
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
      AND home_score != visiting_score  -- Exclude ties
),
game_states AS (
    -- Extract game state for each play
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
    -- Calculate win probability for each unique game state
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
    HAVING COUNT(*) >= $1
)
INSERT INTO win_expectancy_historical (
    inning, is_bottom, outs, runners_state, score_diff,
    win_probability, sample_size, start_year, end_year,
    created_at, updated_at
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
    end_year,
    NOW(),
    NOW()
FROM win_rates
ON CONFLICT (inning, is_bottom, outs, runners_state, score_diff, start_year, end_year)
DO UPDATE SET
    win_probability = EXCLUDED.win_probability,
    sample_size = EXCLUDED.sample_size,
    updated_at = NOW()
