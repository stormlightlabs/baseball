-- Increase line score column length to accommodate extra-inning games
-- Some historical games (especially from the Federal League era) have longer line scores
ALTER TABLE games ALTER COLUMN visiting_line_score TYPE varchar(50);
ALTER TABLE games ALTER COLUMN home_line_score TYPE varchar(50);
