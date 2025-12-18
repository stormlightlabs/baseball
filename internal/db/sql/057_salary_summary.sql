-- Create table for yearly salary summary data
-- This complements the Salaries table with aggregate statistics

DROP TABLE IF EXISTS salary_summary;

CREATE TABLE salary_summary (
    year INTEGER PRIMARY KEY,
    total NUMERIC(15, 2) NOT NULL,
    average NUMERIC(12, 2) NOT NULL,
    median NUMERIC(12, 2) NOT NULL
);

CREATE INDEX idx_salary_summary_year ON salary_summary(year);

COMMENT ON TABLE salary_summary IS 'Yearly salary summary statistics (total, average, median)';
COMMENT ON COLUMN salary_summary.year IS 'Season year';
COMMENT ON COLUMN salary_summary.total IS 'Total salary for all players in the season';
COMMENT ON COLUMN salary_summary.average IS 'Average salary for the season';
COMMENT ON COLUMN salary_summary.median IS 'Median salary for the season';
