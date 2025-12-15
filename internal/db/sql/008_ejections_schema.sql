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

CREATE INDEX idx_ejections_game_id ON ejections(game_id);
CREATE INDEX idx_ejections_ejectee_id ON ejections(ejectee_id);
CREATE INDEX idx_ejections_umpire_id ON ejections(umpire_id);
CREATE INDEX idx_ejections_date ON ejections(date);
CREATE INDEX idx_ejections_team ON ejections(team);
CREATE INDEX idx_ejections_role ON ejections(role);
