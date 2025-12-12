-- Find scoreless innings streaks for a pitcher in a given season
-- Parameters: $1 = player_id (retroID), $2 = season year (YYYY format as string prefix), $3 = min innings pitched

WITH game_pitching AS (
    SELECT
        pitcher,
        gid,
        date,
        -- Count outs recorded (each out = 1/3 inning)
        ROUND(SUM(outs_post - outs_pre)::numeric / 3.0, 1) as innings_pitched,
        SUM(COALESCE(er, 0)) as earned_runs
    FROM plays
    WHERE pitcher = $1
      AND date >= $2 || '0101'
      AND date <= $2 || '1231'
      AND pitcher IS NOT NULL
      AND pitcher != ''
    GROUP BY pitcher, gid, date
    HAVING SUM(outs_post - outs_pre) >= 3  -- At least 1 inning pitched
),
with_streak_id AS (
    SELECT
        pitcher,
        gid,
        date,
        innings_pitched,
        earned_runs,
        CASE WHEN earned_runs = 0 THEN 1 ELSE 0 END as was_scoreless,
        -- Create streak groups: increment when earned_runs > 0
        SUM(CASE WHEN earned_runs > 0 THEN 1 ELSE 0 END) OVER (
            PARTITION BY pitcher
            ORDER BY date, gid
            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
        ) as streak_group
    FROM game_pitching
)
SELECT
    pitcher as player_id,
    MIN(date) as start_date,
    MAX(date) as end_date,
    COUNT(*) as games_in_streak,
    SUM(innings_pitched) as length,
    MIN(gid) as start_game_id,
    MAX(gid) as end_game_id
FROM with_streak_id
WHERE was_scoreless = 1
GROUP BY pitcher, streak_group
HAVING SUM(innings_pitched) >= $3
ORDER BY length DESC, start_date
