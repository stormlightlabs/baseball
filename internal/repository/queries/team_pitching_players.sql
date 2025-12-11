SELECT
    "playerID", "yearID", "teamID", "lgID",
    SUM("W") as w, SUM("L") as l,
    SUM("G") as g, SUM("GS") as gs, SUM("CG") as cg, SUM("SHO") as sho, SUM("SV") as sv,
    SUM("IPouts") as ip_outs,
    SUM("H") as h, SUM("ER") as er, SUM("HR") as hr,
    SUM("BB") as bb, SUM("SO") as so,
    SUM("HBP") as hbp, SUM("BK") as bk, SUM("WP") as wp
FROM "Pitching"
WHERE "teamID" = $1 AND "yearID" = $2
GROUP BY "playerID", "yearID", "teamID", "lgID"
ORDER BY SUM("IPouts") DESC
