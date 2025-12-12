SELECT
	"yearID",
	"round",
	"teamIDwinner",
	"lgIDwinner",
	"teamIDloser",
	"lgIDloser",
	"wins",
	"losses",
	"ties"
FROM "SeriesPost"
WHERE "yearID" = $1
ORDER BY
	CASE "round"
		WHEN 'WS' THEN 1
		WHEN 'ALCS' THEN 2
		WHEN 'NLCS' THEN 2
		WHEN 'ALDS1' THEN 3
		WHEN 'ALDS2' THEN 3
		WHEN 'NLDS1' THEN 3
		WHEN 'NLDS2' THEN 3
		WHEN 'AEDIV' THEN 3
		WHEN 'NEDIV' THEN 3
		WHEN 'ALEWC' THEN 4
		WHEN 'NLWC' THEN 4
		ELSE 5
	END
