-- Sample park factors for a few parks (Fenway, Coors, AT&T)

INSERT INTO park_factors (
    park_id, season, team_id,
    basic_5yr, basic_3yr, basic_1yr,
    factor_1b, factor_2b, factor_3b, factor_hr,
    factor_so, factor_bb, factor_gb, factor_fb,
    factor_ld, factor_iffb, factor_fip
) VALUES
    ('BOS07', 2023, 'BOS', 103, 104, 105, 98, 112, 95, 108, 97, 101, 95, 106, 101, 102, 101),
    ('BOS07', 2024, 'BOS', 104, 105, 106, 99, 113, 96, 109, 96, 100, 94, 107, 100, 103, 102),

    ('DEN02', 2023, 'COL', 115, 117, 119, 108, 115, 125, 105, 88, 103, 102, 114, 109, 95, 93),
    ('DEN02', 2024, 'COL', 116, 118, 120, 109, 116, 126, 106, 87, 104, 103, 115, 110, 96, 92),

    ('SFO03', 2023, 'SFN', 92, 91, 90, 101, 94, 88, 81, 103, 98, 104, 89, 97, 105, 107),
    ('SFO03', 2024, 'SFN', 91, 90, 89, 100, 93, 87, 80, 104, 99, 105, 88, 98, 106, 108)
ON CONFLICT (park_id, season) DO NOTHING;
