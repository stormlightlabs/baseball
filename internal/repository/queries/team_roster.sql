-- Get roster for a team in a specific season with batting and pitching stats
-- Parameters: $1 = teamID, $2 = yearID
WITH batting AS (
	SELECT
		b."playerID",
		b."G" as batting_g,
		b."AB",
		b."H",
		b."HR",
		b."RBI",
		CASE WHEN b."AB" > 0 THEN CAST(b."H" AS float) / b."AB" ELSE NULL END as avg
	FROM "Batting" b
	WHERE b."teamID" = $1 AND b."yearID" = $2
),
pitching AS (
	SELECT
		p."playerID",
		p."G" as pitching_g,
		p."W",
		p."L",
		p."SO",
		CASE WHEN p."IPouts" > 0 THEN (CAST(p."ER" AS float) * 27) / p."IPouts" ELSE NULL END as era
	FROM "Pitching" p
	WHERE p."teamID" = $1 AND p."yearID" = $2
),
appearances AS (
	SELECT
		a."playerID",
		CASE
			WHEN a."G_c" > 0 THEN 'C'
			WHEN a."G_1b" > 0 THEN '1B'
			WHEN a."G_2b" > 0 THEN '2B'
			WHEN a."G_3b" > 0 THEN '3B'
			WHEN a."G_ss" > 0 THEN 'SS'
			WHEN a."G_lf" > 0 THEN 'LF'
			WHEN a."G_cf" > 0 THEN 'CF'
			WHEN a."G_rf" > 0 THEN 'RF'
			WHEN a."G_of" > 0 THEN 'OF'
			WHEN a."G_dh" > 0 THEN 'DH'
			WHEN a."G_p" > 0 THEN 'P'
			ELSE NULL
		END as primary_position
	FROM "Appearances" a
	WHERE a."teamID" = $1 AND a."yearID" = $2
)
SELECT
	p."playerID",
	p."nameFirst",
	p."nameLast",
	a.primary_position,
	b.batting_g,
	b."AB",
	b."H",
	b."HR",
	b."RBI",
	b.avg,
	pt.pitching_g,
	pt."W",
	pt."L",
	pt.era,
	pt."SO"
FROM "People" p
LEFT JOIN batting b ON p."playerID" = b."playerID"
LEFT JOIN pitching pt ON p."playerID" = pt."playerID"
LEFT JOIN appearances a ON p."playerID" = a."playerID"
WHERE b."playerID" IS NOT NULL OR pt."playerID" IS NOT NULL
ORDER BY
	CASE
		WHEN b.batting_g IS NOT NULL THEN b.batting_g
		WHEN pt.pitching_g IS NOT NULL THEN pt.pitching_g
		ELSE 0
	END DESC,
	p."nameLast", p."nameFirst"
