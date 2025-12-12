-- Player batting splits by pitcher handedness (vs LHP/RHP)
-- Parameters: $1 = player_id (retroID), $2 = season year (YYYY format as string prefix)

WITH pitcher_hand_splits AS (
    SELECT
        batter,
        CASE
            WHEN pithand = 'L' THEN 'vs_lhp'
            WHEN pithand = 'R' THEN 'vs_rhp'
            ELSE 'unknown'
        END as split_key,
        CASE
            WHEN pithand = 'L' THEN 'vs LHP'
            WHEN pithand = 'R' THEN 'vs RHP'
            ELSE 'Unknown'
        END as split_label,
        pithand,
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
      AND pithand IS NOT NULL
    GROUP BY batter, pithand
)
SELECT
    split_key,
    split_label,
    pithand as meta_handedness,
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
FROM pitcher_hand_splits
WHERE split_key != 'unknown'
ORDER BY split_key
