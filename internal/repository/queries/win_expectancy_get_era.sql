-- Get win expectancy for a specific game state within a historical era
--
-- Parameters:
--   $1: inning (1-9)
--   $2: is_bottom (boolean)
--   $3: outs (0-2)
--   $4: runners_state (e.g., "___", "1__", "123")
--   $5: score_diff (-11 to +11)
--   $6: start_year (optional, can be NULL)
--   $7: end_year (optional, can be NULL)

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
  AND ($6::int IS NULL OR start_year >= $6)
  AND ($7::int IS NULL OR end_year <= $7)
ORDER BY end_year DESC NULLS LAST
LIMIT 1
