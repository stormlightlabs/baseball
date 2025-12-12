-- Get leverage index data for each plate appearance in a game
SELECT
    pn as event_id,
    inning,
    top_bot,
    score_h as home_score_before,
    score_v as away_score_before,
    outs_pre as outs_before,
    -- Base state encoding (000 = none, 100 = 1st only, etc.)
    CASE WHEN br1_pre IS NOT NULL AND br1_pre != '' THEN '1' ELSE '0' END ||
    CASE WHEN br2_pre IS NOT NULL AND br2_pre != '' THEN '1' ELSE '0' END ||
    CASE WHEN br3_pre IS NOT NULL AND br3_pre != '' THEN '1' ELSE '0' END as bases_before,
    batter,
    pitcher,
    LEFT(event, 50) as description
FROM plays
WHERE gid = $1
ORDER BY pn;
