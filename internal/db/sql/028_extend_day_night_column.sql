-- Extend day_night column to accommodate Negro Leagues data
-- Negro Leagues data contains values like "day", "night" which exceed varchar(1)

ALTER TABLE games
ALTER COLUMN day_night TYPE varchar(10);
