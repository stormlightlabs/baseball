SELECT DISTINCT
    hp_ump_id,
    hp_ump_name
FROM games
WHERE hp_ump_id = $1
UNION
SELECT DISTINCT
    b1_ump_id,
    b1_ump_name
FROM games
WHERE b1_ump_id = $1
UNION
SELECT DISTINCT
    b2_ump_id,
    b2_ump_name
FROM games
WHERE b2_ump_id = $1
UNION
SELECT DISTINCT
    b3_ump_id,
    b3_ump_name
FROM games
WHERE b3_ump_id = $1
UNION
SELECT DISTINCT
    lf_ump_id,
    lf_ump_name
FROM games
WHERE lf_ump_id = $1
UNION
SELECT DISTINCT
    rf_ump_id,
    rf_ump_name
FROM games
WHERE rf_ump_id = $1
LIMIT 1
