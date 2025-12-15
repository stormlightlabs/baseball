-- Create indexes for player_game_fielding_stats materialized view
-- Optimizes common query patterns: player lookups, game lookups, date ranges, season filters, position filters

-- Primary lookup indexes
CREATE INDEX idx_player_game_fielding_player_id ON player_game_fielding_stats(player_id);
CREATE INDEX idx_player_game_fielding_game_id ON player_game_fielding_stats(game_id);
CREATE INDEX idx_player_game_fielding_date ON player_game_fielding_stats(date);
CREATE INDEX idx_player_game_fielding_season ON player_game_fielding_stats(season);
CREATE INDEX idx_player_game_fielding_team ON player_game_fielding_stats(team_id);
CREATE INDEX idx_player_game_fielding_position ON player_game_fielding_stats(position);

-- Composite indexes for common query patterns
CREATE INDEX idx_player_game_fielding_player_season ON player_game_fielding_stats(player_id, season);
CREATE INDEX idx_player_game_fielding_player_position ON player_game_fielding_stats(player_id, position);
CREATE INDEX idx_player_game_fielding_player_date ON player_game_fielding_stats(player_id, date);
CREATE INDEX idx_player_game_fielding_team_date ON player_game_fielding_stats(team_id, date);
CREATE INDEX idx_player_game_fielding_season_position ON player_game_fielding_stats(season, position);

-- Game finder indexes (for queries like "games with 3+ errors")
CREATE INDEX idx_player_game_fielding_errors ON player_game_fielding_stats(e) WHERE e > 0;
CREATE INDEX idx_player_game_fielding_assists ON player_game_fielding_stats(a) WHERE a >= 5;
CREATE INDEX idx_player_game_fielding_putouts ON player_game_fielding_stats(po) WHERE po >= 10;

COMMENT ON INDEX idx_player_game_fielding_player_id IS 'Fast lookup of all games for a specific player';
COMMENT ON INDEX idx_player_game_fielding_player_position IS 'Fast lookup of player games at specific position';
COMMENT ON INDEX idx_player_game_fielding_errors IS 'Game finder: games with errors';
