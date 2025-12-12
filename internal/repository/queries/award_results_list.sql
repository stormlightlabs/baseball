SELECT
	ap."awardID", ap."playerID", ap."yearID", ap."lgID",
	asp."votesFirst", asp."pointsWon"
FROM "AwardsPlayers" ap
LEFT JOIN "AwardsSharePlayers" asp
	ON ap."awardID" = asp."awardID"
	AND ap."playerID" = asp."playerID"
	AND ap."yearID" = asp."yearID"
	AND (ap."lgID" = asp."lgID" OR (ap."lgID" IS NULL AND asp."lgID" IS NULL))
WHERE 1=1
