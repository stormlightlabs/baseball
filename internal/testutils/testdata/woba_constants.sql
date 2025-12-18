-- Sample data for 2023 and 2024 seasons based on FanGraphs values
INSERT INTO woba_constants (
    season, w_bb, w_hbp, w_1b, w_2b, w_3b, w_hr,
    woba_scale, woba, run_sb, run_cs, r_pa, r_w, c_fip
) VALUES
    (2023, 0.690, 0.720, 0.880, 1.247, 1.578, 2.004,
     1.185, 0.313, 0.200, -0.420, 0.119, 9.80, 3.132),
    (2024, 0.692, 0.723, 0.883, 1.252, 1.584, 2.011,
     1.190, 0.315, 0.202, -0.422, 0.121, 9.85, 3.140)
ON CONFLICT (season) DO NOTHING;
