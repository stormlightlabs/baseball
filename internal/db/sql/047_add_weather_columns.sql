-- Add weather and game condition columns to games table
ALTER TABLE games ADD COLUMN IF NOT EXISTS temp_f INTEGER;
ALTER TABLE games ADD COLUMN IF NOT EXISTS sky VARCHAR(20);
ALTER TABLE games ADD COLUMN IF NOT EXISTS wind_direction VARCHAR(20);
ALTER TABLE games ADD COLUMN IF NOT EXISTS wind_speed_mph INTEGER;
ALTER TABLE games ADD COLUMN IF NOT EXISTS precip VARCHAR(20);
ALTER TABLE games ADD COLUMN IF NOT EXISTS field_condition VARCHAR(20);
ALTER TABLE games ADD COLUMN IF NOT EXISTS start_time TIME;
ALTER TABLE games ADD COLUMN IF NOT EXISTS used_dh BOOLEAN;

-- Create index for DH usage queries
CREATE INDEX IF NOT EXISTS idx_games_used_dh ON games(used_dh) WHERE used_dh IS NOT NULL;
