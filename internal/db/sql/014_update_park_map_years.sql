-- Update FanGraphs team-park mappings to cover 2016-2025 period
-- Most teams have been in the same parks during this period

UPDATE fangraphs_team_park_map SET start_year = 2016 WHERE fangraphs_team IN (
    'Angels', 'Diamondbacks', 'Braves', 'Orioles', 'Red Sox', 'White Sox',
    'Cubs', 'Reds', 'Guardians', 'Rockies', 'Tigers', 'Astros', 'Royals',
    'Dodgers', 'Marlins', 'Brewers', 'Twins', 'Yankees', 'Mets', 'Phillies',
    'Pirates', 'Padres', 'Giants', 'Mariners', 'Cardinals', 'Rays', 'Blue Jays',
    'Nationals'
);

-- Rangers: Use current park for all years (slight inaccuracy for 2016-2019)
-- They moved to Globe Life Field (ARL02) in 2020, but we'll use it for all years
-- Previous park was ARL01 (2016-2019) - could be refined later
UPDATE fangraphs_team_park_map SET start_year = 2016 WHERE fangraphs_team = 'Rangers';

-- Athletics moved from Oakland
-- For 2016-2024, they were at Oakland Coliseum
UPDATE fangraphs_team_park_map SET start_year = 2016, end_year = 2024 WHERE fangraphs_team = 'Athletics';
