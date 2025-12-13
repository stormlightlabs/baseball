-- Calculate fielding runs for a player in a season using range factor
-- Compares player's range factor to league average for their position
WITH player_fielding AS (
    SELECT
        f."playerID" as "playerID",
        f."yearID" as season,
        f."teamID" as "teamID",
        f."lgID" as league,
        f."POS" as position,
        f."G" as games,
        f."PO" as putouts,
        f."A" as assists,
        f."E" as errors,
        -- Player's range factor (PO + A) per game
        CASE
            WHEN f."G" > 0 THEN
                (f."PO" + f."A")::numeric / NULLIF(f."G", 0)
            ELSE 0
        END as player_rf
    FROM "Fielding" f
    WHERE f."playerID" = $1
        AND f."yearID" = $2
        AND (CAST($3 AS TEXT) IS NULL OR f."teamID" = CAST($3 AS TEXT))
        AND f."POS" NOT IN ('P', 'DH') -- Exclude pitchers and DH from fielding
    ORDER BY f."G" DESC
    LIMIT 1 -- Get primary position (most games played)
),
league_avg_rf AS (
    SELECT
        "yearID",
        "lgID",
        "POS",
        AVG((("PO" + "A")::numeric / NULLIF("G", 0))) as avg_rf
    FROM "Fielding"
    WHERE "yearID" = $2
        AND "G" > 0
        AND "POS" NOT IN ('P', 'DH')
    GROUP BY "yearID", "lgID", "POS"
)
SELECT
    pf."playerID",
    pf.season,
    pf."teamID",
    pf.league,
    pf.position,
    pf.games,
    pf.putouts,
    pf.assists,
    pf.errors,
    ROUND(pf.player_rf, 3) as range_factor,
    ROUND(la.avg_rf, 3) as league_avg_rf,
    -- Fielding runs above average
    -- (Player RF - League Avg RF) × Games × runs_per_play_factor
    -- Using simplified factor of 0.1 runs per play above average
    ROUND(
        ((pf.player_rf - COALESCE(la.avg_rf, pf.player_rf)) * pf.games * 0.1)::numeric,
        2
    ) as fielding_runs
FROM player_fielding pf
LEFT JOIN league_avg_rf la ON
    la."yearID" = pf.season
    AND la."lgID" = pf.league
    AND la."POS" = pf.position;
