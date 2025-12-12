-- Add unique constraint to ejections table
-- A person can only be ejected once per game, so game_id + ejectee_id is unique

ALTER TABLE ejections ADD CONSTRAINT ejections_game_ejectee_unique UNIQUE (game_id, ejectee_id);
