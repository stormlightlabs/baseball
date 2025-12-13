-- Calculate base running runs (wSB) for a player in a season
SELECT
    "playerID",
    "yearID" as season,
    "teamID",
    "lgID" as league,

    -- Counting stats
    "SB" as sb,
    "CS" as cs,

    -- Weighted stolen base runs (wSB)
    ROUND(
        ("SB" * wc.run_sb + "CS" * wc.run_cs)::numeric,
        2
    ) as baserunning_runs

FROM "Batting"
INNER JOIN woba_constants wc ON wc.season = "yearID"
WHERE "playerID" = $1
    AND "yearID" = $2
    AND (CAST($3 AS TEXT) IS NULL OR "teamID" = CAST($3 AS TEXT))
ORDER BY "stint"
LIMIT 1;
