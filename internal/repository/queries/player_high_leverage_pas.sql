WITH player_pas AS (
    SELECT
        p.gid as game_id,
        p.pn as event_id,
        p.inning,
        p.top_bot,
        p.batter as batter_id,
        p.pitcher as pitcher_id,
        p.score_h as score_home,
        p.score_v as score_vis,
        p.outs_pre,
        p.outs_post,
        p.br1_post as runner1_post,
        p.br2_post as runner2_post,
        p.br3_post as runner3_post,
        -- Convert bases to underscore format for WE table lookup (before)
        COALESCE(
            CASE WHEN p.br1_pre IS NOT NULL AND p.br1_pre != '' THEN '1' ELSE '_' END ||
            CASE WHEN p.br2_pre IS NOT NULL AND p.br2_pre != '' THEN '2' ELSE '_' END ||
            CASE WHEN p.br3_pre IS NOT NULL AND p.br3_pre != '' THEN '3' ELSE '_' END,
            '___'
        ) as runners_state_before,
        -- Convert bases to underscore format for WE table lookup (after)
        COALESCE(
            CASE WHEN p.br1_post IS NOT NULL AND p.br1_post != '' THEN '1' ELSE '_' END ||
            CASE WHEN p.br2_post IS NOT NULL AND p.br2_post != '' THEN '2' ELSE '_' END ||
            CASE WHEN p.br3_post IS NOT NULL AND p.br3_post != '' THEN '3' ELSE '_' END,
            '___'
        ) as runners_state_after,
        -- Display format with 0s
        COALESCE(
            CASE WHEN p.br1_pre IS NOT NULL AND p.br1_pre != '' THEN '1' ELSE '0' END ||
            CASE WHEN p.br2_pre IS NOT NULL AND p.br2_pre != '' THEN '1' ELSE '0' END ||
            CASE WHEN p.br3_pre IS NOT NULL AND p.br3_pre != '' THEN '1' ELSE '0' END,
            '000'
        ) as bases_before,
        p.event as description,
        -- Score diff from batting team perspective (before)
        CASE
            WHEN p.top_bot = 0 THEN LEAST(GREATEST(p.score_v - p.score_h, -11), 11)
            ELSE LEAST(GREATEST(p.score_h - p.score_v, -11), 11)
        END as score_diff_before,
        -- Score diff from batting team perspective (after) - account for runs scored
        CASE
            WHEN p.top_bot = 0 THEN LEAST(GREATEST((p.score_v + COALESCE(p.runs, 0)) - p.score_h, -11), 11)
            ELSE LEAST(GREATEST((p.score_h + COALESCE(p.runs, 0)) - p.score_v, -11), 11)
        END as score_diff_after,
        -- Inning capped at 9 for WE table
        LEAST(p.inning, 9) as we_inning,
        -- Check if inning changed (3 outs)
        CASE WHEN p.outs_post = 0 AND p.outs_pre = 2 THEN true ELSE false END as inning_changed,
        SUBSTRING(p.date FROM 1 FOR 4)::int as year_id
    FROM plays p
    WHERE
        (p.batter = $1 OR p.pitcher = $1)
        AND SUBSTRING(p.date FROM 1 FOR 4)::int = $2
        AND p.pa = 1
),
pas_with_we AS (
    SELECT
        pp.*,
        COALESCE(we_before.win_probability, 0.5) as we_before,
        COALESCE(we_after.win_probability, 0.5) as we_after
    FROM player_pas pp
    LEFT JOIN win_expectancy_historical we_before
        ON we_before.inning = pp.we_inning
        AND we_before.is_bottom = (pp.top_bot = 1)
        AND we_before.outs = pp.outs_pre
        AND we_before.runners_state = pp.runners_state_before
        AND we_before.score_diff = pp.score_diff_before
        AND (we_before.start_year IS NULL OR pp.year_id >= we_before.start_year)
        AND (we_before.end_year IS NULL OR pp.year_id <= we_before.end_year)
    LEFT JOIN win_expectancy_historical we_after
        ON we_after.inning = CASE
            WHEN pp.inning_changed AND pp.top_bot = 1 THEN LEAST(pp.inning + 1, 9)
            WHEN pp.inning_changed AND pp.top_bot = 0 THEN pp.we_inning
            ELSE pp.we_inning
        END
        AND we_after.is_bottom = CASE
            WHEN pp.inning_changed THEN NOT (pp.top_bot = 1)
            ELSE (pp.top_bot = 1)
        END
        AND we_after.outs = pp.outs_post
        AND we_after.runners_state = pp.runners_state_after
        AND we_after.score_diff = pp.score_diff_after
        AND (we_after.start_year IS NULL OR pp.year_id >= we_after.start_year)
        AND (we_after.end_year IS NULL OR pp.year_id <= we_after.end_year)
)
SELECT
    game_id,
    event_id,
    inning,
    top_bot,
    score_home,
    score_vis,
    outs_pre,
    bases_before,
    batter_id,
    pitcher_id,
    description,
    -- Simplified LI: proportional to variance in WE across possible outcomes
    CASE
        WHEN inning >= 9 THEN 2.0
        WHEN inning >= 7 THEN 1.5
        ELSE 1.0
    END *
    CASE
        WHEN ABS(score_diff_before) <= 1 THEN 1.5
        WHEN ABS(score_diff_before) <= 2 THEN 1.2
        WHEN ABS(score_diff_before) >= 5 THEN 0.3
        ELSE 1.0
    END *
    (1.0 + 0.2 * (LENGTH(runners_state_before) - LENGTH(REPLACE(runners_state_before, '_', '')))) *
    CASE WHEN outs_pre = 2 THEN 1.3 ELSE 1.0 END
    as leverage_index,
    we_before,
    we_after,
    (we_after - we_before) as we_change
FROM pas_with_we
WHERE
    CASE
        WHEN inning >= 9 THEN 2.0
        WHEN inning >= 7 THEN 1.5
        ELSE 1.0
    END *
    CASE
        WHEN ABS(score_diff_before) <= 1 THEN 1.5
        WHEN ABS(score_diff_before) <= 2 THEN 1.2
        WHEN ABS(score_diff_before) >= 5 THEN 0.3
        ELSE 1.0
    END *
    (1.0 + 0.2 * (LENGTH(runners_state_before) - LENGTH(REPLACE(runners_state_before, '_', '')))) *
    CASE WHEN outs_pre = 2 THEN 1.3 ELSE 1.0 END
    >= $3
ORDER BY game_id, event_id
