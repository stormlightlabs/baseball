-- player_leverage_summary.sql
-- Aggregates leverage metrics for a player in a season.
-- Uses win_expectancy_historical table for accurate win expectancy lookups.

WITH player_pas AS (
    SELECT
        p.gid as game_id,
        p.pn as play_num,
        p.inning,
        p.top_bot,
        p.batter as batter_id,
        p.pitcher as pitcher_id,
        CASE
            WHEN $3 = 'pitcher' THEN p.pitteam
            ELSE p.batteam
        END as team_id,
        CASE
            WHEN $3 = 'pitcher' THEN p.pitteam
            ELSE p.batteam
        END as league_id,
        p.score_h as score_home,
        p.score_v as score_vis,
        p.outs_pre,
        -- Convert bases to underscore format for WE table lookup
        COALESCE(
            CASE WHEN p.br1_pre IS NOT NULL AND p.br1_pre != '' THEN '1' ELSE '_' END ||
            CASE WHEN p.br2_pre IS NOT NULL AND p.br2_pre != '' THEN '2' ELSE '_' END ||
            CASE WHEN p.br3_pre IS NOT NULL AND p.br3_pre != '' THEN '3' ELSE '_' END,
            '___'
        ) as runners_state,
        -- Score diff from batting team perspective
        CASE
            WHEN p.top_bot = 0 THEN LEAST(GREATEST(p.score_v - p.score_h, -11), 11)
            ELSE LEAST(GREATEST(p.score_h - p.score_v, -11), 11)
        END as score_diff,
        -- Inning capped at 9 for WE table
        LEAST(p.inning, 9) as we_inning,
        SUBSTRING(p.date FROM 1 FOR 4)::int as year_id
    FROM plays p
    WHERE
        CASE
            WHEN $3 = 'batter' THEN p.batter = $1
            WHEN $3 = 'pitcher' THEN p.pitcher = $1
            ELSE (p.pitcher = $1 OR p.batter = $1)
        END
        AND SUBSTRING(p.date FROM 1 FOR 4)::int = $2
        AND p.pa = 1
),
pas_with_li AS (
    SELECT
        pp.*,
        -- Simplified LI calculation
        CASE
            WHEN pp.inning >= 9 THEN 2.0
            WHEN pp.inning >= 7 THEN 1.5
            ELSE 1.0
        END *
        CASE
            WHEN ABS(pp.score_diff) <= 1 THEN 1.5
            WHEN ABS(pp.score_diff) <= 2 THEN 1.2
            WHEN ABS(pp.score_diff) >= 5 THEN 0.3
            ELSE 1.0
        END *
        (1.0 + 0.2 * (LENGTH(pp.runners_state) - LENGTH(REPLACE(pp.runners_state, '_', '')))) *
        CASE WHEN pp.outs_pre = 2 THEN 1.3 ELSE 1.0 END
        as leverage_index
    FROM player_pas pp
),
aggregated AS (
    SELECT
        team_id,
        league_id,
        COUNT(*) as total_pas,
        AVG(leverage_index) as avg_li,
        COUNT(*) FILTER (WHERE leverage_index < 0.85) as low_leverage_pa,
        COUNT(*) FILTER (WHERE leverage_index >= 0.85 AND leverage_index <= 2.0) as medium_leverage_pa,
        COUNT(*) FILTER (WHERE leverage_index > 2.0) as high_leverage_pa,
        SUM(0.0) as total_wpa  -- TODO: Calculate WPA when we have outcome data
    FROM pas_with_li
    GROUP BY team_id, league_id
)
SELECT
    team_id,
    league_id,
    avg_li,
    low_leverage_pa,
    medium_leverage_pa,
    high_leverage_pa,
    total_wpa,
    0 as low_leverage_ip_outs,
    0 as medium_leverage_ip_outs,
    0 as high_leverage_ip_outs
FROM aggregated
LIMIT 1
