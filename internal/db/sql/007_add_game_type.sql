-- Add game_type column to games table to distinguish regular season, postseason, and all-star games
ALTER TABLE games ADD COLUMN IF NOT EXISTS game_type varchar(20) DEFAULT 'regular';

-- Create index for filtering by game type
CREATE INDEX IF NOT EXISTS idx_games_game_type ON games(game_type);
