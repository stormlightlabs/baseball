-- Team aliases table for natural language search
-- Maps common team names and variations to official team IDs
CREATE TABLE IF NOT EXISTS team_aliases (
    alias varchar(100) PRIMARY KEY,
    team_id varchar(3) NOT NULL,
    start_year int,
    end_year int
);

-- Create index for lookups
CREATE INDEX IF NOT EXISTS idx_team_aliases_team_id ON team_aliases(team_id);
CREATE INDEX IF NOT EXISTS idx_team_aliases_lower ON team_aliases(LOWER(alias));

-- Seed common team aliases
-- American League teams
INSERT INTO team_aliases (alias, team_id, start_year, end_year) VALUES
    ('yankees', 'NYA', 1903, NULL),
    ('new york yankees', 'NYA', 1903, NULL),
    ('ny yankees', 'NYA', 1903, NULL),
    ('red sox', 'BOS', 1908, NULL),
    ('boston red sox', 'BOS', 1908, NULL),
    ('orioles', 'BAL', 1954, NULL),
    ('baltimore orioles', 'BAL', 1954, NULL),
    ('rays', 'TBA', 1998, NULL),
    ('tampa bay rays', 'TBA', 1998, NULL),
    ('devil rays', 'TBA', 1998, 2007),
    ('tampa bay devil rays', 'TBA', 1998, 2007),
    ('blue jays', 'TOR', 1977, NULL),
    ('toronto blue jays', 'TOR', 1977, NULL),
    ('white sox', 'CHA', 1901, NULL),
    ('chicago white sox', 'CHA', 1901, NULL),
    ('guardians', 'CLE', 2022, NULL),
    ('cleveland guardians', 'CLE', 2022, NULL),
    ('indians', 'CLE', 1915, 2021),
    ('cleveland indians', 'CLE', 1915, 2021),
    ('tigers', 'DET', 1901, NULL),
    ('detroit tigers', 'DET', 1901, NULL),
    ('royals', 'KCA', 1969, NULL),
    ('kansas city royals', 'KCA', 1969, NULL),
    ('twins', 'MIN', 1961, NULL),
    ('minnesota twins', 'MIN', 1961, NULL),
    ('astros', 'HOU', 1962, NULL),
    ('houston astros', 'HOU', 1962, NULL),
    ('angels', 'ANA', 1961, NULL),
    ('los angeles angels', 'ANA', 1961, NULL),
    ('la angels', 'ANA', 1961, NULL),
    ('anaheim angels', 'ANA', 1997, 2004),
    ('athletics', 'OAK', 1968, NULL),
    ('oakland athletics', 'OAK', 1968, NULL),
    ('as', 'OAK', 1968, NULL),
    ('oakland as', 'OAK', 1968, NULL),
    ('mariners', 'SEA', 1977, NULL),
    ('seattle mariners', 'SEA', 1977, NULL),
    ('rangers', 'TEX', 1972, NULL),
    ('texas rangers', 'TEX', 1972, NULL),

-- National League teams
    ('braves', 'ATL', 1966, NULL),
    ('atlanta braves', 'ATL', 1966, NULL),
    ('marlins', 'FLO', 1993, NULL),
    ('florida marlins', 'FLO', 1993, 2011),
    ('miami marlins', 'FLO', 2012, NULL),
    ('mets', 'NYN', 1962, NULL),
    ('new york mets', 'NYN', 1962, NULL),
    ('ny mets', 'NYN', 1962, NULL),
    ('phillies', 'PHI', 1883, NULL),
    ('philadelphia phillies', 'PHI', 1883, NULL),
    ('nationals', 'WAS', 2005, NULL),
    ('washington nationals', 'WAS', 2005, NULL),
    ('cubs', 'CHN', 1876, NULL),
    ('chicago cubs', 'CHN', 1876, NULL),
    ('reds', 'CIN', 1890, NULL),
    ('cincinnati reds', 'CIN', 1890, NULL),
    ('brewers', 'MIL', 1970, NULL),
    ('milwaukee brewers', 'MIL', 1970, NULL),
    ('pirates', 'PIT', 1887, NULL),
    ('pittsburgh pirates', 'PIT', 1887, NULL),
    ('cardinals', 'SLN', 1892, NULL),
    ('st louis cardinals', 'SLN', 1892, NULL),
    ('stl cardinals', 'SLN', 1892, NULL),
    ('diamondbacks', 'ARI', 1998, NULL),
    ('arizona diamondbacks', 'ARI', 1998, NULL),
    ('dbacks', 'ARI', 1998, NULL),
    ('rockies', 'COL', 1993, NULL),
    ('colorado rockies', 'COL', 1993, NULL),
    ('dodgers', 'LAN', 1958, NULL),
    ('los angeles dodgers', 'LAN', 1958, NULL),
    ('la dodgers', 'LAN', 1958, NULL),
    ('brooklyn dodgers', 'BRO', 1884, 1957),
    ('padres', 'SDN', 1969, NULL),
    ('san diego padres', 'SDN', 1969, NULL),
    ('giants', 'SFN', 1958, NULL),
    ('san francisco giants', 'SFN', 1958, NULL),
    ('sf giants', 'SFN', 1958, NULL),
    ('new york giants', 'NYG', 1883, 1957)
ON CONFLICT (alias) DO NOTHING;

-- Series ID aliases for postseason search
CREATE TABLE IF NOT EXISTS series_aliases (
    alias varchar(50) PRIMARY KEY,
    series_id varchar(10) NOT NULL
);

INSERT INTO series_aliases (alias, series_id) VALUES
    ('world series', 'WS'),
    ('ws', 'WS'),
    ('alcs', 'ALCS'),
    ('nlcs', 'NLCS'),
    ('alds', 'ALDS'),
    ('nlds', 'NLDS'),
    ('al championship', 'ALCS'),
    ('nl championship', 'NLCS'),
    ('al division series', 'ALDS'),
    ('nl division series', 'NLDS'),
    ('wildcard', 'WC'),
    ('wild card', 'WC')
ON CONFLICT (alias) DO NOTHING;
