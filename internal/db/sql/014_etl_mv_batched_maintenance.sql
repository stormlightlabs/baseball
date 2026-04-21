-- Track ETL delta scopes per run so maintenance can target affected slices.
CREATE TABLE IF NOT EXISTS etl_delta_games (
    run_id BIGINT NOT NULL REFERENCES etl_runs(id) ON DELETE CASCADE,
    game_id TEXT NOT NULL,
    season INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (run_id, game_id)
);

CREATE INDEX IF NOT EXISTS idx_etl_delta_games_season
ON etl_delta_games (season, run_id);

CREATE TABLE IF NOT EXISTS etl_delta_seasons (
    run_id BIGINT NOT NULL REFERENCES etl_runs(id) ON DELETE CASCADE,
    season INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (run_id, season)
);

CREATE INDEX IF NOT EXISTS idx_etl_delta_seasons_season
ON etl_delta_seasons (season, run_id);

CREATE TABLE IF NOT EXISTS etl_delta_players (
    run_id BIGINT NOT NULL REFERENCES etl_runs(id) ON DELETE CASCADE,
    player_id TEXT NOT NULL,
    season INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (run_id, player_id)
);

CREATE INDEX IF NOT EXISTS idx_etl_delta_players_season
ON etl_delta_players (season, run_id);

-- Serving/intermediary tables that can be refreshed in bounded maintenance jobs.
CREATE TABLE IF NOT EXISTS serving_player_game_batting_stats AS
SELECT * FROM player_game_batting_stats WITH NO DATA;

CREATE TABLE IF NOT EXISTS serving_player_game_pitching_stats AS
SELECT * FROM player_game_pitching_stats WITH NO DATA;

CREATE TABLE IF NOT EXISTS serving_player_game_fielding_stats AS
SELECT * FROM player_game_fielding_stats WITH NO DATA;

CREATE TABLE IF NOT EXISTS serving_team_game_stats AS
SELECT * FROM team_game_stats WITH NO DATA;

CREATE TABLE IF NOT EXISTS serving_no_hitters AS
SELECT * FROM no_hitters WITH NO DATA;

CREATE TABLE IF NOT EXISTS serving_cycles AS
SELECT * FROM cycles WITH NO DATA;

CREATE TABLE IF NOT EXISTS serving_multi_hr_games AS
SELECT * FROM multi_hr_games WITH NO DATA;

CREATE TABLE IF NOT EXISTS serving_triple_plays AS
SELECT * FROM triple_plays WITH NO DATA;

CREATE TABLE IF NOT EXISTS serving_extra_inning_games AS
SELECT * FROM extra_inning_games WITH NO DATA;

CREATE TABLE IF NOT EXISTS serving_season_batting_leaders AS
SELECT * FROM season_batting_leaders WITH NO DATA;

CREATE TABLE IF NOT EXISTS serving_season_pitching_leaders AS
SELECT * FROM season_pitching_leaders WITH NO DATA;

CREATE TABLE IF NOT EXISTS serving_career_batting_leaders AS
SELECT * FROM career_batting_leaders WITH NO DATA;

CREATE TABLE IF NOT EXISTS serving_career_pitching_leaders AS
SELECT * FROM career_pitching_leaders WITH NO DATA;

-- Incremental win expectancy intermediaries.
CREATE TABLE IF NOT EXISTS serving_win_expectancy_state_counts (
    season INTEGER NOT NULL,
    inning INTEGER NOT NULL,
    is_bottom BOOLEAN NOT NULL,
    outs INTEGER NOT NULL,
    runners_state TEXT NOT NULL,
    score_diff INTEGER NOT NULL,
    sample_size BIGINT NOT NULL,
    home_win_samples BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (season, inning, is_bottom, outs, runners_state, score_diff)
);

CREATE INDEX IF NOT EXISTS idx_serving_win_expectancy_state_counts_state
ON serving_win_expectancy_state_counts (inning, is_bottom, outs, runners_state, score_diff, season);

CREATE TABLE IF NOT EXISTS serving_win_expectancy_historical (
    id BIGSERIAL PRIMARY KEY,
    inning INTEGER NOT NULL,
    is_bottom BOOLEAN NOT NULL,
    outs INTEGER NOT NULL,
    runners_state TEXT NOT NULL,
    score_diff INTEGER NOT NULL,
    win_probability DOUBLE PRECISION NOT NULL,
    sample_size BIGINT NOT NULL,
    start_year INTEGER NOT NULL,
    end_year INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_serving_win_expectancy_state
ON serving_win_expectancy_historical (inning, is_bottom, outs, runners_state, score_diff, start_year, end_year);

CREATE INDEX IF NOT EXISTS idx_serving_win_expectancy_inning
ON serving_win_expectancy_historical (inning);

CREATE INDEX IF NOT EXISTS idx_serving_win_expectancy_outs
ON serving_win_expectancy_historical (outs);

CREATE INDEX IF NOT EXISTS idx_serving_win_expectancy_runners
ON serving_win_expectancy_historical (runners_state);

CREATE INDEX IF NOT EXISTS idx_serving_win_expectancy_score
ON serving_win_expectancy_historical (score_diff);

-- Practical lookup indexes for serving datasets.
CREATE INDEX IF NOT EXISTS idx_serving_player_game_batting_player_date
ON serving_player_game_batting_stats (player_id, date DESC);

CREATE INDEX IF NOT EXISTS idx_serving_player_game_batting_player_season
ON serving_player_game_batting_stats (player_id, season DESC);

CREATE INDEX IF NOT EXISTS idx_serving_player_game_batting_game
ON serving_player_game_batting_stats (game_id);

CREATE INDEX IF NOT EXISTS idx_serving_player_game_pitching_player_date
ON serving_player_game_pitching_stats (player_id, date DESC);

CREATE INDEX IF NOT EXISTS idx_serving_player_game_pitching_player_season
ON serving_player_game_pitching_stats (player_id, season DESC);

CREATE INDEX IF NOT EXISTS idx_serving_player_game_pitching_game
ON serving_player_game_pitching_stats (game_id);

CREATE INDEX IF NOT EXISTS idx_serving_player_game_fielding_player_date
ON serving_player_game_fielding_stats (player_id, date DESC);

CREATE INDEX IF NOT EXISTS idx_serving_player_game_fielding_player_season
ON serving_player_game_fielding_stats (player_id, season DESC);

CREATE INDEX IF NOT EXISTS idx_serving_player_game_fielding_game
ON serving_player_game_fielding_stats (game_id);

CREATE INDEX IF NOT EXISTS idx_serving_team_game_stats_team_date
ON serving_team_game_stats (team_id, date DESC);

CREATE INDEX IF NOT EXISTS idx_serving_team_game_stats_team_season
ON serving_team_game_stats (team_id, season DESC);

CREATE INDEX IF NOT EXISTS idx_serving_team_game_stats_game
ON serving_team_game_stats (game_id);

CREATE INDEX IF NOT EXISTS idx_serving_no_hitters_season
ON serving_no_hitters (season, date DESC);

CREATE INDEX IF NOT EXISTS idx_serving_cycles_player_season
ON serving_cycles (player_id, season, date DESC);

CREATE INDEX IF NOT EXISTS idx_serving_multi_hr_player_season
ON serving_multi_hr_games (player_id, season, date DESC);

CREATE INDEX IF NOT EXISTS idx_serving_triple_plays_team_season
ON serving_triple_plays (team_id, season, date DESC);

CREATE INDEX IF NOT EXISTS idx_serving_extra_inning_season
ON serving_extra_inning_games (season, date DESC);

CREATE INDEX IF NOT EXISTS idx_serving_season_batting_player_season
ON serving_season_batting_leaders (player_id, season DESC);

CREATE INDEX IF NOT EXISTS idx_serving_season_pitching_player_season
ON serving_season_pitching_leaders (player_id, season DESC);

CREATE INDEX IF NOT EXISTS idx_serving_career_batting_player
ON serving_career_batting_leaders (player_id);

CREATE INDEX IF NOT EXISTS idx_serving_career_pitching_player
ON serving_career_pitching_leaders (player_id);
