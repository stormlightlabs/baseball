-- Create indexes for team_game_stats materialized view
-- These indexes support common query patterns for team daily stats API

-- Primary lookup: team + date range queries
CREATE INDEX idx_team_game_stats_team_date
ON team_game_stats(team_id, date DESC);

-- Season-based queries
CREATE INDEX idx_team_game_stats_team_season
ON team_game_stats(team_id, season DESC);

-- Game lookup for joining with games table
CREATE INDEX idx_team_game_stats_game_id
ON team_game_stats(game_id);

-- Date-based queries for league-wide daily stats
CREATE INDEX idx_team_game_stats_date
ON team_game_stats(date DESC);

-- Season queries for leaderboards
CREATE INDEX idx_team_game_stats_season
ON team_game_stats(season DESC);

-- Support concurrent refresh
CREATE UNIQUE INDEX idx_team_game_stats_unique
ON team_game_stats(game_id, team_id);

COMMENT ON INDEX idx_team_game_stats_team_date IS
'Primary index for team daily stats queries by team and date range';

COMMENT ON INDEX idx_team_game_stats_team_season IS
'Supports season-based team performance queries';

COMMENT ON INDEX idx_team_game_stats_unique IS
'Unique constraint to support REFRESH MATERIALIZED VIEW CONCURRENTLY';
