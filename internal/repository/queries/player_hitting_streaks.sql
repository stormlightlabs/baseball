-- Find hitting streaks for a player in a given season
-- Parameters: $1 = player_id (retroID), $2 = season year (YYYY format as string prefix), $3 = min streak length

WITH game_hitting AS (
    SELECT
        batter,
        gid,
        date,
        SUM(ab) as at_bats,
        SUM(single + double + triple + hr) as hits
    FROM plays
    WHERE batter = $1
      AND date >= $2 || '0101'
      AND date <= $2 || '1231'
      AND batter IS NOT NULL
      AND batter != ''
    GROUP BY batter, gid, date
    HAVING SUM(ab) > 0  -- Only count games with at bats
),
with_streak_id AS (
    SELECT
        batter,
        gid,
        date,
        at_bats,
        hits,
        CASE WHEN hits > 0 THEN 1 ELSE 0 END as had_hit,
        -- Create streak groups: increment when hits = 0
        SUM(CASE WHEN hits = 0 THEN 1 ELSE 0 END) OVER (
            PARTITION BY batter
            ORDER BY date, gid
            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
        ) as streak_group
    FROM game_hitting
)
SELECT
    batter as player_id,
    MIN(date) as start_date,
    MAX(date) as end_date,
    COUNT(*) as length,
    MIN(gid) as start_game_id,
    MAX(gid) as end_game_id,
    SUM(at_bats) as total_ab,
    SUM(hits) as total_hits
FROM with_streak_id
WHERE had_hit = 1  -- Only count games with hits
GROUP BY batter, streak_group
HAVING COUNT(*) >= $3
ORDER BY length DESC, start_date
