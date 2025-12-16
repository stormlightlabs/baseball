-- Create materialized view for career batting leaders
-- Aggregates from season_batting_leaders to get career totals
-- Pre-calculates career rate stats

CREATE MATERIALIZED VIEW career_batting_leaders AS
SELECT
    player_id,
    MAX(season) as last_season,
    SUM(pa) as total_pa,
    SUM(ab) as total_ab,
    SUM(h) as total_h,
    SUM(doubles) as total_doubles,
    SUM(triples) as total_triples,
    SUM(hr) as total_hr,
    SUM(rbi) as total_rbi,
    SUM(sb) as total_sb,
    SUM(cs) as total_cs,
    SUM(bb) as total_bb,
    SUM(ibb) as total_ibb,
    SUM(so) as total_so,
    SUM(hbp) as total_hbp,
    SUM(sf) as total_sf,
    SUM(sh) as total_sh,
    SUM(gdp) as total_gdp,
    CASE WHEN SUM(ab) > 0 THEN ROUND((SUM(h)::numeric / SUM(ab)), 3) ELSE 0 END as career_avg,
    CASE WHEN SUM(pa) > 0 THEN ROUND(((SUM(h) + SUM(bb) + SUM(hbp))::numeric / SUM(pa)), 3) ELSE 0 END as career_obp,
    CASE WHEN SUM(ab) > 0 THEN ROUND(((SUM(h) + SUM(doubles) + 2*SUM(triples) + 3*SUM(hr))::numeric / SUM(ab)), 3) ELSE 0 END as career_slg,
    -- Career OPS
    CASE WHEN SUM(pa) > 0 AND SUM(ab) > 0 THEN
        ROUND((
            ((SUM(h) + SUM(bb) + SUM(hbp))::numeric / SUM(pa)) +
            ((SUM(h) + SUM(doubles) + 2*SUM(triples) + 3*SUM(hr))::numeric / SUM(ab))
        ), 3)
    ELSE 0 END as career_ops,
    COUNT(*) as seasons_played
FROM season_batting_leaders
GROUP BY player_id;

CREATE UNIQUE INDEX idx_career_batting_leaders_pk ON career_batting_leaders(player_id);

CREATE INDEX idx_career_batting_leaders_hr ON career_batting_leaders(total_hr DESC, total_h DESC) WHERE total_ab >= 1000;
CREATE INDEX idx_career_batting_leaders_h ON career_batting_leaders(total_h DESC) WHERE total_ab >= 1000;
CREATE INDEX idx_career_batting_leaders_rbi ON career_batting_leaders(total_rbi DESC) WHERE total_ab >= 1000;
CREATE INDEX idx_career_batting_leaders_avg ON career_batting_leaders(career_avg DESC) WHERE total_ab >= 1000;
CREATE INDEX idx_career_batting_leaders_ops ON career_batting_leaders(career_ops DESC) WHERE total_pa >= 3000;

ANALYZE career_batting_leaders;
