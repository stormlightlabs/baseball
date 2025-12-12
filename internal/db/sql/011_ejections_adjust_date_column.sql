-- Adjust date column size in ejections table
-- Dates are in MM/DD/YYYY format which is 10 characters

ALTER TABLE ejections ALTER COLUMN date TYPE varchar(10);
