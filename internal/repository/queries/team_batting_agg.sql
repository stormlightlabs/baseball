SELECT
    "teamID", "yearID", "lgID",
    COUNT(DISTINCT "playerID") as g,
    SUM("AB") as ab,
    SUM("R") as r,
    SUM("H") as h,
    SUM("2B") as doubles,
    SUM("3B") as triples,
    SUM("HR") as hr,
    SUM("RBI") as rbi,
    SUM("SB") as sb,
    SUM("CS") as cs,
    SUM("BB") as bb,
    SUM("SO") as so,
    SUM("HBP") as hbp,
    SUM("SF") as sf
FROM "Batting"
WHERE "teamID" = $1 AND "yearID" = $2
GROUP BY "teamID", "yearID", "lgID"
