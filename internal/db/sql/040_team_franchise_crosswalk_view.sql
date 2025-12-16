-- Create materialized view for team/franchise ID crosswalk
-- Maps Retrosheet team codes to Lahman team IDs and franchise IDs across seasons

DROP MATERIALIZED VIEW IF EXISTS team_franchise_map CASCADE;

CREATE MATERIALIZED VIEW team_franchise_map AS
SELECT DISTINCT
    t."teamID" as team_id,
    t."franchID" as franchise_id,
    t."yearID" as season,
    t.name as team_name,
    t."lgID" as league,
    tf."franchName" as franchise_name,
    -- First and last years for this team code
    MIN(t."yearID") OVER (PARTITION BY t."teamID") as first_season,
    MAX(t."yearID") OVER (PARTITION BY t."teamID") as last_season,
    -- Team is currently active
    CASE
        WHEN MAX(t."yearID") OVER (PARTITION BY t."teamID") >=
             (SELECT MAX("yearID") FROM "Teams") - 1
        THEN true
        ELSE false
    END as is_active
FROM "Teams" t
LEFT JOIN "TeamsFranchises" tf ON t."franchID" = tf."franchID"
WHERE t."teamID" IS NOT NULL;

CREATE INDEX idx_team_franchise_map_team ON team_franchise_map(team_id);
CREATE INDEX idx_team_franchise_map_franchise ON team_franchise_map(franchise_id);
CREATE INDEX idx_team_franchise_map_season ON team_franchise_map(season);
CREATE UNIQUE INDEX idx_team_franchise_map_unique ON team_franchise_map(team_id, season);

CREATE OR REPLACE FUNCTION franchise_current_team(franchise VARCHAR)
RETURNS VARCHAR AS $$
    SELECT team_id
    FROM team_franchise_map
    WHERE franchise_id = franchise AND is_active
    ORDER BY season DESC
    LIMIT 1;
$$ LANGUAGE SQL STABLE;

CREATE OR REPLACE FUNCTION franchise_all_teams(franchise VARCHAR)
RETURNS TABLE(team_id VARCHAR, season INT, team_name VARCHAR) AS $$
    SELECT DISTINCT team_id, season, team_name
    FROM team_franchise_map
    WHERE franchise_id = franchise
    ORDER BY season DESC;
$$ LANGUAGE SQL STABLE;

COMMENT ON MATERIALIZED VIEW team_franchise_map IS
'Team and franchise ID crosswalk for mapping team codes across seasons.
Links Retrosheet team codes (e.g., "NYA") to Lahman franchise IDs (e.g., "NYY").
Includes temporal information for handling relocations and name changes.
Refresh after loading new team data: REFRESH MATERIALIZED VIEW CONCURRENTLY team_franchise_map;';

COMMENT ON FUNCTION franchise_current_team IS
'Get the current/latest team ID for a franchise (e.g., "ATL" for "ATL" franchise)';

COMMENT ON FUNCTION franchise_all_teams IS
'Get all historical team IDs for a franchise, ordered by season (e.g., all teams for Braves franchise)';
