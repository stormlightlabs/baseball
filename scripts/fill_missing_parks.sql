-- Script to fill missing park metadata gaps in the Parks table
-- Focus on high-usage Negro Leagues parks and modern parks lacking metadata
--
-- 2025-12-15

BEGIN;

-- Newark parks (NWK04 - Ruppert Stadium)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'NWK04', 'Ruppert Stadium', 'Davids'' Stadium', 'Newark', 'NJ', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'NWK04');

-- Dayton (DAY03 - typically neutral site games)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'DAY03', 'Unknown Park', 'Dayton neutral site', 'Dayton', 'OH', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'DAY03');

-- Little Rock (LRK02)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'LRK02', 'Travelers Field', 'Ray Winder Field', 'Little Rock', 'AR', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'LRK02');

-- Memphis (MEM03 - Martin Stadium/Russwood Park)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'MEM03', 'Martin Stadium', 'Russwood Park', 'Memphis', 'TN', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'MEM03');

-- Houston (HOU04 - Buff Stadium)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'HOU04', 'Buff Stadium', '', 'Houston', 'TX', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'HOU04');

-- Tampa (TAM02 - George M. Steinbrenner Field, temporary Rays home 2025)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'TAM02', 'George M. Steinbrenner Field', 'Steinbrenner Field', 'Tampa', 'FL', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'TAM02');

-- Sacramento (SAC01 - Sutter Health Park, temporary A's home 2025)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'SAC01', 'Sutter Health Park', '', 'West Sacramento', 'CA', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'SAC01');

-- Los Angeles (LOS05 - Wrigley Field LA, Negro Leagues)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'LOS05', 'Wrigley Field', 'Los Angeles Wrigley Field', 'Los Angeles', 'CA', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'LOS05');

-- Wilmington (WIL04)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'WIL04', 'Unknown Park', '', 'Wilmington', 'DE', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'WIL04');

-- Buffalo (BUF06 - Offermann Stadium)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'BUF06', 'Offermann Stadium', 'Bison Stadium', 'Buffalo', 'NY', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'BUF06');

-- Nashville (NSH02 - Sulphur Dell/Tom Wilson Park)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'NSH02', 'Tom Wilson Park', 'Sulphur Dell', 'Nashville', 'TN', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'NSH02');

-- Harrisburg (HRB01)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'HRB01', 'Unknown Park', '', 'Harrisburg', 'PA', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'HRB01');

-- Jacksonville (JKV01)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'JKV01', 'Unknown Park', '', 'Jacksonville', 'FL', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'JKV01');

-- Trenton (TRE01)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'TRE01', 'Unknown Park', '', 'Trenton', 'NJ', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'TRE01');

-- Norfolk (NOR02)
INSERT INTO "Parks" (parkkey, parkname, parkalias, city, state, country)
SELECT 'NOR02', 'Unknown Park', '', 'Norfolk', 'VA', 'US'
WHERE NOT EXISTS (SELECT 1 FROM "Parks" WHERE parkkey = 'NOR02');

COMMIT;

-- Refresh the park_map materialized view
REFRESH MATERIALIZED VIEW CONCURRENTLY park_map;

-- Verification query - check coverage improvement
SELECT
    COUNT(*) as total_parks,
    SUM(CASE WHEN in_lahman THEN 1 ELSE 0 END) as with_metadata,
    ROUND(100.0 * SUM(CASE WHEN in_lahman THEN 1 ELSE 0 END) / COUNT(*), 1) as coverage_pct
FROM park_map;

-- Show remaining high-usage parks without metadata (after updates)
SELECT
    retro_park_id,
    home_teams,
    game_count,
    era
FROM parks_missing_from_lahman
WHERE game_count > 20
ORDER BY game_count DESC
LIMIT 20;
