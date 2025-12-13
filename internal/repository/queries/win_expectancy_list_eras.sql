-- List all available historical eras in the win expectancy table
-- Returns distinct era ranges with metadata

SELECT
    start_year,
    end_year,
    CASE
        WHEN start_year IS NULL AND end_year IS NULL THEN 'All Time'
        WHEN start_year = end_year THEN start_year::text || ' Season'
        ELSE start_year::text || '-' || end_year::text || ' Era'
    END as label,
    COUNT(*) as state_count,
    SUM(sample_size) as total_sample
FROM win_expectancy_historical
GROUP BY start_year, end_year
ORDER BY start_year DESC NULLS LAST, end_year DESC NULLS LAST
