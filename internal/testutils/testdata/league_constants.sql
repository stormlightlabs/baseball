-- Sample data for 2023 and 2024 AL/NL seasons
INSERT INTO league_constants (
    season, league, woba_avg, wrc_per_pa, runs_per_win,
    replacement_runs_per_pa, total_pa, total_runs
) VALUES
    (2023, 'AL', 0.314, 0.1195, 9.80, -0.030, 98234, 11745),
    (2023, 'NL', 0.312, 0.1185, 9.80, -0.030, 97856, 11589),
    (2024, 'AL', 0.316, 0.1215, 9.85, -0.031, 99012, 12034),
    (2024, 'NL', 0.314, 0.1205, 9.85, -0.031, 98456, 11867)
ON CONFLICT (season, league) DO NOTHING;
