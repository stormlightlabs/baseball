-- Add canonical game_id column matching Retrosheet play IDs (home+date+game number)
ALTER TABLE games
	ADD COLUMN IF NOT EXISTS game_id varchar(16)
	GENERATED ALWAYS AS (home_team || date || COALESCE(game_number::text, '0')) STORED;

CREATE UNIQUE INDEX IF NOT EXISTS idx_games_game_id ON games(game_id);
