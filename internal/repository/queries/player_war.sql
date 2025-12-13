-- Calculate complete WAR for a position player
-- Components: Batting (wRAA) + Baserunning (wSB) + Fielding + Positional Adj + Replacement Level
WITH batting_stats AS (
    -- wRAA from batting
    SELECT
        "playerID",
        "yearID",
        "teamID",
        "lgID",
        ("AB" + "BB" + "HBP" + "SF") as PA,
        ROUND(
            ((wc.w_bb * ("BB" - "IBB") + wc.w_hbp * "HBP" + wc.w_1b * ("H" - "2B" - "3B" - "HR") +
              wc.w_2b * "2B" + wc.w_3b * "3B" + wc.w_hr * "HR")::numeric /
             NULLIF("AB" + "BB" - "IBB" + "SF" + "HBP", 0) - wc.woba) /
            wc.woba_scale * ("AB" + "BB" + "HBP" + "SF"),
            2
        ) as batting_runs
    FROM "Batting"
    INNER JOIN woba_constants wc ON wc.season = "yearID"
    WHERE "playerID" = $1 AND "yearID" = $2
        AND (CAST($3 AS TEXT) IS NULL OR "teamID" = CAST($3 AS TEXT))
    ORDER BY "stint"
    LIMIT 1
),
baserunning_stats AS (
    -- wSB from base running
    SELECT
        "playerID",
        ROUND(("SB" * wc.run_sb + "CS" * wc.run_cs)::numeric, 2) as baserunning_runs
    FROM "Batting"
    INNER JOIN woba_constants wc ON wc.season = "yearID"
    WHERE "playerID" = $1 AND "yearID" = $2
        AND (CAST($3 AS TEXT) IS NULL OR "teamID" = CAST($3 AS TEXT))
    ORDER BY "stint"
    LIMIT 1
),
fielding_stats AS (
    -- Fielding runs from range factor
    SELECT
        f."POS" as position,
        f."G" as fielding_games,
        ROUND(
            (((f."PO" + f."A")::numeric / NULLIF(f."G", 0) -
              COALESCE(la.avg_rf, (f."PO" + f."A")::numeric / NULLIF(f."G", 0))) *
             f."G" * 0.1)::numeric,
            2
        ) as fielding_runs
    FROM "Fielding" f
    LEFT JOIN (
        SELECT "yearID", "lgID", "POS",
               AVG((("PO" + "A")::numeric / NULLIF("G", 0))) as avg_rf
        FROM "Fielding"
        WHERE "yearID" = $2 AND "G" > 0 AND "POS" NOT IN ('P', 'DH')
        GROUP BY "yearID", "lgID", "POS"
    ) la ON la."yearID" = f."yearID" AND la."lgID" = f."lgID" AND la."POS" = f."POS"
    WHERE f."playerID" = $1 AND f."yearID" = $2
        AND (CAST($3 AS TEXT) IS NULL OR f."teamID" = CAST($3 AS TEXT))
        AND f."POS" NOT IN ('P', 'DH')
    ORDER BY f."G" DESC
    LIMIT 1
)
SELECT
    bs."playerID",
    bs."yearID" as season,
    bs."teamID",
    bs."lgID" as league,
    bs.PA,

    -- Individual components
    bs.batting_runs,
    COALESCE(br.baserunning_runs, 0) as baserunning_runs,
    COALESCE(fs.fielding_runs, 0) as fielding_runs,
    COALESCE(pac.runs_per_162, 0) as positional_adjustment,

    -- Replacement level (simplified: -20 runs per 600 PA)
    ROUND((bs.PA * -0.0333)::numeric, 2) as replacement_runs,

    -- Total runs above replacement
    ROUND(
        (bs.batting_runs +
         COALESCE(br.baserunning_runs, 0) +
         COALESCE(fs.fielding_runs, 0) +
         COALESCE(pac.runs_per_162 * COALESCE(fs.fielding_games, 0) / 162.0, 0) -
         (bs.PA * 0.0333))::numeric,
        2
    ) as runs_above_replacement,

    -- WAR (runs above replacement / runs per win)
    ROUND(
        ((bs.batting_runs +
          COALESCE(br.baserunning_runs, 0) +
          COALESCE(fs.fielding_runs, 0) +
          COALESCE(pac.runs_per_162 * COALESCE(fs.fielding_games, 0) / 162.0, 0) -
          (bs.PA * 0.0333)) / wc.r_w)::numeric,
        1
    ) as WAR,

    fs.position
FROM batting_stats bs
LEFT JOIN baserunning_stats br ON br."playerID" = bs."playerID"
LEFT JOIN fielding_stats fs ON TRUE
LEFT JOIN positional_adjustment_constants pac ON
    pac.position = CASE
        WHEN fs.position = 'OF' THEN 'CF'  -- Use CF adjustment for generic OF
        ELSE fs.position
    END
INNER JOIN woba_constants wc ON wc.season = bs."yearID";
