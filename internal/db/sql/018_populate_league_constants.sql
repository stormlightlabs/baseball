-- Populate league constants for wRC+ calculations (2023-2024)
-- Uses Lahman Batting table to calculate league-wide constants

-- Calculate league constants for each season/league
-- Excludes pitchers by filtering for players with >= 100 PA (standard threshold)
WITH league_stats AS (
    SELECT
        "yearID" as season,
        "lgID" as league,
        -- Total PA and runs (for reference)
        SUM("AB" + "BB" + "HBP" + "SF") as total_pa,
        SUM("R") as total_runs,
        -- For wOBA average calculation (exclude pitchers: >= 100 PA)
        SUM(CASE WHEN ("AB" + "BB" + "HBP" + "SF") >= 100 THEN "AB" + "BB" - "IBB" + "SF" + "HBP" ELSE 0 END) as qualified_pa_denom,
        SUM(CASE
            WHEN ("AB" + "BB" + "HBP" + "SF") >= 100 THEN
                wc.w_bb * ("BB" - "IBB") + wc.w_hbp * "HBP" +
                wc.w_1b * ("H" - "2B" - "3B" - "HR") +
                wc.w_2b * "2B" + wc.w_3b * "3B" + wc.w_hr * "HR"
            ELSE 0
        END) as qualified_woba_numerator
    FROM "Batting" b
    INNER JOIN woba_constants wc ON wc.season = b."yearID"
    WHERE "yearID" IN (2023, 2024)
        AND "lgID" IN ('AL', 'NL')
    GROUP BY "yearID", "lgID"
),
league_with_calcs AS (
    SELECT
        season,
        league,
        total_pa,
        total_runs,
        -- League average wOBA (excluding pitchers)
        ROUND((qualified_woba_numerator / NULLIF(qualified_pa_denom, 0))::numeric, 4) as woba_avg,
        -- Runs per PA
        ROUND((total_runs::numeric / NULLIF(total_pa, 0))::numeric, 5) as r_pa
    FROM league_stats
)
INSERT INTO league_constants (season, league, woba_avg, wrc_per_pa, runs_per_win, total_pa, total_runs)
SELECT
    lc.season,
    lc.league,
    lc.woba_avg,
    -- wRC per PA: approximately (wOBA - lgwOBA) / wOBA_scale + r/PA
    -- We use a simplified approach: r_pa as the baseline
    ROUND(lc.r_pa::numeric, 5) as wrc_per_pa,
    -- Runs per win: use FanGraphs value from woba_constants
    wc.r_w as runs_per_win,
    lc.total_pa,
    lc.total_runs
FROM league_with_calcs lc
INNER JOIN woba_constants wc ON wc.season = lc.season
ON CONFLICT (season, league) DO UPDATE SET
    woba_avg = EXCLUDED.woba_avg,
    wrc_per_pa = EXCLUDED.wrc_per_pa,
    runs_per_win = EXCLUDED.runs_per_win,
    total_pa = EXCLUDED.total_pa,
    total_runs = EXCLUDED.total_runs,
    calculated_at = NOW();

-- Verify the results
SELECT
    season,
    league,
    woba_avg,
    wrc_per_pa,
    runs_per_win,
    total_pa,
    total_runs
FROM league_constants
ORDER BY season DESC, league;
