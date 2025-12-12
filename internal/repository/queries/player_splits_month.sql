-- Player batting splits by calendar month
-- Parameters: $1 = player_id (retroID), $2 = season year (YYYY format as string prefix)

WITH monthly_splits AS (
    SELECT
        batter,
        SUBSTRING(date FROM 5 FOR 2) as month_num,
        CASE SUBSTRING(date FROM 5 FOR 2)
            WHEN '03' THEN 'March'
            WHEN '04' THEN 'April'
            WHEN '05' THEN 'May'
            WHEN '06' THEN 'June'
            WHEN '07' THEN 'July'
            WHEN '08' THEN 'August'
            WHEN '09' THEN 'September'
            WHEN '10' THEN 'October'
            WHEN '11' THEN 'November'
            ELSE 'Unknown'
        END as month_name,
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
    GROUP BY batter, SUBSTRING(date FROM 5 FOR 2)
)
SELECT
    month_num as split_key,
    month_name as split_label,
    month_num as meta_month,
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
FROM monthly_splits
ORDER BY month_num
