-- Batch get win expectancy for multiple game states
-- Uses UNNEST to efficiently query multiple states in a single database round-trip
--
-- Parameters:
--   $1: array of innings (int[])
--   $2: array of is_bottom (boolean[])
--   $3: array of outs (int[])
--   $4: array of runners_state (text[])
--   $5: array of score_diff (int[])

WITH input_states AS (
    SELECT
        ROW_NUMBER() OVER () as row_num,
        unnest($1::int[]) as inning,
        unnest($2::boolean[]) as is_bottom,
        unnest($3::int[]) as outs,
        unnest($4::text[]) as runners_state,
        unnest($5::int[]) as score_diff
),
ranked_results AS (
    SELECT
        i.row_num,
        w.id,
        w.inning,
        w.is_bottom,
        w.outs,
        w.runners_state,
        w.score_diff,
        w.win_probability,
        w.sample_size,
        w.start_year,
        w.end_year,
        w.created_at,
        w.updated_at,
        ROW_NUMBER() OVER (
            PARTITION BY i.row_num
            ORDER BY w.end_year DESC NULLS LAST
        ) as rn
    FROM input_states i
    LEFT JOIN win_expectancy_historical w
        ON w.inning = i.inning
        AND w.is_bottom = i.is_bottom
        AND w.outs = i.outs
        AND w.runners_state = i.runners_state
        AND w.score_diff = i.score_diff
)
SELECT
    row_num,
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
FROM ranked_results
WHERE rn = 1
ORDER BY row_num
