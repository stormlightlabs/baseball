-- Get top N pitchers by advanced pitching stat for a season
WITH player_pitching AS (
    SELECT
        "playerID",
        -- Counting stats (aggregated across stints)
        SUM("IPouts") as IPouts,
        SUM("BFP") as BF,
        SUM("H") as H,
        SUM("R") as R,
        SUM("ER") as ER,
        SUM("HR") as HR,
        SUM("BB") as BB,
        SUM("IBB") as IBB,
        SUM("HBP") as HBP,
        SUM("SO") as SO,
        -- Use first stint for team if no filter provided
        MAX("teamID") as teamID
    FROM "Pitching"
    WHERE "yearID" = $1
        AND (CAST($2 AS TEXT) IS NULL OR "teamID" = CAST($2 AS TEXT))
    GROUP BY "playerID"
    HAVING SUM("IPouts") >= $3  -- minimum IP threshold (in outs, so 3 * innings)
)
SELECT
    pp."playerID",
    -- Counting stats
    pp.IPouts,
    pp.BF,
    pp.H,
    pp.R,
    pp.ER,
    pp.HR,
    pp.BB,
    pp.IBB,
    pp.HBP,
    pp.SO,
    -- ERA
    ROUND(9.0 * pp.ER::numeric / NULLIF(pp.IPouts::numeric / 3, 0), 2) as ERA,
    -- WHIP
    ROUND((pp.BB + pp.H)::numeric / NULLIF(pp.IPouts::numeric / 3, 0), 3) as WHIP,
    -- Per-9 rates
    ROUND(9.0 * pp.SO::numeric / NULLIF(pp.IPouts::numeric / 3, 0), 2) as K_per_9,
    ROUND(9.0 * pp.BB::numeric / NULLIF(pp.IPouts::numeric / 3, 0), 2) as BB_per_9,
    ROUND(9.0 * pp.HR::numeric / NULLIF(pp.IPouts::numeric / 3, 0), 2) as HR_per_9,
    -- FIP (using constant C = 3.10 as approximation)
    ROUND(
        (13.0 * pp.HR + 3.0 * (pp.BB + pp.HBP) - 2.0 * pp.SO)::numeric /
        NULLIF(pp.IPouts::numeric / 3, 0) + 3.10,
        2
    ) as FIP,
    pp.teamID
FROM player_pitching pp
ORDER BY
    CASE $4
        WHEN 'IPOUTS' THEN pp.IPouts
        WHEN 'BF' THEN pp.BF
        WHEN 'H' THEN pp.H
        WHEN 'R' THEN pp.R
        WHEN 'ER' THEN pp.ER
        WHEN 'HR' THEN pp.HR
        WHEN 'BB' THEN pp.BB
        WHEN 'SO' THEN pp.SO
        ELSE 0
    END DESC,
    CASE $4
        WHEN 'ERA' THEN ROUND(9.0 * pp.ER::numeric / NULLIF(pp.IPouts::numeric / 3, 0), 2)
        WHEN 'WHIP' THEN ROUND((pp.BB + pp.H)::numeric / NULLIF(pp.IPouts::numeric / 3, 0), 3)
        WHEN 'K_PER_9' THEN ROUND(9.0 * pp.SO::numeric / NULLIF(pp.IPouts::numeric / 3, 0), 2)
        WHEN 'BB_PER_9' THEN ROUND(9.0 * pp.BB::numeric / NULLIF(pp.IPouts::numeric / 3, 0), 2)
        WHEN 'HR_PER_9' THEN ROUND(9.0 * pp.HR::numeric / NULLIF(pp.IPouts::numeric / 3, 0), 2)
        WHEN 'FIP' THEN ROUND(
            (13.0 * pp.HR + 3.0 * (pp.BB + pp.HBP) - 2.0 * pp.SO)::numeric /
            NULLIF(pp.IPouts::numeric / 3, 0) + 3.10,
            2
        )
        ELSE 0
    END ASC,  -- For rate stats like ERA, FIP, lower is better
    pp."playerID"
LIMIT $5;
