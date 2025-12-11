SELECT
    "yearID",
    string_agg(DISTINCT "lgID", ',' ORDER BY "lgID") as leagues,
    COUNT(*) as team_count
FROM "Teams"
GROUP BY "yearID"
ORDER BY "yearID" DESC
