SELECT
    "teamID", "yearID", "lgID",
    SUM("G") as g,
    SUM("PO") as po, SUM("A") as a, SUM("E") as e, SUM("DP") as dp,
    SUM(CASE WHEN "POS" = 'C' THEN "PB" ELSE 0 END) as pb,
    SUM(CASE WHEN "POS" = 'C' THEN "WP" ELSE 0 END) as wp
FROM "Fielding"
WHERE "teamID" = $1 AND "yearID" = $2
GROUP BY "teamID", "yearID", "lgID"
