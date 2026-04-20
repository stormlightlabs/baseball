-- reference data, enrichment tables, and crosswalk tables.

-- Team aliases table for natural language search
-- Maps common team names and variations to official team IDs
CREATE TABLE IF NOT EXISTS team_aliases (
    alias varchar(100) PRIMARY KEY,
    team_id varchar(3) NOT NULL,
    start_year int,
    end_year int
);

-- Create index for lookups
CREATE INDEX IF NOT EXISTS idx_team_aliases_team_id ON team_aliases(team_id);
CREATE INDEX IF NOT EXISTS idx_team_aliases_lower ON team_aliases(LOWER(alias));

-- Seed common team aliases
-- American League teams
INSERT INTO team_aliases (alias, team_id, start_year, end_year) VALUES
    ('yankees', 'NYA', 1903, NULL),
    ('new york yankees', 'NYA', 1903, NULL),
    ('ny yankees', 'NYA', 1903, NULL),
    ('red sox', 'BOS', 1908, NULL),
    ('boston red sox', 'BOS', 1908, NULL),
    ('orioles', 'BAL', 1954, NULL),
    ('baltimore orioles', 'BAL', 1954, NULL),
    ('rays', 'TBA', 1998, NULL),
    ('tampa bay rays', 'TBA', 1998, NULL),
    ('devil rays', 'TBA', 1998, 2007),
    ('tampa bay devil rays', 'TBA', 1998, 2007),
    ('blue jays', 'TOR', 1977, NULL),
    ('toronto blue jays', 'TOR', 1977, NULL),
    ('white sox', 'CHA', 1901, NULL),
    ('chicago white sox', 'CHA', 1901, NULL),
    ('guardians', 'CLE', 2022, NULL),
    ('cleveland guardians', 'CLE', 2022, NULL),
    ('indians', 'CLE', 1915, 2021),
    ('cleveland indians', 'CLE', 1915, 2021),
    ('tigers', 'DET', 1901, NULL),
    ('detroit tigers', 'DET', 1901, NULL),
    ('royals', 'KCA', 1969, NULL),
    ('kansas city royals', 'KCA', 1969, NULL),
    ('twins', 'MIN', 1961, NULL),
    ('minnesota twins', 'MIN', 1961, NULL),
    ('astros', 'HOU', 1962, NULL),
    ('houston astros', 'HOU', 1962, NULL),
    ('angels', 'ANA', 1961, NULL),
    ('los angeles angels', 'ANA', 1961, NULL),
    ('la angels', 'ANA', 1961, NULL),
    ('anaheim angels', 'ANA', 1997, 2004),
    ('athletics', 'OAK', 1968, NULL),
    ('oakland athletics', 'OAK', 1968, NULL),
    ('as', 'OAK', 1968, NULL),
    ('oakland as', 'OAK', 1968, NULL),
    ('mariners', 'SEA', 1977, NULL),
    ('seattle mariners', 'SEA', 1977, NULL),
    ('rangers', 'TEX', 1972, NULL),
    ('texas rangers', 'TEX', 1972, NULL),

-- National League teams
    ('braves', 'ATL', 1966, NULL),
    ('atlanta braves', 'ATL', 1966, NULL),
    ('marlins', 'FLO', 1993, NULL),
    ('florida marlins', 'FLO', 1993, 2011),
    ('miami marlins', 'FLO', 2012, NULL),
    ('mets', 'NYN', 1962, NULL),
    ('new york mets', 'NYN', 1962, NULL),
    ('ny mets', 'NYN', 1962, NULL),
    ('phillies', 'PHI', 1883, NULL),
    ('philadelphia phillies', 'PHI', 1883, NULL),
    ('nationals', 'WAS', 2005, NULL),
    ('washington nationals', 'WAS', 2005, NULL),
    ('cubs', 'CHN', 1876, NULL),
    ('chicago cubs', 'CHN', 1876, NULL),
    ('reds', 'CIN', 1890, NULL),
    ('cincinnati reds', 'CIN', 1890, NULL),
    ('brewers', 'MIL', 1970, NULL),
    ('milwaukee brewers', 'MIL', 1970, NULL),
    ('pirates', 'PIT', 1887, NULL),
    ('pittsburgh pirates', 'PIT', 1887, NULL),
    ('cardinals', 'SLN', 1892, NULL),
    ('st louis cardinals', 'SLN', 1892, NULL),
    ('stl cardinals', 'SLN', 1892, NULL),
    ('diamondbacks', 'ARI', 1998, NULL),
    ('arizona diamondbacks', 'ARI', 1998, NULL),
    ('dbacks', 'ARI', 1998, NULL),
    ('rockies', 'COL', 1993, NULL),
    ('colorado rockies', 'COL', 1993, NULL),
    ('dodgers', 'LAN', 1958, NULL),
    ('los angeles dodgers', 'LAN', 1958, NULL),
    ('la dodgers', 'LAN', 1958, NULL),
    ('brooklyn dodgers', 'BRO', 1884, 1957),
    ('padres', 'SDN', 1969, NULL),
    ('san diego padres', 'SDN', 1969, NULL),
    ('giants', 'SFN', 1958, NULL),
    ('san francisco giants', 'SFN', 1958, NULL),
    ('sf giants', 'SFN', 1958, NULL),
    ('new york giants', 'NYG', 1883, 1957)
ON CONFLICT (alias) DO NOTHING;

-- Series ID aliases for postseason search
CREATE TABLE IF NOT EXISTS series_aliases (
    alias varchar(50) PRIMARY KEY,
    series_id varchar(10) NOT NULL
);

INSERT INTO series_aliases (alias, series_id) VALUES
    ('world series', 'WS'),
    ('ws', 'WS'),
    ('alcs', 'ALCS'),
    ('nlcs', 'NLCS'),
    ('alds', 'ALDS'),
    ('nlds', 'NLDS'),
    ('al championship', 'ALCS'),
    ('nl championship', 'NLCS'),
    ('al division series', 'ALDS'),
    ('nl division series', 'NLDS'),
    ('wildcard', 'WC'),
    ('wild card', 'WC')
ON CONFLICT (alias) DO NOTHING;

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

CREATE INDEX IF NOT EXISTS idx_woba_constants_season ON woba_constants(season);

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

CREATE INDEX IF NOT EXISTS idx_league_constants_season ON league_constants(season);

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

CREATE INDEX IF NOT EXISTS idx_park_factors_season ON park_factors(season);
CREATE INDEX IF NOT EXISTS idx_park_factors_team ON park_factors(team_id, season);

COMMENT ON TABLE park_factors IS 'Multi-year regressed park factors from FanGraphs (100 = neutral, >100 = hitter-friendly)';
COMMENT ON COLUMN park_factors.basic_5yr IS 'Most stable 5-year regressed overall park factor';
COMMENT ON COLUMN park_factors.factor_hr IS 'Home run park factor (e.g., Coors Field ~105-110)';

-- FanGraphs team name to Retrosheet park ID mapping.
-- Includes stable 2016+ mappings plus historical aliases used by FanGraphs.
CREATE TABLE IF NOT EXISTS fangraphs_team_park_map (
    fangraphs_team VARCHAR(50) PRIMARY KEY,
    retrosheet_team_id VARCHAR(3) NOT NULL,
    primary_park_id VARCHAR(5) NOT NULL,
    start_year INT,
    end_year INT,
    notes TEXT
);

COMMENT ON TABLE fangraphs_team_park_map IS 'Maps FanGraphs team names to Retrosheet park IDs for park factors';

INSERT INTO fangraphs_team_park_map (
    fangraphs_team,
    retrosheet_team_id,
    primary_park_id,
    start_year,
    end_year,
    notes
)
VALUES
    ('Angels', 'ANA', 'ANA01', 2016, NULL, 'Angel Stadium'),
    ('Diamondbacks', 'ARI', 'PHO01', 2016, NULL, 'Chase Field'),
    ('Braves', 'ATL', 'ATL03', 2016, NULL, 'Truist Park'),
    ('Orioles', 'BAL', 'BAL12', 2016, NULL, 'Camden Yards'),
    ('Red Sox', 'BOS', 'BOS07', 2016, NULL, 'Fenway Park'),
    ('White Sox', 'CHA', 'CHI12', 2016, NULL, 'Guaranteed Rate Field'),
    ('Cubs', 'CHN', 'CHI11', 2016, NULL, 'Wrigley Field'),
    ('Reds', 'CIN', 'CIN09', 2016, NULL, 'Great American Ball Park'),
    ('Guardians', 'CLE', 'CLE08', 2022, NULL, 'Progressive Field'),
    ('Indians', 'CLE', 'CLE08', 2016, 2021, 'Cleveland Indians (renamed to Guardians in 2022)'),
    ('Cleveland', 'CLE', 'CLE08', 2021, 2021, 'Cleveland (transition year before Guardians)'),
    ('Rockies', 'COL', 'DEN02', 2016, NULL, 'Coors Field'),
    ('Tigers', 'DET', 'DET05', 2016, NULL, 'Comerica Park'),
    ('Astros', 'HOU', 'HOU03', 2016, NULL, 'Minute Maid Park'),
    ('Royals', 'KCA', 'KAN06', 2016, NULL, 'Kauffman Stadium'),
    ('Dodgers', 'LAN', 'LOS03', 2016, NULL, 'Dodger Stadium'),
    ('Marlins', 'MIA', 'MIA02', 2016, NULL, 'loanDepot park'),
    ('Brewers', 'MIL', 'MIL06', 2016, NULL, 'American Family Field'),
    ('Twins', 'MIN', 'MIN04', 2016, NULL, 'Target Field'),
    ('Yankees', 'NYA', 'NYC21', 2016, NULL, 'Yankee Stadium'),
    ('Mets', 'NYN', 'NYC19', 2016, NULL, 'Citi Field'),
    ('Athletics', 'OAK', 'OAK01', 2016, NULL, 'Oakland Coliseum (used for mapping continuity)'),
    ('Phillies', 'PHI', 'PHI13', 2016, NULL, 'Citizens Bank Park'),
    ('Pirates', 'PIT', 'PIT08', 2016, NULL, 'PNC Park'),
    ('Padres', 'SDN', 'SAN02', 2016, NULL, 'Petco Park'),
    ('Giants', 'SFN', 'SFO03', 2016, NULL, 'Oracle Park'),
    ('Mariners', 'SEA', 'SEA03', 2016, NULL, 'T-Mobile Park'),
    ('Cardinals', 'SLN', 'STL10', 2016, NULL, 'Busch Stadium'),
    ('Rays', 'TBA', 'STP01', 2016, NULL, 'Tropicana Field'),
    ('Rangers', 'TEX', 'ARL02', 2016, NULL, 'Globe Life Field'),
    ('Blue Jays', 'TOR', 'TOR02', 2016, NULL, 'Rogers Centre'),
    ('Nationals', 'WAS', 'WAS11', 2016, NULL, 'Nationals Park')
ON CONFLICT (fangraphs_team) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_fangraphs_team_park_retrosheet ON fangraphs_team_park_map(retrosheet_team_id, start_year);

-- Populate league constants for wRC+ calculations (2023-2024)
-- Uses Lahman Batting table to calculate league-wide constants

-- Calculate league constants for each season/league
-- Excludes pitchers by filtering for players with >= 100 PA (standard threshold)
WITH league_stats AS (
    SELECT
        "yearID" as season,
        "lgID" as league,
        -- Total PA and runs (for reference)
        SUM("AB" + "BB" + "HBP" + "SF") as total_pa,
        SUM("R") as total_runs,
        -- For wOBA average calculation (exclude pitchers: >= 100 PA)
        SUM(CASE WHEN ("AB" + "BB" + "HBP" + "SF") >= 100 THEN "AB" + "BB" - "IBB" + "SF" + "HBP" ELSE 0 END) as qualified_pa_denom,
        SUM(CASE
            WHEN ("AB" + "BB" + "HBP" + "SF") >= 100 THEN
                wc.w_bb * ("BB" - "IBB") + wc.w_hbp * "HBP" +
                wc.w_1b * ("H" - "2B" - "3B" - "HR") +
                wc.w_2b * "2B" + wc.w_3b * "3B" + wc.w_hr * "HR"
            ELSE 0
        END) as qualified_woba_numerator
    FROM "Batting" b
    INNER JOIN woba_constants wc ON wc.season = b."yearID"
    WHERE "yearID" IN (2023, 2024)
        AND "lgID" IN ('AL', 'NL')
    GROUP BY "yearID", "lgID"
),
league_with_calcs AS (
    SELECT
        season,
        league,
        total_pa,
        total_runs,
        -- League average wOBA (excluding pitchers)
        ROUND((qualified_woba_numerator / NULLIF(qualified_pa_denom, 0))::numeric, 4) as woba_avg,
        -- Runs per PA
        ROUND((total_runs::numeric / NULLIF(total_pa, 0))::numeric, 5) as r_pa
    FROM league_stats
)
INSERT INTO league_constants (season, league, woba_avg, wrc_per_pa, runs_per_win, total_pa, total_runs)
SELECT
    lc.season,
    lc.league,
    lc.woba_avg,
    -- wRC per PA: approximately (wOBA - lgwOBA) / wOBA_scale + r/PA
    -- We use a simplified approach: r_pa as the baseline
    ROUND(lc.r_pa::numeric, 5) as wrc_per_pa,
    -- Runs per win: use FanGraphs value from woba_constants
    wc.r_w as runs_per_win,
    lc.total_pa,
    lc.total_runs
FROM league_with_calcs lc
INNER JOIN woba_constants wc ON wc.season = lc.season
ON CONFLICT (season, league) DO UPDATE SET
    woba_avg = EXCLUDED.woba_avg,
    wrc_per_pa = EXCLUDED.wrc_per_pa,
    runs_per_win = EXCLUDED.runs_per_win,
    total_pa = EXCLUDED.total_pa,
    total_runs = EXCLUDED.total_runs,
    calculated_at = NOW();

-- Verify the results
SELECT
    season,
    league,
    woba_avg,
    wrc_per_pa,
    runs_per_win,
    total_pa,
    total_runs
FROM league_constants
ORDER BY season DESC, league;

-- Create table for yearly salary summary data
-- This complements the Salaries table with aggregate statistics
DROP TABLE IF EXISTS salary_summary;
CREATE TABLE salary_summary (
    year INTEGER PRIMARY KEY,
    total NUMERIC(15, 2) NOT NULL,
    average NUMERIC(12, 2) NOT NULL,
    median NUMERIC(12, 2) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_salary_summary_year ON salary_summary(year);

COMMENT ON TABLE salary_summary IS 'Yearly salary summary statistics (total, average, median)';
COMMENT ON COLUMN salary_summary.year IS 'Season year';
COMMENT ON COLUMN salary_summary.total IS 'Total salary for all players in the season';
COMMENT ON COLUMN salary_summary.average IS 'Average salary for the season';
COMMENT ON COLUMN salary_summary.median IS 'Median salary for the season';

-- Create player_relatives table for family relationships
-- Data source: Retrosheet relatives.csv
CREATE TABLE IF NOT EXISTS player_relatives (
    player_id_1 VARCHAR(10) NOT NULL,
    relation_type VARCHAR(50) NOT NULL,
    player_id_2 VARCHAR(10) NOT NULL,
    PRIMARY KEY (player_id_1, player_id_2, relation_type)
);

CREATE INDEX IF NOT EXISTS idx_player_relatives_player1 ON player_relatives(player_id_1);
CREATE INDEX IF NOT EXISTS idx_player_relatives_player2 ON player_relatives(player_id_2);
CREATE INDEX IF NOT EXISTS idx_player_relatives_type ON player_relatives(relation_type);

COMMENT ON TABLE player_relatives IS 'Family relationships between players from Retrosheet biodata';
COMMENT ON COLUMN player_relatives.player_id_1 IS 'First player Retrosheet ID';
COMMENT ON COLUMN player_relatives.relation_type IS 'Type of relationship (Brother, Father, Uncle, etc.)';
COMMENT ON COLUMN player_relatives.player_id_2 IS 'Second player Retrosheet ID';

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

-- Align Batting schema with Lahman 2025+ CSV layout.
-- Lahman 2025 removed legacy Batting columns G_batting and G_old.
ALTER TABLE "Batting"
    DROP COLUMN IF EXISTS "G_batting",
    DROP COLUMN IF EXISTS "G_old";

-- Lahman historical batting rows often leave trailing numeric fields blank.
-- Treat blanks as zero for stable aggregations/scan behavior.
ALTER TABLE "Batting"
    ALTER COLUMN "G" SET DEFAULT 0,
    ALTER COLUMN "AB" SET DEFAULT 0,
    ALTER COLUMN "R" SET DEFAULT 0,
    ALTER COLUMN "H" SET DEFAULT 0,
    ALTER COLUMN "2B" SET DEFAULT 0,
    ALTER COLUMN "3B" SET DEFAULT 0,
    ALTER COLUMN "HR" SET DEFAULT 0,
    ALTER COLUMN "RBI" SET DEFAULT 0,
    ALTER COLUMN "SB" SET DEFAULT 0,
    ALTER COLUMN "CS" SET DEFAULT 0,
    ALTER COLUMN "BB" SET DEFAULT 0,
    ALTER COLUMN "SO" SET DEFAULT 0,
    ALTER COLUMN "IBB" SET DEFAULT 0,
    ALTER COLUMN "HBP" SET DEFAULT 0,
    ALTER COLUMN "SH" SET DEFAULT 0,
    ALTER COLUMN "SF" SET DEFAULT 0,
    ALTER COLUMN "GIDP" SET DEFAULT 0;

UPDATE "Batting"
SET
    "G" = COALESCE("G", 0),
    "AB" = COALESCE("AB", 0),
    "R" = COALESCE("R", 0),
    "H" = COALESCE("H", 0),
    "2B" = COALESCE("2B", 0),
    "3B" = COALESCE("3B", 0),
    "HR" = COALESCE("HR", 0),
    "RBI" = COALESCE("RBI", 0),
    "SB" = COALESCE("SB", 0),
    "CS" = COALESCE("CS", 0),
    "BB" = COALESCE("BB", 0),
    "SO" = COALESCE("SO", 0),
    "IBB" = COALESCE("IBB", 0),
    "HBP" = COALESCE("HBP", 0),
    "SH" = COALESCE("SH", 0),
    "SF" = COALESCE("SF", 0),
    "GIDP" = COALESCE("GIDP", 0)
WHERE
    "G" IS NULL OR
    "AB" IS NULL OR
    "R" IS NULL OR
    "H" IS NULL OR
    "2B" IS NULL OR
    "3B" IS NULL OR
    "HR" IS NULL OR
    "RBI" IS NULL OR
    "SB" IS NULL OR
    "CS" IS NULL OR
    "BB" IS NULL OR
    "SO" IS NULL OR
    "IBB" IS NULL OR
    "HBP" IS NULL OR
    "SH" IS NULL OR
    "SF" IS NULL OR
    "GIDP" IS NULL;

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

-- Add lookup indexes to accelerate player crosswalk joins from Chadwick IDs.
-- These are intentionally non-unique: duplicate IDs can exist in historical edge cases.

CREATE INDEX IF NOT EXISTS "People_retroID_idx" ON "People" ("retroID");
CREATE INDEX IF NOT EXISTS "People_bbrefID_idx" ON "People" ("bbrefID");
