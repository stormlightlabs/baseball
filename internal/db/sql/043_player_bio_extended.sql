-- Extended biographical data from Retrosheet biofile
-- Supplements Lahman People table with additional fields

CREATE TABLE IF NOT EXISTS player_bio_extended (
    retro_id VARCHAR(10) PRIMARY KEY,

    use_name VARCHAR(100),      -- Nickname or common name used
    full_name VARCHAR(255),     -- Full legal name
    birth_name VARCHAR(255),    -- Birth name (if different)
    alt_name VARCHAR(255),      -- Alternative name

    cemetery VARCHAR(100),
    cem_city VARCHAR(100),
    cem_state VARCHAR(50),
    cem_country VARCHAR(50),
    cem_note TEXT,

    -- Career dates for non-playing roles
    debut_coach DATE,
    last_coach DATE,
    debut_manager DATE,
    last_manager DATE,
    debut_umpire DATE,
    last_umpire DATE,

    hof_retrosheet VARCHAR(10),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bio_extended_retro_id ON player_bio_extended(retro_id);

COMMENT ON TABLE player_bio_extended IS 'Extended biographical data from Retrosheet biofile, supplementing Lahman People table';
COMMENT ON COLUMN player_bio_extended.retro_id IS 'Retrosheet player ID, matches People.retroID';
COMMENT ON COLUMN player_bio_extended.use_name IS 'Common name or nickname used by player';
COMMENT ON COLUMN player_bio_extended.full_name IS 'Full legal name';
COMMENT ON COLUMN player_bio_extended.birth_name IS 'Name at birth if different from current name';
COMMENT ON COLUMN player_bio_extended.cem_note IS 'Additional cemetery location details';
COMMENT ON COLUMN player_bio_extended.hof_retrosheet IS 'Hall of Fame status from Retrosheet (may differ from Lahman)';
