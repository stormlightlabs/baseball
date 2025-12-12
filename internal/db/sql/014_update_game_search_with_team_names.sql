-- Update search text generation to include team names from aliases
CREATE OR REPLACE FUNCTION update_game_search_trigger()
RETURNS TRIGGER AS $$
DECLARE
    home_name TEXT;
    away_name TEXT;
BEGIN
    -- Get primary team names from aliases
    SELECT alias INTO home_name
    FROM team_aliases
    WHERE team_id = NEW.home_team
      AND (start_year IS NULL OR start_year <= EXTRACT(YEAR FROM TO_DATE(NEW.date, 'YYYYMMDD'))::int)
      AND (end_year IS NULL OR end_year >= EXTRACT(YEAR FROM TO_DATE(NEW.date, 'YYYYMMDD'))::int)
    ORDER BY LENGTH(alias)
    LIMIT 1;

    SELECT alias INTO away_name
    FROM team_aliases
    WHERE team_id = NEW.visiting_team
      AND (start_year IS NULL OR start_year <= EXTRACT(YEAR FROM TO_DATE(NEW.date, 'YYYYMMDD'))::int)
      AND (end_year IS NULL OR end_year >= EXTRACT(YEAR FROM TO_DATE(NEW.date, 'YYYYMMDD'))::int)
    ORDER BY LENGTH(alias)
    LIMIT 1;

    NEW.search_text := (
        NEW.date || ' ' ||
        EXTRACT(YEAR FROM TO_DATE(NEW.date, 'YYYYMMDD'))::text || ' ' ||
        NEW.home_team || ' ' ||
        NEW.visiting_team || ' ' ||
        COALESCE(home_name, '') || ' ' ||
        COALESCE(away_name, '') || ' ' ||
        COALESCE(NEW.game_type, '') || ' ' ||
        CASE
            WHEN NEW.game_type IN ('worldseries', 'lcs', 'divisionseries', 'wildcard') THEN 'playoffs postseason world series'
            WHEN NEW.game_type = 'allstar' THEN 'all-star allstar midsummer classic'
            ELSE 'regular season'
        END || ' ' ||
        COALESCE(NEW.park_id, '') || ' ' ||
        TO_CHAR(TO_DATE(NEW.date, 'YYYYMMDD'), 'Month DD YYYY') || ' ' ||
        NEW.day_of_week
    );

    NEW.search_tsv := to_tsvector('english', NEW.search_text);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Repopulate search_text for all existing games
UPDATE games g
SET search_text = (
    SELECT
        g.date || ' ' ||
        EXTRACT(YEAR FROM TO_DATE(g.date, 'YYYYMMDD'))::text || ' ' ||
        g.home_team || ' ' ||
        g.visiting_team || ' ' ||
        COALESCE(
            (SELECT alias FROM team_aliases
             WHERE team_id = g.home_team
               AND (start_year IS NULL OR start_year <= EXTRACT(YEAR FROM TO_DATE(g.date, 'YYYYMMDD'))::int)
               AND (end_year IS NULL OR end_year >= EXTRACT(YEAR FROM TO_DATE(g.date, 'YYYYMMDD'))::int)
             ORDER BY LENGTH(alias) LIMIT 1),
            ''
        ) || ' ' ||
        COALESCE(
            (SELECT alias FROM team_aliases
             WHERE team_id = g.visiting_team
               AND (start_year IS NULL OR start_year <= EXTRACT(YEAR FROM TO_DATE(g.date, 'YYYYMMDD'))::int)
               AND (end_year IS NULL OR end_year >= EXTRACT(YEAR FROM TO_DATE(g.date, 'YYYYMMDD'))::int)
             ORDER BY LENGTH(alias) LIMIT 1),
            ''
        ) || ' ' ||
        COALESCE(g.game_type, '') || ' ' ||
        CASE
            WHEN g.game_type IN ('worldseries', 'lcs', 'divisionseries', 'wildcard') THEN 'playoffs postseason world series'
            WHEN g.game_type = 'allstar' THEN 'all-star allstar midsummer classic'
            ELSE 'regular season'
        END || ' ' ||
        COALESCE(g.park_id, '') || ' ' ||
        TO_CHAR(TO_DATE(g.date, 'YYYYMMDD'), 'Month DD YYYY') || ' ' ||
        g.day_of_week
);

-- Update tsvector for full-text search
UPDATE games
SET search_tsv = to_tsvector('english', search_text)
WHERE search_text IS NOT NULL;
