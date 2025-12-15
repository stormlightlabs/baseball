-- Add "Cleveland" as transition name (used by FanGraphs in 2021)
-- Between "Indians" and "Guardians"

INSERT INTO fangraphs_team_park_map (fangraphs_team, retrosheet_team_id, primary_park_id, start_year, end_year, notes)
VALUES ('Cleveland', 'CLE', 'CLE08', 2021, 2021, 'Cleveland (transition year before Guardians)')
ON CONFLICT (fangraphs_team) DO NOTHING;
