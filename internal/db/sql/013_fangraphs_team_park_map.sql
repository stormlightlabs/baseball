-- FanGraphs team name to Retrosheet park ID mapping.
-- Includes stable 2016+ mappings plus historical aliases used by FanGraphs.

CREATE TABLE IF NOT EXISTS fangraphs_team_park_map (
    fangraphs_team VARCHAR(50) PRIMARY KEY,
    retrosheet_team_id VARCHAR(3) NOT NULL,
    primary_park_id VARCHAR(5) NOT NULL,
    start_year INT,
    end_year INT,
    notes TEXT
);

COMMENT ON TABLE fangraphs_team_park_map IS 'Maps FanGraphs team names to Retrosheet park IDs for park factors';

INSERT INTO fangraphs_team_park_map (
    fangraphs_team,
    retrosheet_team_id,
    primary_park_id,
    start_year,
    end_year,
    notes
)
VALUES
    ('Angels', 'ANA', 'ANA01', 2016, NULL, 'Angel Stadium'),
    ('Diamondbacks', 'ARI', 'PHO01', 2016, NULL, 'Chase Field'),
    ('Braves', 'ATL', 'ATL03', 2016, NULL, 'Truist Park'),
    ('Orioles', 'BAL', 'BAL12', 2016, NULL, 'Camden Yards'),
    ('Red Sox', 'BOS', 'BOS07', 2016, NULL, 'Fenway Park'),
    ('White Sox', 'CHA', 'CHI12', 2016, NULL, 'Guaranteed Rate Field'),
    ('Cubs', 'CHN', 'CHI11', 2016, NULL, 'Wrigley Field'),
    ('Reds', 'CIN', 'CIN09', 2016, NULL, 'Great American Ball Park'),
    ('Guardians', 'CLE', 'CLE08', 2022, NULL, 'Progressive Field'),
    ('Indians', 'CLE', 'CLE08', 2016, 2021, 'Cleveland Indians (renamed to Guardians in 2022)'),
    ('Cleveland', 'CLE', 'CLE08', 2021, 2021, 'Cleveland (transition year before Guardians)'),
    ('Rockies', 'COL', 'DEN02', 2016, NULL, 'Coors Field'),
    ('Tigers', 'DET', 'DET05', 2016, NULL, 'Comerica Park'),
    ('Astros', 'HOU', 'HOU03', 2016, NULL, 'Minute Maid Park'),
    ('Royals', 'KCA', 'KAN06', 2016, NULL, 'Kauffman Stadium'),
    ('Dodgers', 'LAN', 'LOS03', 2016, NULL, 'Dodger Stadium'),
    ('Marlins', 'MIA', 'MIA02', 2016, NULL, 'loanDepot park'),
    ('Brewers', 'MIL', 'MIL06', 2016, NULL, 'American Family Field'),
    ('Twins', 'MIN', 'MIN04', 2016, NULL, 'Target Field'),
    ('Yankees', 'NYA', 'NYC21', 2016, NULL, 'Yankee Stadium'),
    ('Mets', 'NYN', 'NYC19', 2016, NULL, 'Citi Field'),
    ('Athletics', 'OAK', 'OAK01', 2016, NULL, 'Oakland Coliseum (used for mapping continuity)'),
    ('Phillies', 'PHI', 'PHI13', 2016, NULL, 'Citizens Bank Park'),
    ('Pirates', 'PIT', 'PIT08', 2016, NULL, 'PNC Park'),
    ('Padres', 'SDN', 'SAN02', 2016, NULL, 'Petco Park'),
    ('Giants', 'SFN', 'SFO03', 2016, NULL, 'Oracle Park'),
    ('Mariners', 'SEA', 'SEA03', 2016, NULL, 'T-Mobile Park'),
    ('Cardinals', 'SLN', 'STL10', 2016, NULL, 'Busch Stadium'),
    ('Rays', 'TBA', 'STP01', 2016, NULL, 'Tropicana Field'),
    ('Rangers', 'TEX', 'ARL02', 2016, NULL, 'Globe Life Field'),
    ('Blue Jays', 'TOR', 'TOR02', 2016, NULL, 'Rogers Centre'),
    ('Nationals', 'WAS', 'WAS11', 2016, NULL, 'Nationals Park')
ON CONFLICT (fangraphs_team) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_fangraphs_team_park_retrosheet ON fangraphs_team_park_map(retrosheet_team_id, start_year);
