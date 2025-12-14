-- Get top N players by WAR for a season
WITH player_batting AS (
    SELECT
        "playerID",
        MAX("teamID") as "teamID",
        MAX("lgID") as "lgID",
        SUM("AB" + "BB" + "HBP" + "SF") as PA,
        SUM("AB") as AB,
        SUM("H") as H,
        SUM("2B") as doubles,
        SUM("3B") as triples,
        SUM("HR") as HR,
        SUM("BB") as BB,
        SUM("IBB") as IBB,
        SUM("HBP") as HBP,
        SUM("SF") as SF,
        SUM("SB") as SB,
        SUM("CS") as CS
    FROM "Batting"
    WHERE "yearID" = $1
        AND (CAST($2 AS TEXT) IS NULL OR "teamID" = CAST($2 AS TEXT))
    GROUP BY "playerID"
    HAVING SUM("AB" + "BB" + "HBP" + "SF") >= $3  -- minimum PA threshold
),
batting_runs AS (
    SELECT
        pb."playerID",
        pb."teamID",
        pb."lgID",
        pb.PA,
        -- wRAA from batting
        ROUND(
            ((wc.w_bb * (pb.BB - pb.IBB) + wc.w_hbp * pb.HBP + wc.w_1b * (pb.H - pb.doubles - pb.triples - pb.HR) +
              wc.w_2b * pb.doubles + wc.w_3b * pb.triples + wc.w_hr * pb.HR)::numeric /
             NULLIF(pb.AB + pb.BB - pb.IBB + pb.SF + pb.HBP, 0) - wc.woba) /
            wc.woba_scale * pb.PA,
            2
        ) as batting_runs,
        -- wSB from base running
        ROUND((pb.SB * wc.run_sb + pb.CS * wc.run_cs)::numeric, 2) as baserunning_runs
    FROM player_batting pb
    INNER JOIN woba_constants wc ON wc.season = $1
),
fielding_runs AS (
    SELECT
        f."playerID",
        f."POS" as position,
        SUM(f."G") as fielding_games,
        SUM(
            ROUND(
                (((f."PO" + f."A")::numeric / NULLIF(f."G", 0) -
                  COALESCE(la.avg_rf, (f."PO" + f."A")::numeric / NULLIF(f."G", 0))) *
                 f."G" * 0.1)::numeric,
                2
            )
        ) as fielding_runs
    FROM "Fielding" f
    LEFT JOIN (
        SELECT "yearID", "lgID", "POS",
               AVG((("PO" + "A")::numeric / NULLIF("G", 0))) as avg_rf
        FROM "Fielding"
        WHERE "yearID" = $1 AND "G" > 0 AND "POS" NOT IN ('P', 'DH')
        GROUP BY "yearID", "lgID", "POS"
    ) la ON la."yearID" = f."yearID" AND la."lgID" = f."lgID" AND la."POS" = f."POS"
    WHERE f."yearID" = $1
        AND (CAST($2 AS TEXT) IS NULL OR f."teamID" = CAST($2 AS TEXT))
        AND f."POS" NOT IN ('P', 'DH')
    GROUP BY f."playerID", f."POS"
),
primary_fielding AS (
    -- Get primary position (most games) for each player
    SELECT DISTINCT ON ("playerID")
        "playerID",
        position,
        fielding_games,
        fielding_runs
    FROM fielding_runs
    ORDER BY "playerID", fielding_games DESC
)
SELECT
    br."playerID",
    $1 as season,
    br."teamID",
    br."lgID" as league,
    br.PA,
    -- Individual components
    br.batting_runs,
    br.baserunning_runs,
    COALESCE(pf.fielding_runs, 0) as fielding_runs,
    COALESCE(pac.runs_per_162 * COALESCE(pf.fielding_games, 0) / 162.0, 0) as positional_adjustment,
    -- Replacement level (simplified: -20 runs per 600 PA)
    ROUND((br.PA * -0.0333)::numeric, 2) as replacement_runs,
    -- Total runs above replacement
    ROUND(
        (br.batting_runs +
         br.baserunning_runs +
         COALESCE(pf.fielding_runs, 0) +
         COALESCE(pac.runs_per_162 * COALESCE(pf.fielding_games, 0) / 162.0, 0) -
         (br.PA * 0.0333))::numeric,
        2
    ) as runs_above_replacement,
    -- WAR (runs above replacement / runs per win)
    ROUND(
        ((br.batting_runs +
          br.baserunning_runs +
          COALESCE(pf.fielding_runs, 0) +
          COALESCE(pac.runs_per_162 * COALESCE(pf.fielding_games, 0) / 162.0, 0) -
          (br.PA * 0.0333)) / wc.r_w)::numeric,
        1
    ) as WAR,
    pf.position
FROM batting_runs br
LEFT JOIN primary_fielding pf ON pf."playerID" = br."playerID"
LEFT JOIN positional_adjustment_constants pac ON
    pac.position = CASE
        WHEN pf.position = 'OF' THEN 'CF'  -- Use CF adjustment for generic OF
        ELSE pf.position
    END
INNER JOIN woba_constants wc ON wc.season = $1
ORDER BY
    ROUND(
        ((br.batting_runs +
          br.baserunning_runs +
          COALESCE(pf.fielding_runs, 0) +
          COALESCE(pac.runs_per_162 * COALESCE(pf.fielding_games, 0) / 162.0, 0) -
          (br.PA * 0.0333)) / wc.r_w)::numeric,
        1
    ) DESC,
    br."playerID"
LIMIT $4;
