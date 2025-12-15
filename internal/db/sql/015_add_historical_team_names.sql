-- Add historical team names that FanGraphs used before rebranding

-- Cleveland: Indians (old) -> Guardians (new, 2022+)
INSERT INTO fangraphs_team_park_map (fangraphs_team, retrosheet_team_id, primary_park_id, start_year, end_year, notes)
VALUES ('Indians', 'CLE', 'CLE08', 2016, 2021, 'Cleveland Indians (renamed to Guardians in 2022)')
ON CONFLICT (fangraphs_team) DO NOTHING;

-- Update Guardians to start from 2022
UPDATE fangraphs_team_park_map
SET start_year = 2022
WHERE fangraphs_team = 'Guardians';
