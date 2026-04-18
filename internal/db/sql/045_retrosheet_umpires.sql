-- Create umpires table from Retrosheet biodata
-- Tracks umpire biographical and career data

CREATE TABLE IF NOT EXISTS umpires (
    retro_id VARCHAR(10) PRIMARY KEY,
    last_name VARCHAR(100) NOT NULL,
    first_name VARCHAR(100),
    first_game DATE,
    last_game DATE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_umpires_last_name ON umpires(last_name);
CREATE INDEX IF NOT EXISTS idx_umpires_first_last ON umpires(first_name, last_name);

COMMENT ON TABLE umpires IS 'Umpire biographical data from Retrosheet';
COMMENT ON COLUMN umpires.retro_id IS 'Retrosheet umpire ID';
COMMENT ON COLUMN umpires.first_game IS 'Date of first game umpired';
COMMENT ON COLUMN umpires.last_game IS 'Date of last game umpired';
