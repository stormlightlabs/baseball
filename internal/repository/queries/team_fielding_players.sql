SELECT
    "playerID", "yearID", "teamID", "lgID", "POS",
    SUM("G") as g, SUM("GS") as gs,
    SUM("InnOuts") as inn_outs,
    SUM("PO") as po, SUM("A") as a, SUM("E") as e, SUM("DP") as dp,
    SUM("PB") as pb,
    SUM(CASE WHEN "POS" = 'C' THEN 0 ELSE 0 END) as sb,
    SUM(CASE WHEN "POS" = 'C' THEN 0 ELSE 0 END) as cs
FROM "Fielding"
WHERE "teamID" = $1 AND "yearID" = $2
GROUP BY "playerID", "yearID", "teamID", "lgID", "POS"
ORDER BY SUM("InnOuts") DESC
