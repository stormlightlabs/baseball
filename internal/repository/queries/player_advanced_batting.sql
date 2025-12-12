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
    -- wOBA (simplified, without year-specific weights)
    -- Using approximate constant weights: wBB=0.69, wHBP=0.72, w1B=0.88, w2B=1.24, w3B=1.56, wHR=2.08
    ROUND(
        (0.69 * ("BB" - "IBB") + 0.72 * "HBP" + 0.88 * ("H" - "2B" - "3B" - "HR") +
         1.24 * "2B" + 1.56 * "3B" + 2.08 * "HR")::numeric /
        NULLIF("AB" + "BB" - "IBB" + "SF" + "HBP", 0),
        3
    ) as wOBA,
    "teamID"
FROM "Batting"
WHERE "playerID" = $1
    AND "yearID" = $2
    AND (CAST($3 AS TEXT) IS NULL OR "teamID" = CAST($3 AS TEXT))
ORDER BY "stint"
LIMIT 1;
