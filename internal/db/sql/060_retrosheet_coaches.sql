-- Create coaches table from Retrosheet biodata
-- Tracks coaching stints for players

CREATE TABLE IF NOT EXISTS coaches (
    retro_id VARCHAR(10) NOT NULL,
    year INTEGER NOT NULL,
    team_id VARCHAR(3) NOT NULL,
    role VARCHAR(50),          -- Coaching role/position
    first_game DATE,
    last_game DATE,
    PRIMARY KEY (retro_id, year, team_id)
);

-- Indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_coaches_retro_id ON coaches(retro_id);
CREATE INDEX IF NOT EXISTS idx_coaches_year ON coaches(year);
CREATE INDEX IF NOT EXISTS idx_coaches_team ON coaches(team_id);
CREATE INDEX IF NOT EXISTS idx_coaches_year_team ON coaches(year, team_id);

COMMENT ON TABLE coaches IS 'Coaching records from Retrosheet biodata';
COMMENT ON COLUMN coaches.retro_id IS 'Retrosheet player ID';
COMMENT ON COLUMN coaches.role IS 'Coaching position/role';
