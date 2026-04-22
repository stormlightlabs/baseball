-- Current-season bridge tables for in-season MLB Stats API ingestion.

CREATE SCHEMA IF NOT EXISTS current_season;

CREATE TABLE IF NOT EXISTS current_season.batting (
    mlb_id INTEGER NOT NULL,
    player_id VARCHAR,
    season INTEGER NOT NULL,
    team_mlb_id INTEGER NOT NULL,
    team_id VARCHAR,
    g INTEGER,
    pa INTEGER,
    ab INTEGER,
    r INTEGER,
    h INTEGER,
    "2b" INTEGER,
    "3b" INTEGER,
    hr INTEGER,
    rbi INTEGER,
    sb INTEGER,
    cs INTEGER,
    bb INTEGER,
    so INTEGER,
    hbp INTEGER,
    sf INTEGER,
    sh INTEGER,
    avg NUMERIC(4, 3),
    obp NUMERIC(4, 3),
    slg NUMERIC(4, 3),
    ops NUMERIC(4, 3),
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (mlb_id, season, team_mlb_id)
);

CREATE TABLE IF NOT EXISTS current_season.pitching (
    mlb_id INTEGER NOT NULL,
    player_id VARCHAR,
    season INTEGER NOT NULL,
    team_mlb_id INTEGER NOT NULL,
    team_id VARCHAR,
    w INTEGER,
    l INTEGER,
    g INTEGER,
    gs INTEGER,
    sv INTEGER,
    ip NUMERIC(5, 1),
    h INTEGER,
    r INTEGER,
    er INTEGER,
    hr INTEGER,
    bb INTEGER,
    so INTEGER,
    hbp INTEGER,
    era NUMERIC(4, 2),
    whip NUMERIC(4, 2),
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (mlb_id, season, team_mlb_id)
);

CREATE TABLE IF NOT EXISTS current_season.standings (
    season INTEGER NOT NULL,
    division_id INTEGER NOT NULL,
    division_name VARCHAR NOT NULL,
    team_mlb_id INTEGER NOT NULL,
    team_id VARCHAR,
    franchise_id VARCHAR,
    w INTEGER,
    l INTEGER,
    pct NUMERIC(4, 3),
    gb VARCHAR,
    wc_gb VARCHAR,
    streak VARCHAR,
    l10 VARCHAR,
    run_diff INTEGER,
    rs INTEGER,
    ra INTEGER,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (season, team_mlb_id)
);

CREATE TABLE IF NOT EXISTS current_season.games (
    game_pk INTEGER PRIMARY KEY,
    season INTEGER NOT NULL,
    game_date DATE NOT NULL,
    status VARCHAR NOT NULL,
    away_mlb_id INTEGER NOT NULL,
    away_team_id VARCHAR,
    away_score INTEGER,
    home_mlb_id INTEGER NOT NULL,
    home_team_id VARCHAR,
    home_score INTEGER,
    venue VARCHAR,
    innings INTEGER,
    day_night VARCHAR,
    doubleheader VARCHAR,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cs_batting_season ON current_season.batting(season);
CREATE INDEX IF NOT EXISTS idx_cs_pitching_season ON current_season.pitching(season);
CREATE INDEX IF NOT EXISTS idx_cs_standings_season ON current_season.standings(season);
CREATE INDEX IF NOT EXISTS idx_cs_games_date ON current_season.games(game_date);
CREATE INDEX IF NOT EXISTS idx_cs_games_season ON current_season.games(season);

ALTER TABLE etl_jobs DROP CONSTRAINT IF EXISTS chk_etl_jobs_type;
ALTER TABLE etl_jobs
ADD CONSTRAINT chk_etl_jobs_type
CHECK (
    job_type IN (
        'full-run',
        'yearly-sync',
        'validate-only',
        'cleanup-only',
        'maintenance',
        'current-season-sync'
    )
);
