-- Adjust column sizes in ejections table to accommodate actual data
-- Game IDs can be longer than 8 characters (e.g., CL6188905250)

-- Drop the unique constraint temporarily
ALTER TABLE ejections DROP CONSTRAINT IF EXISTS ejections_game_ejectee_unique;

-- Increase column sizes
ALTER TABLE ejections ALTER COLUMN game_id TYPE varchar(20);
ALTER TABLE ejections ALTER COLUMN ejectee_id TYPE varchar(20);
ALTER TABLE ejections ALTER COLUMN umpire_id TYPE varchar(20);

-- Recreate the unique constraint
ALTER TABLE ejections ADD CONSTRAINT ejections_game_ejectee_unique UNIQUE (game_id, ejectee_id);
