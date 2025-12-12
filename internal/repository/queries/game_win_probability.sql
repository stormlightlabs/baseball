-- Extract game state for win probability calculation
-- Parameters: $1 = game_id

SELECT
    pn as event_index,
    inning,
    CASE WHEN top_bot = 1 THEN true ELSE false END as top_of_inning,
    score_h as home_score,
    score_v as away_score,
    outs_post as outs,
    CONCAT(
        CASE WHEN br1_post IS NOT NULL AND br1_post != '' THEN '1' ELSE '0' END,
        CASE WHEN br2_post IS NOT NULL AND br2_post != '' THEN '1' ELSE '0' END,
        CASE WHEN br3_post IS NOT NULL AND br3_post != '' THEN '1' ELSE '0' END
    ) as bases,
    event as description
FROM plays
WHERE gid = $1
ORDER BY pn
