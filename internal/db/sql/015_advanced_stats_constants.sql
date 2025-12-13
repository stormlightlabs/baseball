-- Advanced statistics constants tables
-- Supports wOBA, wRC+, FIP, and WAR calculations

-- wOBA constants by year (from FanGraphs Guts data)
CREATE TABLE IF NOT EXISTS woba_constants (
    season INT PRIMARY KEY,

    -- wOBA weights
    w_bb DECIMAL(5,3) NOT NULL,      -- unintentional walk
    w_hbp DECIMAL(5,3) NOT NULL,     -- hit by pitch
    w_1b DECIMAL(5,3) NOT NULL,      -- single
    w_2b DECIMAL(5,3) NOT NULL,      -- double
    w_3b DECIMAL(5,3) NOT NULL,      -- triple
    w_hr DECIMAL(5,3) NOT NULL,      -- home run

    -- wOBA scale and conversion
    woba_scale DECIMAL(5,3) NOT NULL,
    woba DECIMAL(5,3) NOT NULL,       -- league average wOBA

    -- Base running weights
    run_sb DECIMAL(5,3) NOT NULL,     -- stolen base runs
    run_cs DECIMAL(5,3) NOT NULL,     -- caught stealing runs

    -- League context
    r_pa DECIMAL(5,3) NOT NULL,       -- runs per plate appearance
    r_w DECIMAL(5,2) NOT NULL,        -- runs per win

    -- FIP constant
    c_fip DECIMAL(5,3) NOT NULL,      -- FIP constant (normalizes to ERA scale)

    -- Metadata
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_woba_constants_season ON woba_constants(season);

COMMENT ON TABLE woba_constants IS 'Year-specific constants for wOBA, FIP, and related advanced metrics from FanGraphs';
COMMENT ON COLUMN woba_constants.woba_scale IS 'Scaling factor to convert wOBA to runs above average';
COMMENT ON COLUMN woba_constants.woba IS 'League average wOBA for the season';
COMMENT ON COLUMN woba_constants.r_w IS 'Runs per win (typically ~10, varies by run environment)';
COMMENT ON COLUMN woba_constants.c_fip IS 'FIP constant to normalize FIP to ERA scale';

-- League-specific constants by year and league
CREATE TABLE IF NOT EXISTS league_constants (
    season INT NOT NULL,
    league VARCHAR(2) NOT NULL CHECK (league IN ('AL', 'NL')),

    -- wOBA/wRC+ context
    woba_avg DECIMAL(5,3),           -- league average wOBA
    wrc_per_pa DECIMAL(6,4),         -- wRC per PA (excluding pitchers)

    -- WAR context
    runs_per_win DECIMAL(4,2),       -- league-specific runs per win
    replacement_runs_per_pa DECIMAL(8,6), -- replacement level runs per PA

    -- League totals for calculations
    total_pa BIGINT,                 -- total plate appearances
    total_runs BIGINT,               -- total runs scored

    -- Metadata
    calculated_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (season, league)
);

CREATE INDEX idx_league_constants_season ON league_constants(season);

COMMENT ON TABLE league_constants IS 'League-specific constants calculated annually for park/league adjustments';
COMMENT ON COLUMN league_constants.wrc_per_pa IS 'League wRC/PA excluding pitchers for wRC+ denominator';
COMMENT ON COLUMN league_constants.replacement_runs_per_pa IS 'Replacement level value for WAR calculations';

-- Positional adjustment constants (relatively static, from FanGraphs)
CREATE TABLE IF NOT EXISTS positional_adjustment_constants (
    position VARCHAR(10) PRIMARY KEY,
    runs_per_162 DECIMAL(4,1) NOT NULL,

    CONSTRAINT check_valid_position CHECK (
        position IN ('C', '1B', '2B', '3B', 'SS', 'LF', 'CF', 'RF', 'DH', 'P')
    )
);

COMMENT ON TABLE positional_adjustment_constants IS 'Fixed positional adjustments for WAR (runs per 162 defensive games)';

-- Insert standard FanGraphs positional adjustments
INSERT INTO positional_adjustment_constants (position, runs_per_162) VALUES
    ('C', 12.5),
    ('1B', -12.5),
    ('2B', 2.5),
    ('3B', 2.5),
    ('SS', 7.5),
    ('LF', -7.5),
    ('CF', 2.5),
    ('RF', -7.5),
    ('DH', -17.5),
    ('P', 0.0)
ON CONFLICT (position) DO NOTHING;

-- Park factors (from FanGraphs multi-year regressed data)
CREATE TABLE IF NOT EXISTS park_factors (
    park_id VARCHAR(10) NOT NULL,
    season INT NOT NULL,
    team_id VARCHAR(3),              -- Team playing at this park

    -- Overall park factors (100 = neutral)
    basic_5yr INT,                   -- 5-year regressed (most stable)
    basic_3yr INT,                   -- 3-year regressed
    basic_1yr INT,                   -- Single year (most volatile)

    -- Component park factors
    factor_1b INT,                   -- Singles
    factor_2b INT,                   -- Doubles
    factor_3b INT,                   -- Triples
    factor_hr INT,                   -- Home runs
    factor_so INT,                   -- Strikeouts
    factor_bb INT,                   -- Walks
    factor_gb INT,                   -- Ground balls
    factor_fb INT,                   -- Fly balls
    factor_ld INT,                   -- Line drives
    factor_iffb INT,                 -- Infield fly balls
    factor_fip INT,                  -- FIP

    -- Metadata
    imported_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (park_id, season)
);

CREATE INDEX idx_park_factors_season ON park_factors(season);
CREATE INDEX idx_park_factors_team ON park_factors(team_id, season);

COMMENT ON TABLE park_factors IS 'Multi-year regressed park factors from FanGraphs (100 = neutral, >100 = hitter-friendly)';
COMMENT ON COLUMN park_factors.basic_5yr IS 'Most stable 5-year regressed overall park factor';
COMMENT ON COLUMN park_factors.factor_hr IS 'Home run park factor (e.g., Coors Field ~105-110)';
