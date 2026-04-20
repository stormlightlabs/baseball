-- win expectancy helper function and comments.


-- SECTION 047_win_expectancy_metadata.sql


-- Keep function and object comments separate from view-shape migrations for clarity.

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


