-- crosswalk and win expectancy materialized views.


-- SECTION 024_player_id_crosswalk_view.sql


-- Create materialized view for player ID crosswalk
-- Normalizes Lahman ↔ Retrosheet player identifiers for seamless joins

DROP MATERIALIZED VIEW IF EXISTS player_id_map CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS player_id_map AS
WITH ranked_players AS (
    SELECT
        "playerID" as lahman_id,
        "retroID" as retro_id,
        "nameFirst" as first_name,
        "nameLast" as last_name,
        "nameGiven" as given_name,
        "debut" as debut_date,
        "finalGame" as final_game_date,
        "birthYear" as birth_year,
        "birthCountry" as birth_country,
        -- Prefer players with debut dates, then lexicographically first playerID
        ROW_NUMBER() OVER (
            PARTITION BY "retroID"
            ORDER BY
                CASE WHEN "debut" IS NOT NULL THEN 0 ELSE 1 END,
                "playerID"
        ) as rn
    FROM "People"
    WHERE "retroID" IS NOT NULL AND "retroID" <> ''
)
SELECT
    lahman_id,
    retro_id,
    first_name,
    last_name,
    given_name,
    debut_date,
    final_game_date,
    birth_year,
    birth_country
FROM ranked_players
WHERE rn = 1;

-- Add indexes for fast bidirectional lookups
CREATE INDEX IF NOT EXISTS idx_player_id_map_lahman ON player_id_map(lahman_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_player_id_map_retro ON player_id_map(retro_id);

-- Create lookup functions for convenience
CREATE OR REPLACE FUNCTION lahman_to_retro(lahman_player_id VARCHAR)
RETURNS VARCHAR AS $$
    SELECT retro_id FROM player_id_map WHERE lahman_id = lahman_player_id;
$$ LANGUAGE SQL IMMUTABLE;

CREATE OR REPLACE FUNCTION retro_to_lahman(retro_player_id VARCHAR)
RETURNS VARCHAR AS $$
    SELECT lahman_id FROM player_id_map WHERE retro_id = retro_player_id;
$$ LANGUAGE SQL IMMUTABLE;

COMMENT ON MATERIALIZED VIEW player_id_map IS
'Player ID crosswalk between Lahman and Retrosheet systems.
Enables seamless joins between Lahman career stats and Retrosheet play-by-play data.
Coverage: ~88% of all players have both IDs.
Refresh after loading new player data: REFRESH MATERIALIZED VIEW CONCURRENTLY player_id_map;';

COMMENT ON FUNCTION lahman_to_retro IS
'Convert Lahman player ID (e.g., "troutmi01") to Retrosheet ID (e.g., "trout001")';

COMMENT ON FUNCTION retro_to_lahman IS
'Convert Retrosheet player ID (e.g., "trout001") to Lahman ID (e.g., "troutmi01")';




-- SECTION 025_team_franchise_crosswalk_view.sql


-- Create materialized view for team/franchise ID crosswalk
-- Maps Retrosheet team codes to Lahman team IDs and franchise IDs across seasons

DROP MATERIALIZED VIEW IF EXISTS team_franchise_map CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS team_franchise_map AS
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

CREATE INDEX IF NOT EXISTS idx_team_franchise_map_team ON team_franchise_map(team_id);
CREATE INDEX IF NOT EXISTS idx_team_franchise_map_franchise ON team_franchise_map(franchise_id);
CREATE INDEX IF NOT EXISTS idx_team_franchise_map_season ON team_franchise_map(season);
CREATE UNIQUE INDEX IF NOT EXISTS idx_team_franchise_map_unique ON team_franchise_map(team_id, season);

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




-- SECTION 026_park_crosswalk_view.sql


-- Create materialized view for park ID crosswalk
-- Maps park codes across Lahman and Retrosheet, handling missing mappings

DROP MATERIALIZED VIEW IF EXISTS park_map CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS park_map AS
WITH retrosheet_parks AS (
    -- Get all parks from games with game counts
    SELECT
        g.park_id,
        MIN(g.date) as first_game_date,
        MAX(g.date) as last_game_date,
        COUNT(*) as game_count,
        -- Infer city/region from team codes
        STRING_AGG(DISTINCT g.home_team, ', ' ORDER BY g.home_team) as home_teams
    FROM games g
    WHERE g.park_id IS NOT NULL
    GROUP BY g.park_id
),
deduplicated_parks AS (
    -- Deduplicate Parks table (some parks like NSH01 have duplicate rows)
    SELECT DISTINCT ON (parkkey)
        parkkey,
        parkname,
        parkalias,
        city,
        state,
        country
    FROM "Parks"
    ORDER BY parkkey, "ID"
)
SELECT
    rp.park_id as retro_park_id,
    p.parkkey as lahman_park_id,
    COALESCE(p.parkname, 'Unknown Park') as park_name,
    COALESCE(p.parkalias, '') as park_alias,
    p.city,
    p.state,
    p.country,
    rp.first_game_date,
    rp.last_game_date,
    rp.game_count,
    rp.home_teams,
    -- Flag for whether this park exists in Lahman
    CASE WHEN p.parkkey IS NOT NULL THEN true ELSE false END as in_lahman,
    -- Determine park era
    CASE
        WHEN rp.first_game_date::INT >= 20000000 THEN 'modern'
        WHEN rp.first_game_date::INT >= 19600000 THEN 'expansion'
        WHEN rp.first_game_date::INT >= 19200000 THEN 'golden_age'
        ELSE 'deadball'
    END as era
FROM retrosheet_parks rp
LEFT JOIN deduplicated_parks p ON rp.park_id = p.parkkey;

CREATE UNIQUE INDEX IF NOT EXISTS idx_park_map_retro ON park_map(retro_park_id);
CREATE INDEX IF NOT EXISTS idx_park_map_lahman ON park_map(lahman_park_id) WHERE lahman_park_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_park_map_city ON park_map(city) WHERE city IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_park_map_era ON park_map(era);

CREATE OR REPLACE FUNCTION get_park_info(park_code VARCHAR)
RETURNS TABLE(
    park_name VARCHAR,
    city VARCHAR,
    state VARCHAR,
    games INT
) AS $$
    SELECT park_name, city, state, game_count::INT
    FROM park_map
    WHERE retro_park_id = park_code;
$$ LANGUAGE SQL STABLE;

CREATE OR REPLACE FUNCTION active_parks(since_year INT DEFAULT 2000)
RETURNS TABLE(
    park_id VARCHAR,
    park_name VARCHAR,
    city VARCHAR,
    games INT
) AS $$
    SELECT retro_park_id, park_name, city, game_count::INT
    FROM park_map
    WHERE last_game_date::INT >= (since_year * 10000)
    ORDER BY game_count DESC;
$$ LANGUAGE SQL STABLE;

-- View for parks missing from Lahman (need manual enrichment)
CREATE OR REPLACE VIEW parks_missing_from_lahman AS
SELECT
    retro_park_id,
    home_teams,
    first_game_date,
    last_game_date,
    game_count,
    era
FROM park_map
WHERE NOT in_lahman
ORDER BY game_count DESC;

COMMENT ON MATERIALIZED VIEW park_map IS
'Comprehensive park ID crosswalk between Retrosheet and Lahman systems.
Includes all parks from Retrosheet games, with Lahman metadata where available.
Coverage: 127/449 parks (28%) have Lahman metadata; rest are Retrosheet-only (mostly Negro Leagues).
Refresh after loading new game data: REFRESH MATERIALIZED VIEW CONCURRENTLY park_map;';

COMMENT ON VIEW parks_missing_from_lahman IS
'Parks that appear in Retrosheet games but lack Lahman metadata.
These parks (mostly Negro Leagues venues) need manual enrichment for full details.';

COMMENT ON FUNCTION get_park_info IS
'Get park information by Retrosheet park code (e.g., "NYC16" for Yankee Stadium)';

COMMENT ON FUNCTION active_parks IS
'Get all parks used since specified year, ordered by number of games played';




-- SECTION 035_win_expectancy_add_columns.sql


-- Create/rebuild the win_expectancy_historical materialized view.
-- Includes synthetic id and timestamps for API contract compatibility.

DROP TABLE IF EXISTS win_expectancy_historical CASCADE;
DROP MATERIALIZED VIEW IF EXISTS win_expectancy_historical CASCADE;

CREATE MATERIALIZED VIEW IF NOT EXISTS win_expectancy_historical AS
WITH game_outcomes AS (
    SELECT
        date,
        home_team,
        game_number,
        CASE
            WHEN home_score > visiting_score THEN true
            WHEN home_score < visiting_score THEN false
            ELSE NULL
        END as home_won
    FROM games
    WHERE home_score IS NOT NULL
      AND visiting_score IS NOT NULL
      AND home_score != visiting_score
),
game_states AS (
    SELECT
        LEAST(p.inning, 9) as inning,
        (p.top_bot = 1)::boolean as is_bottom,
        p.outs_pre as outs,
        CONCAT(
            CASE WHEN p.br1_pre IS NOT NULL AND p.br1_pre != '' THEN '1' ELSE '_' END,
            CASE WHEN p.br2_pre IS NOT NULL AND p.br2_pre != '' THEN '2' ELSE '_' END,
            CASE WHEN p.br3_pre IS NOT NULL AND p.br3_pre != '' THEN '3' ELSE '_' END
        ) as runners_state,
        LEAST(GREATEST(p.score_h - p.score_v, -11), 11) as score_diff,
        go.home_won,
        SUBSTRING(p.date, 1, 4)::int as year
    FROM plays p
    INNER JOIN game_outcomes go ON
        SUBSTRING(p.gid, 4, 8) = go.date AND
        LEFT(p.gid, 3) = go.home_team AND
        RIGHT(p.gid, 1)::int = go.game_number
    WHERE p.outs_pre IS NOT NULL
      AND p.inning IS NOT NULL
      AND go.home_won IS NOT NULL
),
win_rates AS (
    SELECT
        inning,
        is_bottom,
        outs,
        runners_state,
        score_diff,
        MIN(year) as start_year,
        MAX(year) as end_year,
        AVG(CASE WHEN home_won THEN 1.0 ELSE 0.0 END) as win_probability,
        COUNT(*) as sample_size
    FROM game_states
    GROUP BY inning, is_bottom, outs, runners_state, score_diff
    HAVING COUNT(*) >= 100
)
SELECT
    ROW_NUMBER() OVER (ORDER BY inning, is_bottom, outs, runners_state, score_diff, start_year, end_year)::int as id,
    inning,
    is_bottom,
    outs,
    runners_state,
    score_diff,
    win_probability,
    sample_size,
    start_year,
    end_year,
    NOW() as created_at,
    NOW() as updated_at
FROM win_rates;

CREATE UNIQUE INDEX IF NOT EXISTS idx_win_expectancy_state ON win_expectancy_historical(
    inning, is_bottom, outs, runners_state, score_diff, start_year, end_year
);
CREATE INDEX IF NOT EXISTS idx_win_expectancy_inning ON win_expectancy_historical(inning);
CREATE INDEX IF NOT EXISTS idx_win_expectancy_outs ON win_expectancy_historical(outs);
CREATE INDEX IF NOT EXISTS idx_win_expectancy_runners ON win_expectancy_historical(runners_state);
CREATE INDEX IF NOT EXISTS idx_win_expectancy_score ON win_expectancy_historical(score_diff);


