-- Get advanced pitching stats for a player in a season
SELECT
    -- Counting stats
    "IPouts",
    "BFP" as BF,
    H,
    R,
    ER,
    HR,
    BB,
    IBB,
    HBP,
    SO,
    -- ERA
    ROUND(9.0 * ER::numeric / NULLIF("IPouts"::numeric / 3, 0), 2) as ERA,
    -- WHIP
    ROUND((BB + H)::numeric / NULLIF("IPouts"::numeric / 3, 0), 3) as WHIP,
    -- Per-9 rates
    ROUND(9.0 * SO::numeric / NULLIF("IPouts"::numeric / 3, 0), 2) as K_per_9,
    ROUND(9.0 * BB::numeric / NULLIF("IPouts"::numeric / 3, 0), 2) as BB_per_9,
    ROUND(9.0 * HR::numeric / NULLIF("IPouts"::numeric / 3, 0), 2) as HR_per_9,
    -- FIP (using constant C = 3.10 as approximation)
    ROUND(
        (13.0 * HR + 3.0 * (BB + HBP) - 2.0 * SO)::numeric /
        NULLIF("IPouts"::numeric / 3, 0) + 3.10,
        2
    ) as FIP,
    "teamID"
FROM "Pitching"
WHERE "playerID" = $1
    AND "yearID" = $2
    AND (CAST($3 AS TEXT) IS NULL OR "teamID" = CAST($3 AS TEXT))
ORDER BY "stint"
LIMIT 1;
