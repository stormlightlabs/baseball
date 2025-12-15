-- Fix day_night column in negro_leagues_games table
-- The day_night field contains values like "day", "night", etc. which exceed varchar(1)

ALTER TABLE negro_leagues_games
ALTER COLUMN day_night TYPE varchar(10);
