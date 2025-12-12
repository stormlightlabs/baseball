SELECT
	"teamID", "yearID", "lgID",
	SUM("G") as g,
	SUM("PO") as po, SUM("A") as a, SUM("E") as e, SUM("DP") as dp,
	SUM(CASE WHEN "POS" = 'C' THEN "PB" ELSE 0 END) as pb,
	SUM(CASE WHEN "POS" = 'C' THEN "WP" ELSE 0 END) as wp,
	SUM(CASE WHEN "POS" = 'C' THEN "SB" ELSE 0 END) as sb,
	SUM(CASE WHEN "POS" = 'C' THEN "CS" ELSE 0 END) as cs
FROM "Fielding"
WHERE 1=1
