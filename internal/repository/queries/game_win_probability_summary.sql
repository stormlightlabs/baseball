WITH game_info AS (
    SELECT
        game_id,
        SUBSTRING(date FROM 1 FOR 4)::int as year_id,
        home_team as home_team_id,
        visiting_team as away_team_id
    FROM games
    WHERE game_id = $1
),
plate_appearances AS (
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
        -- Score diff from batting team perspective (after)
        CASE
            WHEN p.top_bot = 0 THEN LEAST(GREATEST((p.score_v + COALESCE(p.runs, 0)) - p.score_h, -11), 11)
            ELSE LEAST(GREATEST((p.score_h + COALESCE(p.runs, 0)) - p.score_v, -11), 11)
        END as score_diff_after,
        LEAST(p.inning, 9) as we_inning,
        CASE WHEN p.outs_post = 0 AND p.outs_pre = 2 THEN true ELSE false END as inning_changed,
        SUBSTRING(p.date FROM 1 FOR 4)::int as year_id
    FROM plays p
    WHERE p.gid = $1
    AND p.pa = 1
    ORDER BY p.pn
),
pas_with_we AS (
    SELECT
        pp.*,
        COALESCE(we_before.win_probability, 0.5) as we_before,
        COALESCE(we_after.win_probability, 0.5) as we_after,
        -- Calculate WE change from home team perspective
        CASE
            WHEN pp.top_bot = 1 THEN COALESCE(we_after.win_probability, 0.5) - COALESCE(we_before.win_probability, 0.5)
            ELSE COALESCE(we_before.win_probability, 0.5) - COALESCE(we_after.win_probability, 0.5)
        END as we_change
    FROM plate_appearances pp
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
),
first_last_we AS (
    SELECT
        MIN(event_id) as first_event,
        MAX(event_id) as last_event
    FROM pas_with_we
),
max_swings AS (
    SELECT
        MAX(we_change) as max_positive,
        MIN(we_change) as max_negative
    FROM pas_with_we
),
positive_swing_pa AS (
    SELECT
        pw.event_id,
        pw.inning,
        pw.top_bot,
        pw.batter_id,
        pw.pitcher_id,
        pw.score_home as home_score_before,
        pw.score_vis as away_score_before,
        pw.outs_pre as outs_before,
        pw.bases_before,
        pw.description,
        pw.we_before,
        pw.we_after,
        pw.we_change
    FROM pas_with_we pw
    CROSS JOIN max_swings ms
    WHERE pw.we_change = ms.max_positive
    LIMIT 1
),
negative_swing_pa AS (
    SELECT
        pw.event_id,
        pw.inning,
        pw.top_bot,
        pw.batter_id,
        pw.pitcher_id,
        pw.score_home as home_score_before,
        pw.score_vis as away_score_before,
        pw.outs_pre as outs_before,
        pw.bases_before,
        pw.description,
        pw.we_before,
        pw.we_after,
        pw.we_change
    FROM pas_with_we pw
    CROSS JOIN max_swings ms
    WHERE pw.we_change = ms.max_negative
    LIMIT 1
)
SELECT
    gi.game_id,
    gi.year_id as season,
    gi.home_team_id,
    gi.away_team_id,
    (SELECT we_before FROM pas_with_we WHERE event_id = (SELECT first_event FROM first_last_we) LIMIT 1) as home_win_prob_start,
    (SELECT we_after FROM pas_with_we WHERE event_id = (SELECT last_event FROM first_last_we) LIMIT 1) as home_win_prob_end,
    pos.event_id as pos_event_id,
    pos.inning as pos_inning,
    pos.top_bot as pos_top_bot,
    pos.home_score_before as pos_home_score_before,
    pos.away_score_before as pos_away_score_before,
    pos.outs_before as pos_outs_before,
    pos.bases_before as pos_bases_before,
    pos.batter_id as pos_batter_id,
    pos.pitcher_id as pos_pitcher_id,
    pos.description as pos_description,
    pos.we_before as pos_we_before,
    pos.we_after as pos_we_after,
    pos.we_change as pos_we_change,
    neg.event_id as neg_event_id,
    neg.inning as neg_inning,
    neg.top_bot as neg_top_bot,
    neg.home_score_before as neg_home_score_before,
    neg.away_score_before as neg_away_score_before,
    neg.outs_before as neg_outs_before,
    neg.bases_before as neg_bases_before,
    neg.batter_id as neg_batter_id,
    neg.pitcher_id as neg_pitcher_id,
    neg.description as neg_description,
    neg.we_before as neg_we_before,
    neg.we_after as neg_we_after,
    neg.we_change as neg_we_change
FROM game_info gi
CROSS JOIN positive_swing_pa pos
CROSS JOIN negative_swing_pa neg
LIMIT 1
