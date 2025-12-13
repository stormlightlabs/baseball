-- Get win expectancy for a specific game state
-- Uses the most recent era available if no era is specified
--
-- Parameters:
--   $1: inning (1-9)
--   $2: is_bottom (boolean)
--   $3: outs (0-2)
--   $4: runners_state (e.g., "___", "1__", "123")
--   $5: score_diff (-11 to +11)

SELECT
    id,
    inning,
    is_bottom,
    outs,
    runners_state,
    score_diff,
    win_probability,
    sample_size,
    start_year,
    end_year,
    created_at,
    updated_at
FROM win_expectancy_historical
WHERE inning = $1
  AND is_bottom = $2
  AND outs = $3
  AND runners_state = $4
  AND score_diff = $5
ORDER BY end_year DESC NULLS LAST
LIMIT 1
