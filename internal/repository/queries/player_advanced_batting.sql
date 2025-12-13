-- Get advanced batting stats for a player in a season
SELECT
    -- Counting stats
    ("AB" + "BB" + "HBP" + "SF") as PA,
    "AB",
    "H",
    "2B" as doubles,
    "3B" as triples,
    "HR",
    "BB",
    "IBB",
    "HBP",
    "SF",
    "SH",
    "SO",
    -- Basic rate stats
    ROUND("H"::numeric / NULLIF("AB", 0), 3) as AVG,
    ROUND(("H" + "BB" + "HBP")::numeric / NULLIF("AB" + "BB" + "HBP" + "SF", 0), 3) as OBP,
    ROUND(("H" + "2B" + 2*"3B" + 3*"HR")::numeric / NULLIF("AB", 0), 3) as SLG,
    -- ISO
    ROUND(("H" + "2B" + 2*"3B" + 3*"HR")::numeric / NULLIF("AB", 0) - "H"::numeric / NULLIF("AB", 0), 3) as ISO,
    -- BABIP
    ROUND(("H" - "HR")::numeric / NULLIF("AB" - "SO" - "HR" + "SF", 0), 3) as BABIP,
    -- Rate stats
    ROUND("SO"::numeric / NULLIF("AB" + "BB" + "HBP" + "SF", 0), 3) as K_rate,
    ROUND("BB"::numeric / NULLIF("AB" + "BB" + "HBP" + "SF", 0), 3) as BB_rate,
    -- wOBA (using year-specific FanGraphs constants)
    ROUND(
        (wc.w_bb * ("BB" - "IBB") + wc.w_hbp * "HBP" + wc.w_1b * ("H" - "2B" - "3B" - "HR") +
         wc.w_2b * "2B" + wc.w_3b * "3B" + wc.w_hr * "HR")::numeric /
        NULLIF("AB" + "BB" - "IBB" + "SF" + "HBP", 0),
        3
    ) as wOBA,
    -- wRAA (weighted runs above average)
    ROUND(
        ((wc.w_bb * ("BB" - "IBB") + wc.w_hbp * "HBP" + wc.w_1b * ("H" - "2B" - "3B" - "HR") +
          wc.w_2b * "2B" + wc.w_3b * "3B" + wc.w_hr * "HR")::numeric /
         NULLIF("AB" + "BB" - "IBB" + "SF" + "HBP", 0) - wc.woba) /
        wc.woba_scale * ("AB" + "BB" + "HBP" + "SF"),
        2
    ) as wRAA,
    -- wRC (weighted runs created)
    ROUND(
        ((wc.w_bb * ("BB" - "IBB") + wc.w_hbp * "HBP" + wc.w_1b * ("H" - "2B" - "3B" - "HR") +
          wc.w_2b * "2B" + wc.w_3b * "3B" + wc.w_hr * "HR")::numeric /
         NULLIF("AB" + "BB" - "IBB" + "SF" + "HBP", 0) - wc.woba) /
        wc.woba_scale * ("AB" + "BB" + "HBP" + "SF") + (lc.wrc_per_pa * ("AB" + "BB" + "HBP" + "SF")),
        2
    ) as wRC,
    -- wRC+ (park and league adjusted, 100 = league average)
    ROUND(
        (((((wc.w_bb * ("BB" - "IBB") + wc.w_hbp * "HBP" + wc.w_1b * ("H" - "2B" - "3B" - "HR") +
            wc.w_2b * "2B" + wc.w_3b * "3B" + wc.w_hr * "HR")::numeric /
           NULLIF("AB" + "BB" - "IBB" + "SF" + "HBP", 0) - wc.woba) /
          wc.woba_scale + lc.wrc_per_pa) +
          (lc.wrc_per_pa - COALESCE(pf.basic_5yr, 100) / 100.0 * lc.wrc_per_pa)) /
         lc.wrc_per_pa) * 100,
        0
    )::int as wRC_plus,
    "teamID",
    "lgID"
FROM "Batting"
INNER JOIN woba_constants wc ON wc.season = "yearID"
INNER JOIN league_constants lc ON lc.season = "yearID" AND lc.league = "lgID"
LEFT JOIN fangraphs_team_park_map pm ON pm.retrosheet_team_id = "teamID"
    AND "yearID" BETWEEN COALESCE(pm.start_year, 1871) AND COALESCE(pm.end_year, 2100)
LEFT JOIN park_factors pf ON pf.park_id = pm.primary_park_id AND pf.season = "yearID"
WHERE "playerID" = $1
    AND "yearID" = $2
    AND (CAST($3 AS TEXT) IS NULL OR "teamID" = CAST($3 AS TEXT))
ORDER BY "stint"
LIMIT 1;
