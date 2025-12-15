-- FanGraphs team name to Retrosheet park ID mapping
-- Maps FanGraphs full team names to their primary home park_id from Retrosheet

CREATE TABLE IF NOT EXISTS fangraphs_team_park_map (
    fangraphs_team VARCHAR(50) PRIMARY KEY,
    retrosheet_team_id VARCHAR(3) NOT NULL,
    primary_park_id VARCHAR(5) NOT NULL,
    start_year INT,
    end_year INT,
    notes TEXT
);

COMMENT ON TABLE fangraphs_team_park_map IS 'Maps FanGraphs team names to Retrosheet park IDs for park factors';

-- Insert mappings for current teams (2023-2025 era)
INSERT INTO fangraphs_team_park_map (fangraphs_team, retrosheet_team_id, primary_park_id, start_year, notes) VALUES
    ('Angels', 'ANA', 'ANA01', 2023, 'Angel Stadium'),
    ('Diamondbacks', 'ARI', 'PHO01', 2023, 'Chase Field'),
    ('Braves', 'ATL', 'ATL03', 2023, 'Truist Park'),
    ('Orioles', 'BAL', 'BAL12', 2023, 'Camden Yards'),
    ('Red Sox', 'BOS', 'BOS07', 2023, 'Fenway Park'),
    ('White Sox', 'CHA', 'CHI12', 2023, 'Guaranteed Rate Field'),
    ('Cubs', 'CHN', 'CHI11', 2023, 'Wrigley Field'),
    ('Reds', 'CIN', 'CIN09', 2023, 'Great American Ball Park'),
    ('Guardians', 'CLE', 'CLE08', 2023, 'Progressive Field'),
    ('Rockies', 'COL', 'DEN02', 2023, 'Coors Field'),
    ('Tigers', 'DET', 'DET05', 2023, 'Comerica Park'),
    ('Astros', 'HOU', 'HOU03', 2023, 'Minute Maid Park'),
    ('Royals', 'KCA', 'KAN06', 2023, 'Kauffman Stadium'),
    ('Dodgers', 'LAN', 'LOS03', 2023, 'Dodger Stadium'),
    ('Marlins', 'MIA', 'MIA02', 2023, 'loanDepot park'),
    ('Brewers', 'MIL', 'MIL06', 2023, 'American Family Field'),
    ('Twins', 'MIN', 'MIN04', 2023, 'Target Field'),
    ('Yankees', 'NYA', 'NYC21', 2023, 'Yankee Stadium'),
    ('Mets', 'NYN', 'NYC19', 2023, 'Citi Field'),
    ('Athletics', 'OAK', 'OAK01', 2023, 'Oakland Coliseum'),
    ('Phillies', 'PHI', 'PHI13', 2023, 'Citizens Bank Park'),
    ('Pirates', 'PIT', 'PIT08', 2023, 'PNC Park'),
    ('Padres', 'SDN', 'SAN02', 2023, 'Petco Park'),
    ('Giants', 'SFN', 'SFO03', 2023, 'Oracle Park'),
    ('Mariners', 'SEA', 'SEA03', 2023, 'T-Mobile Park'),
    ('Cardinals', 'SLN', 'STL10', 2023, 'Busch Stadium'),
    ('Rays', 'TBA', 'STP01', 2023, 'Tropicana Field'),
    ('Rangers', 'TEX', 'ARL02', 2023, 'Globe Life Field'),
    ('Blue Jays', 'TOR', 'TOR02', 2023, 'Rogers Centre'),
    ('Nationals', 'WAS', 'WAS11', 2023, 'Nationals Park')
ON CONFLICT (fangraphs_team) DO NOTHING;

-- Create index for reverse lookup
CREATE INDEX idx_fangraphs_team_park_retrosheet ON fangraphs_team_park_map(retrosheet_team_id, start_year);
