SELECT ump_id, ump_name
FROM (
    SELECT hp_ump_id AS ump_id, hp_ump_name AS ump_name FROM games WHERE hp_ump_id IS NOT NULL AND hp_ump_id != ''
    UNION
    SELECT b1_ump_id, b1_ump_name FROM games WHERE b1_ump_id IS NOT NULL AND b1_ump_id != ''
    UNION
    SELECT b2_ump_id, b2_ump_name FROM games WHERE b2_ump_id IS NOT NULL AND b2_ump_id != ''
    UNION
    SELECT b3_ump_id, b3_ump_name FROM games WHERE b3_ump_id IS NOT NULL AND b3_ump_id != ''
    UNION
    SELECT lf_ump_id, lf_ump_name FROM games WHERE lf_ump_id IS NOT NULL AND lf_ump_id != ''
    UNION
    SELECT rf_ump_id, rf_ump_name FROM games WHERE rf_ump_id IS NOT NULL AND rf_ump_id != ''
) AS all_umps
ORDER BY ump_name
