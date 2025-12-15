-- Create indexes for player_game_batting_stats materialized view
-- Optimizes common query patterns: player lookups, game lookups, date ranges, season filters

-- Primary lookup indexes
CREATE INDEX idx_player_game_batting_player_id ON player_game_batting_stats(player_id);
CREATE INDEX idx_player_game_batting_game_id ON player_game_batting_stats(game_id);
CREATE INDEX idx_player_game_batting_date ON player_game_batting_stats(date);
CREATE INDEX idx_player_game_batting_season ON player_game_batting_stats(season);
CREATE INDEX idx_player_game_batting_team ON player_game_batting_stats(team_id);

-- Composite indexes for common query patterns
CREATE INDEX idx_player_game_batting_player_season ON player_game_batting_stats(player_id, season);
CREATE INDEX idx_player_game_batting_player_date ON player_game_batting_stats(player_id, date);
CREATE INDEX idx_player_game_batting_team_date ON player_game_batting_stats(team_id, date);
CREATE INDEX idx_player_game_batting_season_date ON player_game_batting_stats(season, date);

-- Game finder indexes (for queries like "games with 2+ HR")
CREATE INDEX idx_player_game_batting_hr ON player_game_batting_stats(hr) WHERE hr > 0;
CREATE INDEX idx_player_game_batting_h ON player_game_batting_stats(h) WHERE h >= 3;
CREATE INDEX idx_player_game_batting_rbi ON player_game_batting_stats(rbi) WHERE rbi > 0;
CREATE INDEX idx_player_game_batting_sb ON player_game_batting_stats(sb) WHERE sb > 0;

COMMENT ON INDEX idx_player_game_batting_player_id IS 'Fast lookup of all games for a specific player';
COMMENT ON INDEX idx_player_game_batting_player_season IS 'Fast lookup of player season game logs';
COMMENT ON INDEX idx_player_game_batting_hr IS 'Game finder: games with home runs';
