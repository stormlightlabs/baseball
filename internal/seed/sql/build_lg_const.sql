-- NOTE:
-- This runtime ETL query mirrors the same league-constants derivation approach as
-- internal/db/sql/018_populate_league_constants.sql, but without a fixed season filter.
WITH league_stats AS (
    SELECT
        b."yearID" AS season,
        b."lgID" AS league,
        SUM((b."AB" + b."BB" + b."HBP" + b."SF"))::BIGINT AS total_pa,
        SUM(b."R")::BIGINT AS total_runs,
        SUM(CASE
            WHEN (b."AB" + b."BB" + b."HBP" + b."SF") >= 100
            THEN (b."AB" + b."BB" - b."IBB" + b."SF" + b."HBP")
            ELSE 0
        END)::NUMERIC AS qualified_pa_denom,
        SUM(CASE
            WHEN (b."AB" + b."BB" + b."HBP" + b."SF") >= 100
            THEN
                wc.w_bb * (b."BB" - b."IBB") +
                wc.w_hbp * b."HBP" +
                wc.w_1b * (b."H" - b."2B" - b."3B" - b."HR") +
                wc.w_2b * b."2B" +
                wc.w_3b * b."3B" +
                wc.w_hr * b."HR"
            ELSE 0
        END)::NUMERIC AS qualified_woba_numerator
    FROM "Batting" b
    INNER JOIN woba_constants wc ON wc.season = b."yearID"
    WHERE b."lgID" IN ('AL', 'NL')
    GROUP BY b."yearID", b."lgID"
),
league_calcs AS (
    SELECT
        season,
        league,
        total_pa,
        total_runs,
        CASE
            WHEN qualified_pa_denom > 0
            THEN ROUND((qualified_woba_numerator / qualified_pa_denom)::NUMERIC, 3)
            ELSE NULL
        END AS woba_avg,
        CASE
            WHEN total_pa > 0
            THEN ROUND((total_runs::NUMERIC / total_pa)::NUMERIC, 4)
            ELSE NULL
        END AS wrc_per_pa
    FROM league_stats
)
INSERT INTO league_constants (
    season, league, woba_avg, wrc_per_pa, runs_per_win, total_pa, total_runs, calculated_at
)
SELECT
    lc.season,
    lc.league,
    COALESCE(lc.woba_avg, wc.woba),
    COALESCE(lc.wrc_per_pa, wc.r_pa),
    wc.r_w,
    lc.total_pa,
    lc.total_runs,
    NOW()
FROM league_calcs lc
INNER JOIN woba_constants wc ON wc.season = lc.season
ON CONFLICT (season, league) DO UPDATE SET
    woba_avg = EXCLUDED.woba_avg,
    wrc_per_pa = EXCLUDED.wrc_per_pa,
    runs_per_win = EXCLUDED.runs_per_win,
    total_pa = EXCLUDED.total_pa,
    total_runs = EXCLUDED.total_runs,
    calculated_at = EXCLUDED.calculated_at;
