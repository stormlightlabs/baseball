-- Fresh migration set: Retrosheet core schema, indexing, partitioning, and supporting tables.


-- SECTION 002_retrosheet_schema.sql


DROP TABLE IF EXISTS games;
CREATE TABLE games (
    date varchar(8),
    game_number int,
    day_of_week varchar(3),
    visiting_team varchar(3),
    visiting_team_league varchar(3),
    visiting_team_game_number int,
    home_team varchar(3),
    home_team_league varchar(3),
    home_team_game_number int,
    visiting_score int,
    home_score int,
    game_length_outs int,
    day_night varchar(10),
    completion_info text,
    forfeit_info text,
    protest_info text,
    park_id varchar(5),
    attendance int,
    game_time_minutes int,
    visiting_line_score varchar(50),
    home_line_score varchar(50),
    visiting_at_bats int,
    visiting_hits int,
    visiting_doubles int,
    visiting_triples int,
    visiting_homeruns int,
    visiting_rbi int,
    visiting_sac_hits int,
    visiting_sac_flies int,
    visiting_hit_by_pitch int,
    visiting_walks int,
    visiting_int_walks int,
    visiting_strikeouts int,
    visiting_stolen_bases int,
    visiting_caught_stealing int,
    visiting_gdp int,
    visiting_interference int,
    visiting_lob int,
    visiting_pitchers_used int,
    visiting_ind_er int,
    visiting_team_er int,
    visiting_wild_pitches int,
    visiting_balks int,
    visiting_putouts int,
    visiting_assists int,
    visiting_errors int,
    visiting_passed_balls int,
    visiting_double_plays int,
    visiting_triple_plays int,
    home_at_bats int,
    home_hits int,
    home_doubles int,
    home_triples int,
    home_homeruns int,
    home_rbi int,
    home_sac_hits int,
    home_sac_flies int,
    home_hit_by_pitch int,
    home_walks int,
    home_int_walks int,
    home_strikeouts int,
    home_stolen_bases int,
    home_caught_stealing int,
    home_gdp int,
    home_interference int,
    home_lob int,
    home_pitchers_used int,
    home_ind_er int,
    home_team_er int,
    home_wild_pitches int,
    home_balks int,
    home_putouts int,
    home_assists int,
    home_errors int,
    home_passed_balls int,
    home_double_plays int,
    home_triple_plays int,
    hp_ump_id varchar(8),
    hp_ump_name text,
    b1_ump_id varchar(8),
    b1_ump_name text,
    b2_ump_id varchar(8),
    b2_ump_name text,
    b3_ump_id varchar(8),
    b3_ump_name text,
    lf_ump_id varchar(8),
    lf_ump_name text,
    rf_ump_id varchar(8),
    rf_ump_name text,
    v_manager_id varchar(8),
    v_manager_name text,
    h_manager_id varchar(8),
    h_manager_name text,
    winning_pitcher_id varchar(8),
    winning_pitcher_name text,
    losing_pitcher_id varchar(8),
    losing_pitcher_name text,
    saving_pitcher_id varchar(8),
    saving_pitcher_name text,
    goahead_rbi_id varchar(8),
    goahead_rbi_name text,
    v_starting_pitcher_id varchar(8),
    v_starting_pitcher_name text,
    h_starting_pitcher_id varchar(8),
    h_starting_pitcher_name text,
    v_player_1_id varchar(8),
    v_player_1_name text,
    v_player_1_pos int,
    v_player_2_id varchar(8),
    v_player_2_name text,
    v_player_2_pos int,
    v_player_3_id varchar(8),
    v_player_3_name text,
    v_player_3_pos int,
    v_player_4_id varchar(8),
    v_player_4_name text,
    v_player_4_pos int,
    v_player_5_id varchar(8),
    v_player_5_name text,
    v_player_5_pos int,
    v_player_6_id varchar(8),
    v_player_6_name text,
    v_player_6_pos int,
    v_player_7_id varchar(8),
    v_player_7_name text,
    v_player_7_pos int,
    v_player_8_id varchar(8),
    v_player_8_name text,
    v_player_8_pos int,
    v_player_9_id varchar(8),
    v_player_9_name text,
    v_player_9_pos int,
    h_player_1_id varchar(8),
    h_player_1_name text,
    h_player_1_pos int,
    h_player_2_id varchar(8),
    h_player_2_name text,
    h_player_2_pos int,
    h_player_3_id varchar(8),
    h_player_3_name text,
    h_player_3_pos int,
    h_player_4_id varchar(8),
    h_player_4_name text,
    h_player_4_pos int,
    h_player_5_id varchar(8),
    h_player_5_name text,
    h_player_5_pos int,
    h_player_6_id varchar(8),
    h_player_6_name text,
    h_player_6_pos int,
    h_player_7_id varchar(8),
    h_player_7_name text,
    h_player_7_pos int,
    h_player_8_id varchar(8),
    h_player_8_name text,
    h_player_8_pos int,
    h_player_9_id varchar(8),
    h_player_9_name text,
    h_player_9_pos int,
    additional_info text,
    acquisition_info text,
    game_id varchar(16) GENERATED ALWAYS AS (home_team || date || COALESCE(game_number::text, '0')) STORED
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_games_game_id ON games(game_id);




-- SECTION 003_retrosheet_plays_schema.sql


-- Retrosheet Play-by-Play (Plays) Table
-- Source: https://www.retrosheet.org/downloads/plays.html
-- This table contains detailed play-by-play data for each game

DROP TABLE IF EXISTS plays;
CREATE TABLE plays (
    -- Game identification
    gid varchar(12) NOT NULL,           -- Game ID
    event text,                         -- Play as it appears in event file
    inning int,                         -- Inning number
    top_bot int,                        -- Top (0) or bottom (1) of inning
    vis_home int,                       -- Visiting (0) or home (1) team batting
    site varchar(5),                    -- Location (ballpark) of event
    batteam varchar(3),                 -- Batting team
    pitteam varchar(3),                 -- Pitching team
    score_v int,                        -- Visiting team score at start of play
    score_h int,                        -- Home team score at start of play

    -- Players involved
    batter varchar(8),                  -- Batter ID
    pitcher varchar(8),                 -- Pitcher ID
    lp int,                             -- Lineup position of batter
    bat_f int,                          -- Fielding position of batter
    bathand varchar(1),                 -- Batter handedness (B, L, R)
    pithand varchar(1),                 -- Pitcher handedness

    -- Pitch count
    balls int,                          -- Number of balls
    strikes int,                        -- Number of strikes
    count varchar(10),                  -- Pitch count (e.g., "31", "??" if unknown)
    pitches text,                       -- Pitch sequence
    nump int,                           -- Number of pitches

    -- Plate appearance outcome
    pa int,                             -- Plate appearance (1 if yes, 0 if no)
    ab int,                             -- At bat
    single int,                         -- Single
    double int,                         -- Double
    triple int,                         -- Triple
    hr int,                             -- Home run
    sh int,                             -- Sacrifice bunt
    sf int,                             -- Sacrifice fly
    hbp int,                            -- Hit-by-pitch
    walk int,                           -- Walk
    k int,                              -- Strikeout
    xi int,                             -- Reached on interference/obstruction
    roe int,                            -- Reached on error
    fc int,                             -- Fielder's choice
    othout int,                         -- Other batting out
    noout int,                          -- No out (PA not otherwise specified)
    oth int,                            -- Sum of othout and noout (legacy)

    -- Ball in play details
    bip int,                            -- Ball-in-play
    bunt int,                           -- Bunt
    ground int,                         -- Ground ball
    fly int,                            -- Fly ball or pop up
    line int,                           -- Line drive

    -- Special plays
    iw int,                             -- Intentional walk
    gdp int,                            -- Grounded into double play
    othdp int,                          -- Other double play
    tp int,                             -- Triple play
    fle int,                            -- Dropped foul-ball error
    wp int,                             -- Wild pitch
    pb int,                             -- Passed ball
    bk int,                             -- Balk
    oa int,                             -- Out advancing or other advance
    di int,                             -- Defensive indifference

    -- Stolen bases and caught stealing
    sb2 int,                            -- Stolen base of second
    sb3 int,                            -- Stolen base of third
    sbh int,                            -- Stolen base of home
    cs2 int,                            -- Caught stealing second
    cs3 int,                            -- Caught stealing third
    csh int,                            -- Caught stealing home

    -- Pickoffs
    pko1 int,                           -- Pickoff at first
    pko2 int,                           -- Pickoff at second
    pko3 int,                           -- Pickoff at third
    k_safe int,                         -- Strikeout but reached base safely

    -- Errors by position
    e1 int,                             -- Error by pitcher
    e2 int,                             -- Error by catcher
    e3 int,                             -- Error by first baseman
    e4 int,                             -- Error by second baseman
    e5 int,                             -- Error by third baseman
    e6 int,                             -- Error by shortstop
    e7 int,                             -- Error by left fielder
    e8 int,                             -- Error by center fielder
    e9 int,                             -- Error by right fielder

    -- Outs and baserunners before/after play
    outs_pre int,                       -- Outs before the play
    outs_post int,                      -- Outs after the play
    br1_pre varchar(8),                 -- Runner on first before play
    br2_pre varchar(8),                 -- Runner on second before play
    br3_pre varchar(8),                 -- Runner on third before play
    br1_post varchar(8),                -- Runner on first after play
    br2_post varchar(8),                -- Runner on second after play
    br3_post varchar(8),                -- Runner on third after play

    -- Runners left on base
    lob_id1 varchar(8),                 -- Runner 1 left on base after third out
    lob_id2 varchar(8),                 -- Runner 2 left on base after third out
    lob_id3 varchar(8),                 -- Runner 3 left on base after third out

    -- Pitcher responsible for runners before/after
    pr1_pre varchar(8),                 -- Pitcher responsible for runner on first (pre)
    pr2_pre varchar(8),                 -- Pitcher responsible for runner on second (pre)
    pr3_pre varchar(8),                 -- Pitcher responsible for runner on third (pre)
    pr1_post varchar(8),                -- Pitcher responsible for runner on first (post)
    pr2_post varchar(8),                -- Pitcher responsible for runner on second (post)
    pr3_post varchar(8),                -- Pitcher responsible for runner on third (post)

    -- Runs scored
    run_b varchar(8),                   -- Batter if he scored
    run1 varchar(8),                    -- Runner on first if he scored
    run2 varchar(8),                    -- Runner on second if he scored
    run3 varchar(8),                    -- Runner on third if he scored
    prun_b varchar(8),                  -- Pitcher charged with run (batter)
    prun1 varchar(8),                   -- Pitcher charged with run (runner 1)
    prun2 varchar(8),                   -- Pitcher charged with run (runner 2)
    prun3 varchar(8),                   -- Pitcher charged with run (runner 3)

    -- Unearned runs
    ur_b int,                           -- Unearned run scored by batter
    ur1 int,                            -- Unearned run scored by runner 1
    ur2 int,                            -- Unearned run scored by runner 2
    ur3 int,                            -- Unearned run scored by runner 3

    -- RBIs
    rbi_b int,                          -- RBI for batter's run
    rbi1 int,                           -- RBI for runner 1's run
    rbi2 int,                           -- RBI for runner 2's run
    rbi3 int,                           -- RBI for runner 3's run
    runs int,                           -- Total runs scored on the play
    rbi int,                            -- Total RBI credited to batter
    er int,                             -- Total earned runs scored
    tur int,                            -- Team unearned runs

    -- Lineups (batting team)
    l1 varchar(8),                      -- Lineup position 1
    l2 varchar(8),                      -- Lineup position 2
    l3 varchar(8),                      -- Lineup position 3
    l4 varchar(8),                      -- Lineup position 4
    l5 varchar(8),                      -- Lineup position 5
    l6 varchar(8),                      -- Lineup position 6
    l7 varchar(8),                      -- Lineup position 7
    l8 varchar(8),                      -- Lineup position 8
    l9 varchar(8),                      -- Lineup position 9

    -- Fielding positions of lineup (batting team)
    lf1 int,                            -- Fielding position of lineup 1
    lf2 int,                            -- Fielding position of lineup 2
    lf3 int,                            -- Fielding position of lineup 3
    lf4 int,                            -- Fielding position of lineup 4
    lf5 int,                            -- Fielding position of lineup 5
    lf6 int,                            -- Fielding position of lineup 6
    lf7 int,                            -- Fielding position of lineup 7
    lf8 int,                            -- Fielding position of lineup 8
    lf9 int,                            -- Fielding position of lineup 9

    -- Fielding team positions
    f2 varchar(8),                      -- Catcher
    f3 varchar(8),                      -- First baseman
    f4 varchar(8),                      -- Second baseman
    f5 varchar(8),                      -- Third baseman
    f6 varchar(8),                      -- Shortstop
    f7 varchar(8),                      -- Left fielder
    f8 varchar(8),                      -- Center fielder
    f9 varchar(8),                      -- Right fielder

    -- Putouts by position
    po0 int,                            -- Putouts (fielder unknown)
    po1 int,                            -- Putouts by pitcher
    po2 int,                            -- Putouts by catcher
    po3 int,                            -- Putouts by first baseman
    po4 int,                            -- Putouts by second baseman
    po5 int,                            -- Putouts by third baseman
    po6 int,                            -- Putouts by shortstop
    po7 int,                            -- Putouts by left fielder
    po8 int,                            -- Putouts by center fielder
    po9 int,                            -- Putouts by right fielder

    -- Assists by position
    a1 int,                             -- Assists by pitcher
    a2 int,                             -- Assists by catcher
    a3 int,                             -- Assists by first baseman
    a4 int,                             -- Assists by second baseman
    a5 int,                             -- Assists by third baseman
    a6 int,                             -- Assists by shortstop
    a7 int,                             -- Assists by left fielder
    a8 int,                             -- Assists by center fielder
    a9 int,                             -- Assists by right fielder

    -- Fielding sequence and outs
    fseq varchar(20),                   -- Fielding sequence (e.g., "643" for 6-4-3 DP)
    batout1 int,                        -- Position initiating first batting out
    batout2 int,                        -- Position initiating second batting out (DP)
    batout3 int,                        -- Position initiating third batting out (TP)
    brout_b int,                        -- Position initiating baserunning out (batter)
    brout1 int,                         -- Position initiating baserunning out (runner 1)
    brout2 int,                         -- Position initiating baserunning out (runner 2)
    brout3 int,                         -- Position initiating baserunning out (runner 3)

    -- Ball in play details
    firstf int,                         -- First fielder to field ball (1-9, 0 if unknown)
    loc varchar(10),                    -- Location of ball in play
    hittype varchar(5),                 -- Hit type (BG, BP, BL, G, P, F, L)
    dpopp int,                          -- Double play opportunity (runner on first, <2 outs)
    pivot int,                          -- Pivot man on DP opportunity
    pn int,                             -- Play number (sequential within game)

    -- Umpires
    umphome varchar(8),                 -- Home plate umpire
    ump1b varchar(8),                   -- First base umpire
    ump2b varchar(8),                   -- Second base umpire
    ump3b varchar(8),                   -- Third base umpire
    umplf varchar(8),                   -- Left field umpire
    umprf varchar(8),                   -- Right field umpire

    -- Game metadata
    date varchar(8),                    -- Date of game (YYYYMMDD)
    gametype varchar(20),               -- Type of game (regular, etc.)
    pbp varchar(10)                     -- Play-by-play type (deduced or full)
);

CREATE INDEX IF NOT EXISTS idx_plays_gid ON plays(gid);
CREATE INDEX IF NOT EXISTS idx_plays_batter ON plays(batter);
CREATE INDEX IF NOT EXISTS idx_plays_pitcher ON plays(pitcher);
CREATE INDEX IF NOT EXISTS idx_plays_date ON plays(date);
CREATE INDEX IF NOT EXISTS idx_plays_batteam ON plays(batteam);
CREATE INDEX IF NOT EXISTS idx_plays_pitteam ON plays(pitteam);




-- SECTION 006_add_primary_keys.sql


-- Add primary key to games table
-- A game is uniquely identified by date, home team, and game number (for doubleheaders)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'games'::regclass
            AND contype = 'p'
    ) THEN
        ALTER TABLE games ADD PRIMARY KEY (date, home_team, game_number);
    END IF;
END $$;

-- Add primary key to plays table
-- A play is uniquely identified by game ID and play number
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'plays'::regclass
            AND contype = 'p'
    ) THEN
        ALTER TABLE plays ADD PRIMARY KEY (gid, pn);
    END IF;
END $$;




-- SECTION 007_add_game_type.sql


-- Add game_type column to games table to distinguish regular season, postseason, and all-star games
ALTER TABLE games ADD COLUMN IF NOT EXISTS game_type varchar(20) DEFAULT 'regular';

-- Create index for filtering by game type
CREATE INDEX IF NOT EXISTS idx_games_game_type ON games(game_type);




-- SECTION 008_ejections_schema.sql


-- Retrosheet Ejections Table
-- Source: https://www.retrosheet.org/eject.htm
-- This table contains ejection data from Retrosheet's ejections database

DROP TABLE IF EXISTS ejections;
CREATE TABLE ejections (
    -- Game identification
    game_id varchar(20) NOT NULL,          -- Retrosheet game ID
    date varchar(10) NOT NULL,             -- Date (MM/DD/YYYY)
    game_number int,                       -- Game number (blank for single games)

    -- Ejected person
    ejectee_id varchar(20) NOT NULL,       -- Retrosheet ID of ejected person
    ejectee_name text NOT NULL,            -- Name of ejected person
    team varchar(3),                       -- Team abbreviation
    role varchar(1) NOT NULL,              -- P=Player, M=Manager, C=Coach

    -- Umpire who made the ejection
    umpire_id varchar(20),                 -- Retrosheet ID of umpire
    umpire_name text,                      -- Name of umpire

    -- Ejection details
    inning int,                            -- Inning when ejection occurred (-1 if unknown)
    reason text,                           -- Brief explanation for ejection

    CONSTRAINT ejections_game_ejectee_unique UNIQUE (game_id, ejectee_id)
);

CREATE INDEX IF NOT EXISTS idx_ejections_game_id ON ejections(game_id);
CREATE INDEX IF NOT EXISTS idx_ejections_ejectee_id ON ejections(ejectee_id);
CREATE INDEX IF NOT EXISTS idx_ejections_umpire_id ON ejections(umpire_id);
CREATE INDEX IF NOT EXISTS idx_ejections_date ON ejections(date);
CREATE INDEX IF NOT EXISTS idx_ejections_team ON ejections(team);
CREATE INDEX IF NOT EXISTS idx_ejections_role ON ejections(role);




-- SECTION 009_game_search_columns.sql


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




-- SECTION 015_add_query_indexes.sql


-- Additional indexes to optimize Negro Leagues endpoints and pagination-heavy queries

-- Helps /negroleagues/plays when sorting by play number within games
CREATE INDEX IF NOT EXISTS idx_plays_gid_pn ON plays(gid, pn);

-- Reduce scans when filtering games by league + date (both home and visiting)
CREATE INDEX IF NOT EXISTS idx_games_home_league_date ON games(home_team_league, date);
CREATE INDEX IF NOT EXISTS idx_games_visiting_league_date ON games(visiting_team_league, date);

-- General date filter for schedule endpoints
CREATE INDEX IF NOT EXISTS idx_games_date ON games(date);




-- SECTION 032_add_weather_columns.sql


-- Add weather and game condition columns to games table
ALTER TABLE games ADD COLUMN IF NOT EXISTS temp_f INTEGER;
ALTER TABLE games ADD COLUMN IF NOT EXISTS sky VARCHAR(20);
ALTER TABLE games ADD COLUMN IF NOT EXISTS wind_direction VARCHAR(20);
ALTER TABLE games ADD COLUMN IF NOT EXISTS wind_speed_mph INTEGER;
ALTER TABLE games ADD COLUMN IF NOT EXISTS precip VARCHAR(20);
ALTER TABLE games ADD COLUMN IF NOT EXISTS field_condition VARCHAR(20);
ALTER TABLE games ADD COLUMN IF NOT EXISTS start_time TIME;
ALTER TABLE games ADD COLUMN IF NOT EXISTS used_dh BOOLEAN;

CREATE INDEX IF NOT EXISTS idx_games_used_dh ON games(used_dh) WHERE used_dh IS NOT NULL;




-- SECTION 033_partition_plays_table.sql


-- Partition the plays table by year to improve query performance.
-- Idempotent: skips when plays is already partitioned.

DO $$
DECLARE
    part record;
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_partitioned_table pt
        INNER JOIN pg_class c ON c.oid = pt.partrelid
        WHERE c.relname = 'plays'
    ) THEN
        RAISE NOTICE 'plays is already partitioned; skipping migration 033';
        RETURN;
    END IF;

    -- Step 1: Create new partitioned table with same schema as plays
    EXECUTE $ddl$
        CREATE TABLE plays_partitioned (
            gid character varying(12) NOT NULL,
            event text,
            inning integer,
            top_bot integer,
            vis_home integer,
            site character varying(5),
            batteam character varying(3),
            pitteam character varying(3),
            score_v integer,
            score_h integer,
            batter character varying(8),
            pitcher character varying(8),
            lp integer,
            bat_f integer,
            bathand character varying(1),
            pithand character varying(1),
            balls integer,
            strikes integer,
            count character varying(10),
            pitches text,
            nump integer,
            pa integer,
            ab integer,
            single integer,
            double integer,
            triple integer,
            hr integer,
            sh integer,
            sf integer,
            hbp integer,
            walk integer,
            k integer,
            xi integer,
            roe integer,
            fc integer,
            othout integer,
            noout integer,
            oth integer,
            bip integer,
            bunt integer,
            ground integer,
            fly integer,
            line integer,
            iw integer,
            gdp integer,
            othdp integer,
            tp integer,
            fle integer,
            wp integer,
            pb integer,
            bk integer,
            oa integer,
            di integer,
            sb2 integer,
            sb3 integer,
            sbh integer,
            cs2 integer,
            cs3 integer,
            csh integer,
            pko1 integer,
            pko2 integer,
            pko3 integer,
            k_safe integer,
            e1 integer,
            e2 integer,
            e3 integer,
            e4 integer,
            e5 integer,
            e6 integer,
            e7 integer,
            e8 integer,
            e9 integer,
            outs_pre integer,
            outs_post integer,
            br1_pre character varying(8),
            br2_pre character varying(8),
            br3_pre character varying(8),
            br1_post character varying(8),
            br2_post character varying(8),
            br3_post character varying(8),
            lob_id1 character varying(8),
            lob_id2 character varying(8),
            lob_id3 character varying(8),
            pr1_pre character varying(8),
            pr2_pre character varying(8),
            pr3_pre character varying(8),
            pr1_post character varying(8),
            pr2_post character varying(8),
            pr3_post character varying(8),
            run_b character varying(8),
            run1 character varying(8),
            run2 character varying(8),
            run3 character varying(8),
            prun_b character varying(8),
            prun1 character varying(8),
            prun2 character varying(8),
            prun3 character varying(8),
            ur_b integer,
            ur1 integer,
            ur2 integer,
            ur3 integer,
            rbi_b integer,
            rbi1 integer,
            rbi2 integer,
            rbi3 integer,
            runs integer,
            rbi integer,
            er integer,
            tur integer,
            l1 character varying(8),
            l2 character varying(8),
            l3 character varying(8),
            l4 character varying(8),
            l5 character varying(8),
            l6 character varying(8),
            l7 character varying(8),
            l8 character varying(8),
            l9 character varying(8),
            lf1 integer,
            lf2 integer,
            lf3 integer,
            lf4 integer,
            lf5 integer,
            lf6 integer,
            lf7 integer,
            lf8 integer,
            lf9 integer,
            f2 character varying(8),
            f3 character varying(8),
            f4 character varying(8),
            f5 character varying(8),
            f6 character varying(8),
            f7 character varying(8),
            f8 character varying(8),
            f9 character varying(8),
            po0 integer,
            po1 integer,
            po2 integer,
            po3 integer,
            po4 integer,
            po5 integer,
            po6 integer,
            po7 integer,
            po8 integer,
            po9 integer,
            a1 integer,
            a2 integer,
            a3 integer,
            a4 integer,
            a5 integer,
            a6 integer,
            a7 integer,
            a8 integer,
            a9 integer,
            fseq character varying(20),
            batout1 integer,
            batout2 integer,
            batout3 integer,
            brout_b integer,
            brout1 integer,
            brout2 integer,
            brout3 integer,
            firstf integer,
            loc character varying(10),
            hittype character varying(5),
            dpopp integer,
            pivot integer,
            pn integer NOT NULL,
            umphome character varying(8),
            ump1b character varying(8),
            ump2b character varying(8),
            ump3b character varying(8),
            umplf character varying(8),
            umprf character varying(8),
            date character varying(8),
            gametype character varying(20),
            pbp character varying(10),
            PRIMARY KEY (gid, pn, date)
        ) PARTITION BY RANGE (date)
    $ddl$;

    -- Step 2: Create partitions for each era/year
    FOR part IN
        SELECT *
        FROM (
            VALUES
                ('plays_pre1914', '00000000', '19140000'),
                ('plays_1914', '19140000', '19150000'),
                ('plays_1915', '19150000', '19160000'),
                ('plays_1916_1934', '19160000', '19350000'),
                ('plays_1935', '19350000', '19360000'),
                ('plays_1936', '19360000', '19370000'),
                ('plays_1937', '19370000', '19380000'),
                ('plays_1938', '19380000', '19390000'),
                ('plays_1939', '19390000', '19400000'),
                ('plays_1940', '19400000', '19410000'),
                ('plays_1941', '19410000', '19420000'),
                ('plays_1942', '19420000', '19430000'),
                ('plays_1943', '19430000', '19440000'),
                ('plays_1944', '19440000', '19450000'),
                ('plays_1945', '19450000', '19460000'),
                ('plays_1946', '19460000', '19470000'),
                ('plays_1947', '19470000', '19480000'),
                ('plays_1948', '19480000', '19490000'),
                ('plays_1949', '19490000', '19500000'),
                ('plays_1950_1962', '19500000', '19630000'),
                ('plays_1963_1968', '19630000', '19690000'),
                ('plays_1969_1973', '19690000', '19740000'),
                ('plays_1974_1978', '19740000', '19790000'),
                ('plays_1979_1983', '19790000', '19840000'),
                ('plays_1984_1988', '19840000', '19890000'),
                ('plays_1989_1993', '19890000', '19940000'),
                ('plays_1994', '19940000', '19950000'),
                ('plays_1995', '19950000', '19960000'),
                ('plays_1996', '19960000', '19970000'),
                ('plays_1997', '19970000', '19980000'),
                ('plays_1998', '19980000', '19990000'),
                ('plays_1999', '19990000', '20000000'),
                ('plays_2000', '20000000', '20010000'),
                ('plays_2001', '20010000', '20020000'),
                ('plays_2002', '20020000', '20030000'),
                ('plays_2003', '20030000', '20040000'),
                ('plays_2004', '20040000', '20050000'),
                ('plays_2005', '20050000', '20060000'),
                ('plays_2006', '20060000', '20070000'),
                ('plays_2007', '20070000', '20080000'),
                ('plays_2008', '20080000', '20090000'),
                ('plays_2009', '20090000', '20100000'),
                ('plays_2010', '20100000', '20110000'),
                ('plays_2011', '20110000', '20120000'),
                ('plays_2012', '20120000', '20130000'),
                ('plays_2013', '20130000', '20140000'),
                ('plays_2014', '20140000', '20150000'),
                ('plays_2015', '20150000', '20160000'),
                ('plays_2016', '20160000', '20170000'),
                ('plays_2017', '20170000', '20180000'),
                ('plays_2018', '20180000', '20190000'),
                ('plays_2019', '20190000', '20200000'),
                ('plays_2020', '20200000', '20210000'),
                ('plays_2021', '20210000', '20220000'),
                ('plays_2022', '20220000', '20230000'),
                ('plays_2023', '20230000', '20240000'),
                ('plays_2024', '20240000', '20250000'),
                ('plays_2025', '20250000', '20260000'),
                ('plays_2026', '20260000', '20270000'),
                ('plays_2027', '20270000', '20280000'),
                ('plays_2028', '20280000', '20290000'),
                ('plays_2029', '20290000', '20300000'),
                ('plays_2030', '20300000', '20310000')
        ) AS partitions(partition_name, range_start, range_end)
    LOOP
        EXECUTE format(
            'CREATE TABLE %I PARTITION OF plays_partitioned FOR VALUES FROM (%L) TO (%L)',
            part.partition_name,
            part.range_start,
            part.range_end
        );
    END LOOP;

    EXECUTE 'CREATE TABLE plays_default PARTITION OF plays_partitioned DEFAULT';

    -- Step 3: Migrate existing data.
    EXECUTE 'INSERT INTO plays_partitioned SELECT * FROM plays';

    -- Step 4: Create indexes on the partitioned table.
    -- These will automatically be created on all child partitions.
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_gid ON plays_partitioned(gid)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_gid_pn ON plays_partitioned(gid, pn)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_date ON plays_partitioned(date)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_batter ON plays_partitioned(batter)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_pitcher ON plays_partitioned(pitcher)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_batteam ON plays_partitioned(batteam)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_pitteam ON plays_partitioned(pitteam)';

    -- Step 5: Swap tables atomically. Migration runner already wraps this migration in a transaction.
    EXECUTE 'DROP TABLE plays CASCADE';
    EXECUTE 'ALTER TABLE plays_partitioned RENAME TO plays';

    -- Step 6: Analyze new table for query planner.
    EXECUTE 'ANALYZE plays';
END
$$;




-- SECTION 034_denormalize_league_to_plays.sql


-- Add league columns to plays table for better partition pruning

ALTER TABLE plays ADD COLUMN IF NOT EXISTS home_team_league character varying(10);
ALTER TABLE plays ADD COLUMN IF NOT EXISTS visiting_team_league character varying(10);

UPDATE plays p
SET
    home_team_league = g.home_team_league,
    visiting_team_league = g.visiting_team_league
FROM games g
WHERE p.gid = g.game_id;

CREATE INDEX IF NOT EXISTS idx_plays_home_team_league ON plays(home_team_league);
CREATE INDEX IF NOT EXISTS idx_plays_visiting_team_league ON plays(visiting_team_league);

CREATE INDEX IF NOT EXISTS idx_plays_home_league_date ON plays(home_team_league, date);
CREATE INDEX IF NOT EXISTS idx_plays_visiting_league_date ON plays(visiting_team_league, date);

ANALYZE plays;




-- SECTION 040_retrosheet_players.sql


-- Retrosheet allplayers data
-- Provides per-team-season player appearances with granular positional data

DROP TABLE IF EXISTS retrosheet_players CASCADE;

CREATE TABLE retrosheet_players (
    player_id VARCHAR(8) NOT NULL,       -- Retrosheet player ID
    last_name VARCHAR(50),
    first_name VARCHAR(50),
    bats VARCHAR(1),                     -- L, R, B (both), ? (unknown)
    throws VARCHAR(1),                   -- L, R, ? (unknown)
    team_id VARCHAR(3) NOT NULL,
    season INTEGER NOT NULL,
    games INTEGER DEFAULT 0,             -- Total games
    games_p INTEGER DEFAULT 0,           -- Games as pitcher
    games_sp INTEGER DEFAULT 0,          -- Games as starting pitcher
    games_rp INTEGER DEFAULT 0,          -- Games as relief pitcher
    games_c INTEGER DEFAULT 0,           -- Catcher
    games_1b INTEGER DEFAULT 0,          -- First base
    games_2b INTEGER DEFAULT 0,          -- Second base
    games_3b INTEGER DEFAULT 0,          -- Third base
    games_ss INTEGER DEFAULT 0,          -- Shortstop
    games_lf INTEGER DEFAULT 0,          -- Left field
    games_cf INTEGER DEFAULT 0,          -- Center field
    games_rf INTEGER DEFAULT 0,          -- Right field
    games_of INTEGER DEFAULT 0,          -- Outfield
    games_dh INTEGER DEFAULT 0,          -- Designated hitter
    games_ph INTEGER DEFAULT 0,          -- Pinch hitter
    games_pr INTEGER DEFAULT 0,          -- Pinch runner

    -- Date range (YYYYMMDD format)
    first_game DATE,                     -- First game with this team/season
    last_game DATE,                      -- Last game with this team/season

    PRIMARY KEY (player_id, team_id, season)
);

CREATE INDEX IF NOT EXISTS idx_retrosheet_players_player ON retrosheet_players(player_id);
CREATE INDEX IF NOT EXISTS idx_retrosheet_players_season ON retrosheet_players(season);
CREATE INDEX IF NOT EXISTS idx_retrosheet_players_team_season ON retrosheet_players(team_id, season);
CREATE INDEX IF NOT EXISTS idx_retrosheet_players_player_season ON retrosheet_players(player_id, season);

COMMENT ON TABLE retrosheet_players IS 'Per-team-season player appearances from Retrosheet allplayers.csv. Provides granular positional data including pitcher roles (starter/reliever) and exact game date ranges.';
COMMENT ON COLUMN retrosheet_players.games_sp IS 'Games as starting pitcher (first pitcher of game)';
COMMENT ON COLUMN retrosheet_players.games_rp IS 'Games as relief pitcher (entered after game started)';
COMMENT ON COLUMN retrosheet_players.first_game IS 'Date of first game with this team in this season';
COMMENT ON COLUMN retrosheet_players.last_game IS 'Date of last game with this team in this season';




-- SECTION 044_retrosheet_coaches.sql


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




-- SECTION 045_retrosheet_umpires.sql


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


