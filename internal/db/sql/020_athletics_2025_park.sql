-- Athletics: Remove end_year constraint to allow 2025 data to load
-- Note: This will use OAK01 for 2025 in the mapping, which is technically
-- incorrect (they moved to SAC01), but the park_factors table will still
-- have the FanGraphs data which is team-based not park-based

UPDATE fangraphs_team_park_map
SET end_year = NULL
WHERE fangraphs_team = 'Athletics';

-- TODO: Consider redesigning the mapping table to support year-specific parks
-- if we need more accuracy for historical park changes
