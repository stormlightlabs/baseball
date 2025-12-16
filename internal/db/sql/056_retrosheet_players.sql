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

CREATE INDEX idx_retrosheet_players_player ON retrosheet_players(player_id);
CREATE INDEX idx_retrosheet_players_season ON retrosheet_players(season);
CREATE INDEX idx_retrosheet_players_team_season ON retrosheet_players(team_id, season);
CREATE INDEX idx_retrosheet_players_player_season ON retrosheet_players(player_id, season);

COMMENT ON TABLE retrosheet_players IS 'Per-team-season player appearances from Retrosheet allplayers.csv. Provides granular positional data including pitcher roles (starter/reliever) and exact game date ranges.';
COMMENT ON COLUMN retrosheet_players.games_sp IS 'Games as starting pitcher (first pitcher of game)';
COMMENT ON COLUMN retrosheet_players.games_rp IS 'Games as relief pitcher (entered after game started)';
COMMENT ON COLUMN retrosheet_players.first_game IS 'Date of first game with this team in this season';
COMMENT ON COLUMN retrosheet_players.last_game IS 'Date of last game with this team in this season';
