-- Persisted MLBAM crosswalk tables for players and teams.

CREATE TABLE IF NOT EXISTS player_mlbam_map (
    mlbam_id INTEGER PRIMARY KEY,
    lahman_id VARCHAR(9),
    retro_id VARCHAR(8),
    bbref_id VARCHAR(16),
    full_name TEXT,
    source VARCHAR(32) NOT NULL DEFAULT 'chadwick',
    confidence VARCHAR(16) NOT NULL DEFAULT 'high',
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_player_mlbam_map_lahman ON player_mlbam_map(lahman_id);
CREATE INDEX IF NOT EXISTS idx_player_mlbam_map_retro ON player_mlbam_map(retro_id);
CREATE INDEX IF NOT EXISTS idx_player_mlbam_map_bbref ON player_mlbam_map(bbref_id);

CREATE TABLE IF NOT EXISTS team_mlbam_map (
    season INTEGER NOT NULL,
    mlbam_team_id INTEGER NOT NULL,
    mlb_abbreviation VARCHAR(8),
    mlb_team_code VARCHAR(16),
    mlb_file_code VARCHAR(16),
    mlb_team_name TEXT,
    mlb_franchise_name TEXT,
    mlb_club_name TEXT,
    local_team_id VARCHAR(3),
    local_franchise_id VARCHAR(3),
    local_team_name TEXT,
    local_league VARCHAR(8),
    match_method VARCHAR(32),
    confidence VARCHAR(16) NOT NULL DEFAULT 'none',
    source VARCHAR(32) NOT NULL DEFAULT 'mlb-api',
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (season, mlbam_team_id)
);

CREATE INDEX IF NOT EXISTS idx_team_mlbam_map_local_team ON team_mlbam_map(local_team_id);
CREATE INDEX IF NOT EXISTS idx_team_mlbam_map_local_franchise ON team_mlbam_map(local_franchise_id);
CREATE INDEX IF NOT EXISTS idx_team_mlbam_map_mlbam ON team_mlbam_map(mlbam_team_id);

COMMENT ON TABLE player_mlbam_map IS
'Persisted MLBAM personId crosswalk to local Lahman/Retrosheet identifiers.';

COMMENT ON TABLE team_mlbam_map IS
'Season-scoped MLBAM teamId crosswalk to local team_id/franchise_id identifiers.';
