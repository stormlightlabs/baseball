-- Create materialized view for career pitching leaders
-- Aggregates from season_pitching_leaders to get career totals
-- Pre-calculates career rate stats

CREATE MATERIALIZED VIEW career_pitching_leaders AS
SELECT
    player_id,
    MAX(season) as last_season,
    SUM(w) as total_w,
    SUM(l) as total_l,
    SUM(sv) as total_sv,
    SUM(gs) as total_gs,
    SUM(cg) as total_cg,
    SUM(sho) as total_sho,
    SUM(g) as total_g,
    SUM(ipouts) as total_ipouts,
    SUM(h) as total_h,
    SUM(er) as total_er,
    SUM(hr) as total_hr,
    SUM(bb) as total_bb,
    SUM(so) as total_so,
    SUM(ibb) as total_ibb,
    SUM(hbp) as total_hbp,
    SUM(wp) as total_wp,
    SUM(bk) as total_bk,
    SUM(bfp) as total_bfp,
    ROUND((SUM(ipouts)::numeric / 3), 1) as career_ip,
    CASE WHEN SUM(ipouts) > 0 THEN ROUND((SUM(er)::numeric * 27.0 / SUM(ipouts)), 2) ELSE 0 END as career_era,
    CASE WHEN SUM(ipouts) > 0 THEN ROUND(((SUM(h) + SUM(bb))::numeric * 3 / SUM(ipouts)), 2) ELSE 0 END as career_whip,
    CASE WHEN SUM(ipouts) > 0 THEN ROUND((SUM(so)::numeric * 27.0 / SUM(ipouts)), 2) ELSE 0 END as career_k_per_9,
    CASE WHEN SUM(ipouts) > 0 THEN ROUND((SUM(bb)::numeric * 27.0 / SUM(ipouts)), 2) ELSE 0 END as career_bb_per_9,
    CASE WHEN SUM(ipouts) > 0 THEN ROUND((SUM(hr)::numeric * 27.0 / SUM(ipouts)), 2) ELSE 0 END as career_hr_per_9,
    -- TODO: use a computed constant
    CASE WHEN SUM(ipouts) > 0 THEN
        ROUND(
            (((13 * SUM(hr) + 3 * (SUM(bb) + SUM(hbp)) - 2 * SUM(so))::numeric / (SUM(ipouts) / 3.0)) + 3.2)::numeric,
            2
        )
    ELSE NULL END as career_fip,
    COUNT(*) as seasons_pitched
FROM season_pitching_leaders
GROUP BY player_id;

CREATE UNIQUE INDEX idx_career_pitching_leaders_pk ON career_pitching_leaders(player_id);

CREATE INDEX idx_career_pitching_leaders_w ON career_pitching_leaders(total_w DESC) WHERE total_ipouts >= 1500;
CREATE INDEX idx_career_pitching_leaders_so ON career_pitching_leaders(total_so DESC) WHERE total_ipouts >= 1500;
CREATE INDEX idx_career_pitching_leaders_sv ON career_pitching_leaders(total_sv DESC) WHERE total_g >= 100;
CREATE INDEX idx_career_pitching_leaders_era ON career_pitching_leaders(career_era ASC) WHERE total_ipouts >= 1500;
CREATE INDEX idx_career_pitching_leaders_whip ON career_pitching_leaders(career_whip ASC) WHERE total_ipouts >= 1500;

ANALYZE career_pitching_leaders;
