-- Create indexes for player_game_pitching_stats materialized view
-- Optimizes common query patterns: pitcher lookups, game lookups, date ranges, season filters

-- Primary lookup indexes
CREATE INDEX idx_player_game_pitching_player_id ON player_game_pitching_stats(player_id);
CREATE INDEX idx_player_game_pitching_game_id ON player_game_pitching_stats(game_id);
CREATE INDEX idx_player_game_pitching_date ON player_game_pitching_stats(date);
CREATE INDEX idx_player_game_pitching_season ON player_game_pitching_stats(season);
CREATE INDEX idx_player_game_pitching_team ON player_game_pitching_stats(team_id);

-- Composite indexes for common query patterns
CREATE INDEX idx_player_game_pitching_player_season ON player_game_pitching_stats(player_id, season);
CREATE INDEX idx_player_game_pitching_player_date ON player_game_pitching_stats(player_id, date);
CREATE INDEX idx_player_game_pitching_team_date ON player_game_pitching_stats(team_id, date);
CREATE INDEX idx_player_game_pitching_season_date ON player_game_pitching_stats(season, date);

-- Game finder indexes (for queries like "games with 10+ strikeouts")
CREATE INDEX idx_player_game_pitching_so ON player_game_pitching_stats(so) WHERE so >= 10;
CREATE INDEX idx_player_game_pitching_era ON player_game_pitching_stats(era) WHERE era > 0;
CREATE INDEX idx_player_game_pitching_ip ON player_game_pitching_stats(ip) WHERE ip >= 5;

COMMENT ON INDEX idx_player_game_pitching_player_id IS 'Fast lookup of all games for a specific pitcher';
COMMENT ON INDEX idx_player_game_pitching_player_season IS 'Fast lookup of pitcher season game logs';
COMMENT ON INDEX idx_player_game_pitching_so IS 'Game finder: games with 10+ strikeouts';
