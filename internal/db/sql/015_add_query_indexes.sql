-- Additional indexes to optimize Negro Leagues endpoints and pagination-heavy queries

-- Helps /negroleagues/plays when sorting by play number within games
CREATE INDEX IF NOT EXISTS idx_plays_gid_pn ON plays(gid, pn);

-- Reduce scans when filtering games by league + date (both home and visiting)
CREATE INDEX IF NOT EXISTS idx_games_home_league_date ON games(home_team_league, date);
CREATE INDEX IF NOT EXISTS idx_games_visiting_league_date ON games(visiting_team_league, date);

-- General date filter for schedule endpoints
CREATE INDEX IF NOT EXISTS idx_games_date ON games(date);
