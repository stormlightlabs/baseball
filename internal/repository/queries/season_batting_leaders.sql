-- Get top N players by advanced batting stat for a season
WITH player_batting AS (
    SELECT
        "playerID",
        -- Counting stats (aggregated across stints)
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
        SUM("SH") as SH,
        SUM("SO") as SO,
        -- Use first stint for team/league if no filter provided
        MAX("teamID") as teamID,
        MAX("lgID") as lgID
    FROM "Batting"
    WHERE "yearID" = $1
        AND (CAST($2 AS TEXT) IS NULL OR "teamID" = CAST($2 AS TEXT))
        AND (CAST($3 AS TEXT) IS NULL OR "lgID" = CAST($3 AS TEXT))
    GROUP BY "playerID"
    HAVING SUM("AB" + "BB" + "HBP" + "SF") >= $4  -- minimum PA threshold
)
SELECT
    pb."playerID",
    -- Counting stats
    pb.PA,
    pb.AB,
    pb.H,
    pb.doubles,
    pb.triples,
    pb.HR,
    pb.BB,
    pb.IBB,
    pb.HBP,
    pb.SF,
    pb.SH,
    pb.SO,
    -- Basic rate stats
    ROUND(pb.H::numeric / NULLIF(pb.AB, 0), 3) as AVG,
    ROUND((pb.H + pb.BB + pb.HBP)::numeric / NULLIF(pb.PA, 0), 3) as OBP,
    ROUND((pb.H + pb.doubles + 2*pb.triples + 3*pb.HR)::numeric / NULLIF(pb.AB, 0), 3) as SLG,
    -- ISO
    ROUND((pb.H + pb.doubles + 2*pb.triples + 3*pb.HR)::numeric / NULLIF(pb.AB, 0) - pb.H::numeric / NULLIF(pb.AB, 0), 3) as ISO,
    -- BABIP
    ROUND((pb.H - pb.HR)::numeric / NULLIF(pb.AB - pb.SO - pb.HR + pb.SF, 0), 3) as BABIP,
    -- Rate stats
    ROUND(pb.SO::numeric / NULLIF(pb.PA, 0), 3) as K_rate,
    ROUND(pb.BB::numeric / NULLIF(pb.PA, 0), 3) as BB_rate,
    -- wOBA (using year-specific FanGraphs constants)
    ROUND(
        (wc.w_bb * (pb.BB - pb.IBB) + wc.w_hbp * pb.HBP + wc.w_1b * (pb.H - pb.doubles - pb.triples - pb.HR) +
         wc.w_2b * pb.doubles + wc.w_3b * pb.triples + wc.w_hr * pb.HR)::numeric /
        NULLIF(pb.AB + pb.BB - pb.IBB + pb.SF + pb.HBP, 0),
        3
    ) as wOBA,
    -- wRAA (weighted runs above average)
    ROUND(
        ((wc.w_bb * (pb.BB - pb.IBB) + wc.w_hbp * pb.HBP + wc.w_1b * (pb.H - pb.doubles - pb.triples - pb.HR) +
          wc.w_2b * pb.doubles + wc.w_3b * pb.triples + wc.w_hr * pb.HR)::numeric /
         NULLIF(pb.AB + pb.BB - pb.IBB + pb.SF + pb.HBP, 0) - wc.woba) /
        wc.woba_scale * pb.PA,
        2
    ) as wRAA,
    -- wRC (weighted runs created)
    ROUND(
        ((wc.w_bb * (pb.BB - pb.IBB) + wc.w_hbp * pb.HBP + wc.w_1b * (pb.H - pb.doubles - pb.triples - pb.HR) +
          wc.w_2b * pb.doubles + wc.w_3b * pb.triples + wc.w_hr * pb.HR)::numeric /
         NULLIF(pb.AB + pb.BB - pb.IBB + pb.SF + pb.HBP, 0) - wc.woba) /
        wc.woba_scale * pb.PA + (lc.wrc_per_pa * pb.PA),
        2
    ) as wRC,
    -- wRC+ (park and league adjusted, 100 = league average)
    ROUND(
        (((((wc.w_bb * (pb.BB - pb.IBB) + wc.w_hbp * pb.HBP + wc.w_1b * (pb.H - pb.doubles - pb.triples - pb.HR) +
            wc.w_2b * pb.doubles + wc.w_3b * pb.triples + wc.w_hr * pb.HR)::numeric /
           NULLIF(pb.AB + pb.BB - pb.IBB + pb.SF + pb.HBP, 0) - wc.woba) /
          wc.woba_scale + lc.wrc_per_pa) +
          (lc.wrc_per_pa - COALESCE(pf.basic_5yr, 100) / 100.0 * lc.wrc_per_pa)) /
         lc.wrc_per_pa) * 100,
        0
    )::int as wRC_plus,
    pb.teamID,
    pb.lgID
FROM player_batting pb
INNER JOIN woba_constants wc ON wc.season = $1
INNER JOIN league_constants lc ON lc.season = $1 AND lc.league = pb.lgID
LEFT JOIN fangraphs_team_park_map pm ON pm.retrosheet_team_id = pb.teamID
    AND $1 BETWEEN COALESCE(pm.start_year, 1871) AND COALESCE(pm.end_year, 2100)
LEFT JOIN park_factors pf ON pf.park_id = pm.primary_park_id AND pf.season = $1
ORDER BY
    CASE $5
        WHEN 'PA' THEN pb.PA
        WHEN 'AB' THEN pb.AB
        WHEN 'H' THEN pb.H
        WHEN 'HR' THEN pb.HR
        WHEN 'BB' THEN pb.BB
        WHEN 'SO' THEN pb.SO
        ELSE 0
    END DESC,
    CASE $5
        WHEN 'AVG' THEN ROUND(pb.H::numeric / NULLIF(pb.AB, 0), 3)
        WHEN 'OBP' THEN ROUND((pb.H + pb.BB + pb.HBP)::numeric / NULLIF(pb.PA, 0), 3)
        WHEN 'SLG' THEN ROUND((pb.H + pb.doubles + 2*pb.triples + 3*pb.HR)::numeric / NULLIF(pb.AB, 0), 3)
        WHEN 'ISO' THEN ROUND((pb.H + pb.doubles + 2*pb.triples + 3*pb.HR)::numeric / NULLIF(pb.AB, 0) - pb.H::numeric / NULLIF(pb.AB, 0), 3)
        WHEN 'BABIP' THEN ROUND((pb.H - pb.HR)::numeric / NULLIF(pb.AB - pb.SO - pb.HR + pb.SF, 0), 3)
        WHEN 'K_RATE' THEN ROUND(pb.SO::numeric / NULLIF(pb.PA, 0), 3)
        WHEN 'BB_RATE' THEN ROUND(pb.BB::numeric / NULLIF(pb.PA, 0), 3)
        WHEN 'WOBA' THEN ROUND(
            (wc.w_bb * (pb.BB - pb.IBB) + wc.w_hbp * pb.HBP + wc.w_1b * (pb.H - pb.doubles - pb.triples - pb.HR) +
             wc.w_2b * pb.doubles + wc.w_3b * pb.triples + wc.w_hr * pb.HR)::numeric /
            NULLIF(pb.AB + pb.BB - pb.IBB + pb.SF + pb.HBP, 0),
            3
        )
        WHEN 'WRAA' THEN ROUND(
            ((wc.w_bb * (pb.BB - pb.IBB) + wc.w_hbp * pb.HBP + wc.w_1b * (pb.H - pb.doubles - pb.triples - pb.HR) +
              wc.w_2b * pb.doubles + wc.w_3b * pb.triples + wc.w_hr * pb.HR)::numeric /
             NULLIF(pb.AB + pb.BB - pb.IBB + pb.SF + pb.HBP, 0) - wc.woba) /
            wc.woba_scale * pb.PA,
            2
        )
        WHEN 'WRC' THEN ROUND(
            ((wc.w_bb * (pb.BB - pb.IBB) + wc.w_hbp * pb.HBP + wc.w_1b * (pb.H - pb.doubles - pb.triples - pb.HR) +
              wc.w_2b * pb.doubles + wc.w_3b * pb.triples + wc.w_hr * pb.HR)::numeric /
             NULLIF(pb.AB + pb.BB - pb.IBB + pb.SF + pb.HBP, 0) - wc.woba) /
            wc.woba_scale * pb.PA + (lc.wrc_per_pa * pb.PA),
            2
        )
        WHEN 'WRC_PLUS' THEN ROUND(
            (((((wc.w_bb * (pb.BB - pb.IBB) + wc.w_hbp * pb.HBP + wc.w_1b * (pb.H - pb.doubles - pb.triples - pb.HR) +
                wc.w_2b * pb.doubles + wc.w_3b * pb.triples + wc.w_hr * pb.HR)::numeric /
               NULLIF(pb.AB + pb.BB - pb.IBB + pb.SF + pb.HBP, 0) - wc.woba) /
              wc.woba_scale + lc.wrc_per_pa) +
              (lc.wrc_per_pa - COALESCE(pf.basic_5yr, 100) / 100.0 * lc.wrc_per_pa)) /
             lc.wrc_per_pa) * 100,
            0
        )::int
        ELSE 0
    END DESC,
    pb."playerID"
LIMIT $6;
