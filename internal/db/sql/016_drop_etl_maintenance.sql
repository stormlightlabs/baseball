-- Remove dormant ETL maintenance delta/serving tables.
-- Maintenance now refreshes canonical materialized views directly.

DROP TABLE IF EXISTS etl_delta_games CASCADE;
DROP TABLE IF EXISTS etl_delta_seasons CASCADE;
DROP TABLE IF EXISTS etl_delta_players CASCADE;

DROP TABLE IF EXISTS serving_player_game_batting_stats CASCADE;
DROP TABLE IF EXISTS serving_player_game_pitching_stats CASCADE;
DROP TABLE IF EXISTS serving_player_game_fielding_stats CASCADE;
DROP TABLE IF EXISTS serving_team_game_stats CASCADE;
DROP TABLE IF EXISTS serving_no_hitters CASCADE;
DROP TABLE IF EXISTS serving_cycles CASCADE;
DROP TABLE IF EXISTS serving_multi_hr_games CASCADE;
DROP TABLE IF EXISTS serving_triple_plays CASCADE;
DROP TABLE IF EXISTS serving_extra_inning_games CASCADE;
DROP TABLE IF EXISTS serving_season_batting_leaders CASCADE;
DROP TABLE IF EXISTS serving_season_pitching_leaders CASCADE;
DROP TABLE IF EXISTS serving_career_batting_leaders CASCADE;
DROP TABLE IF EXISTS serving_career_pitching_leaders CASCADE;
DROP TABLE IF EXISTS serving_win_expectancy_state_counts CASCADE;
DROP TABLE IF EXISTS serving_win_expectancy_historical CASCADE;
