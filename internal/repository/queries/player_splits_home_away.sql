-- Player batting splits by home/away
-- Parameters: $1 = player_id (retroID), $2 = season year (YYYY format as string prefix)

WITH home_away_splits AS (
    SELECT
        batter,
        CASE WHEN vis_home = 0 THEN 'away' ELSE 'home' END as split_key,
        CASE WHEN vis_home = 0 THEN 'Away' ELSE 'Home' END as split_label,
        COUNT(DISTINCT gid) as games,
        COUNT(*) as pa,
        SUM(ab) as ab,
        SUM(single + double + triple + hr) as h,
        SUM(double) as doubles,
        SUM(triple) as triples,
        SUM(hr) as hr,
        SUM(walk) as bb,
        SUM(k) as so,
        SUM(sf) as sf,
        SUM(hbp) as hbp
    FROM plays
    WHERE batter = $1
      AND date >= $2 || '0101'
      AND date <= $2 || '1231'
      AND batter IS NOT NULL
    GROUP BY batter, vis_home
)
SELECT
    split_key,
    split_label,
    NULL as meta_value,
    games,
    pa,
    ab,
    h,
    hr,
    bb,
    so,
    CASE WHEN ab > 0 THEN h::numeric / ab::numeric ELSE 0 END as avg,
    CASE WHEN (ab + bb + hbp + sf) > 0
         THEN (h + bb + hbp)::numeric / (ab + bb + hbp + sf)::numeric
         ELSE 0
    END as obp,
    CASE WHEN ab > 0
         THEN (h + doubles + (2 * triples) + (3 * hr))::numeric / ab::numeric
         ELSE 0
    END as slg
FROM home_away_splits
ORDER BY split_key
