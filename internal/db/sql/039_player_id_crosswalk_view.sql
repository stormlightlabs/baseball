-- Create materialized view for player ID crosswalk
-- Normalizes Lahman ↔ Retrosheet player identifiers for seamless joins

DROP MATERIALIZED VIEW IF EXISTS player_id_map CASCADE;

CREATE MATERIALIZED VIEW player_id_map AS
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
CREATE INDEX idx_player_id_map_lahman ON player_id_map(lahman_id);
CREATE UNIQUE INDEX idx_player_id_map_retro ON player_id_map(retro_id);

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
