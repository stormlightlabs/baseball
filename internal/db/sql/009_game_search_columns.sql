-- Add search columns to games table for natural language game search
ALTER TABLE games ADD COLUMN IF NOT EXISTS search_text text;
ALTER TABLE games ADD COLUMN IF NOT EXISTS search_tsv tsvector;

-- Enable trigram extension for fuzzy matching
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create indexes for full-text and fuzzy search
CREATE INDEX IF NOT EXISTS games_search_tsv_idx ON games USING GIN (search_tsv);
CREATE INDEX IF NOT EXISTS games_search_trgm_idx ON games USING GIN (search_text gin_trgm_ops);

-- Function to populate search_text from game data
-- Combines game ID, season, team names, and series information
CREATE OR REPLACE FUNCTION update_game_search_text()
RETURNS void AS $$
BEGIN
    UPDATE games g
    SET search_text = (
        -- Construct searchable text from game metadata
        SELECT COALESCE(
            g.date || ' ' ||
            EXTRACT(YEAR FROM TO_DATE(g.date, 'YYYYMMDD'))::text || ' ' ||
            g.home_team || ' ' ||
            g.visiting_team || ' ' ||
            COALESCE(g.game_type, '') || ' ' ||
            CASE
                WHEN g.game_type = 'postseason' THEN 'playoffs postseason'
                WHEN g.game_type = 'allstar' THEN 'all-star allstar midsummer classic'
                ELSE 'regular season'
            END || ' ' ||
            g.park_id || ' ' ||
            TO_CHAR(TO_DATE(g.date, 'YYYYMMDD'), 'Month DD YYYY') || ' ' ||
            g.day_of_week,
            ''
        )
    );

    -- Update tsvector for full-text search
    UPDATE games
    SET search_tsv = to_tsvector('english', search_text)
    WHERE search_text IS NOT NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger to keep search columns updated on INSERT/UPDATE
CREATE OR REPLACE FUNCTION update_game_search_trigger()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_text := (
        NEW.date || ' ' ||
        EXTRACT(YEAR FROM TO_DATE(NEW.date, 'YYYYMMDD'))::text || ' ' ||
        NEW.home_team || ' ' ||
        NEW.visiting_team || ' ' ||
        COALESCE(NEW.game_type, '') || ' ' ||
        CASE
            WHEN NEW.game_type = 'postseason' THEN 'playoffs postseason'
            WHEN NEW.game_type = 'allstar' THEN 'all-star allstar midsummer classic'
            ELSE 'regular season'
        END || ' ' ||
        NEW.park_id || ' ' ||
        TO_CHAR(TO_DATE(NEW.date, 'YYYYMMDD'), 'Month DD YYYY') || ' ' ||
        NEW.day_of_week
    );

    NEW.search_tsv := to_tsvector('english', NEW.search_text);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS games_search_update ON games;
CREATE TRIGGER games_search_update
    BEFORE INSERT OR UPDATE ON games
    FOR EACH ROW
    EXECUTE FUNCTION update_game_search_trigger();
