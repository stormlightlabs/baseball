-- Create materialized view for park ID crosswalk
-- Maps park codes across Lahman and Retrosheet, handling missing mappings

DROP MATERIALIZED VIEW IF EXISTS park_map CASCADE;

CREATE MATERIALIZED VIEW park_map AS
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

-- Indexes for efficient lookups
CREATE UNIQUE INDEX idx_park_map_retro ON park_map(retro_park_id);
CREATE INDEX idx_park_map_lahman ON park_map(lahman_park_id) WHERE lahman_park_id IS NOT NULL;
CREATE INDEX idx_park_map_city ON park_map(city) WHERE city IS NOT NULL;
CREATE INDEX idx_park_map_era ON park_map(era);

-- Helper functions
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
