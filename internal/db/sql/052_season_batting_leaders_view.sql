-- Create materialized view for season batting leaders
-- Combines Retrosheet per-game stats (1903-2025) with Lahman pre-1903 data (1871-1902)
-- Pre-aggregates all stats including advanced metrics (wOBA, wRC+)

CREATE MATERIALIZED VIEW season_batting_leaders AS
WITH retrosheet_batting AS (
    SELECT
        player_id,
        season,
        SUM(pa) as pa,
        SUM(ab) as ab,
        SUM(h) as h,
        SUM(doubles) as doubles,
        SUM(triples) as triples,
        SUM(hr) as hr,
        SUM(rbi) as rbi,
        SUM(sb) as sb,
        SUM(cs) as cs,
        SUM(bb) as bb,
        SUM(ibb) as ibb,
        SUM(so) as so,
        SUM(hbp) as hbp,
        SUM(sf) as sf,
        SUM(sh) as sh,
        SUM(gdp) as gdp,
        MAX(team_id) as team_id
    FROM player_game_batting_stats
    GROUP BY player_id, season
),
retrosheet_with_league AS (
    -- Add league information based on team_id
    SELECT
        rb.*,
        COALESCE(
            (SELECT DISTINCT home_team_league
             FROM games
             WHERE CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) = rb.season
               AND home_team = rb.team_id
             LIMIT 1),
            (SELECT DISTINCT visiting_team_league
             FROM games
             WHERE CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) = rb.season
               AND visiting_team = rb.team_id
             LIMIT 1)
        ) as league
    FROM retrosheet_batting rb
),
lahman_batting AS (
    SELECT
        "playerID" as player_id,
        "yearID" as season,
        SUM("AB" + "BB" + COALESCE("HBP", 0) + COALESCE("SF", 0)) as pa,
        SUM("AB") as ab,
        SUM("H") as h,
        SUM("2B") as doubles,
        SUM("3B") as triples,
        SUM("HR") as hr,
        SUM("RBI") as rbi,
        SUM(COALESCE("SB", 0)) as sb,
        SUM(COALESCE("CS", 0)) as cs,
        SUM("BB") as bb,
        SUM(COALESCE("IBB", 0)) as ibb,
        SUM(COALESCE("SO", 0)) as so,
        SUM(COALESCE("HBP", 0)) as hbp,
        SUM(COALESCE("SF", 0)) as sf,
        SUM(COALESCE("SH", 0)) as sh,
        SUM(COALESCE("GIDP", 0)) as gdp,
        MAX("teamID") as team_id,
        MAX("lgID") as league
    FROM "Batting"
    WHERE "yearID" < 1903
    GROUP BY "playerID", "yearID"
),
all_batting AS (
    SELECT * FROM retrosheet_with_league
    UNION ALL
    SELECT * FROM lahman_batting
),
stats_with_advanced AS (
    SELECT
        ab.*,
        CASE WHEN ab.ab > 0 THEN ROUND((ab.h::numeric / ab.ab), 3) ELSE 0 END as avg,
        CASE WHEN ab.pa > 0 THEN ROUND(((ab.h + ab.bb + ab.hbp)::numeric / ab.pa), 3) ELSE 0 END as obp,
        CASE WHEN ab.ab > 0 THEN ROUND(((ab.h + ab.doubles + 2*ab.triples + 3*ab.hr)::numeric / ab.ab), 3) ELSE 0 END as slg,
        CASE WHEN ab.ab > 0 THEN ROUND(((ab.h + ab.doubles + 2*ab.triples + 3*ab.hr)::numeric / ab.ab - ab.h::numeric / ab.ab), 3) ELSE 0 END as iso,
        CASE WHEN (ab.ab - ab.so - ab.hr + ab.sf) > 0 THEN ROUND(((ab.h - ab.hr)::numeric / (ab.ab - ab.so - ab.hr + ab.sf)), 3) ELSE 0 END as babip,
        CASE WHEN ab.pa > 0 THEN ROUND((ab.so::numeric / ab.pa), 3) ELSE 0 END as k_rate,
        CASE WHEN ab.pa > 0 THEN ROUND((ab.bb::numeric / ab.pa), 3) ELSE 0 END as bb_rate,
        CASE WHEN wc.season IS NOT NULL AND (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) > 0 THEN
            ROUND(
                (wc.w_bb * (ab.bb - ab.ibb) + wc.w_hbp * ab.hbp +
                 wc.w_1b * (ab.h - ab.doubles - ab.triples - ab.hr) +
                 wc.w_2b * ab.doubles + wc.w_3b * ab.triples + wc.w_hr * ab.hr)::numeric /
                (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp),
                3
            )
        ELSE NULL END as woba,
        CASE WHEN wc.season IS NOT NULL AND (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) > 0 THEN
            ROUND(
                (((wc.w_bb * (ab.bb - ab.ibb) + wc.w_hbp * ab.hbp +
                   wc.w_1b * (ab.h - ab.doubles - ab.triples - ab.hr) +
                   wc.w_2b * ab.doubles + wc.w_3b * ab.triples + wc.w_hr * ab.hr)::numeric /
                  (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) - wc.woba) /
                 wc.woba_scale * ab.pa)::numeric,
                2
            )
        ELSE NULL END as wraa,
        CASE WHEN wc.season IS NOT NULL AND lc.season IS NOT NULL AND (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) > 0 THEN
            ROUND(
                (((wc.w_bb * (ab.bb - ab.ibb) + wc.w_hbp * ab.hbp +
                   wc.w_1b * (ab.h - ab.doubles - ab.triples - ab.hr) +
                   wc.w_2b * ab.doubles + wc.w_3b * ab.triples + wc.w_hr * ab.hr)::numeric /
                  (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) - wc.woba) /
                 wc.woba_scale * ab.pa + (lc.wrc_per_pa * ab.pa))::numeric,
                2
            )
        ELSE NULL END as wrc,
        -- wRC+ (park adjusted)
        CASE WHEN lc.season IS NOT NULL AND lc.wrc_per_pa > 0 AND (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) > 0 THEN
            ROUND(
                ((((wc.w_bb * (ab.bb - ab.ibb) + wc.w_hbp * ab.hbp +
                    wc.w_1b * (ab.h - ab.doubles - ab.triples - ab.hr) +
                    wc.w_2b * ab.doubles + wc.w_3b * ab.triples + wc.w_hr * ab.hr)::numeric /
                   (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) - wc.woba) /
                  wc.woba_scale + lc.wrc_per_pa) +
                  (lc.wrc_per_pa - COALESCE(pf.basic_5yr, 100) / 100.0 * lc.wrc_per_pa)) /
                 lc.wrc_per_pa * 100,
                0
            )::int
        ELSE NULL END as wrc_plus
    FROM all_batting ab
    LEFT JOIN woba_constants wc ON wc.season = ab.season
    LEFT JOIN league_constants lc ON lc.season = ab.season AND lc.league = ab.league
    LEFT JOIN fangraphs_team_park_map pm ON pm.retrosheet_team_id = ab.team_id
        AND ab.season BETWEEN COALESCE(pm.start_year, 1871) AND COALESCE(pm.end_year, 2100)
    LEFT JOIN park_factors pf ON pf.park_id = pm.primary_park_id AND pf.season = ab.season
)
SELECT
    player_id,
    season,
    team_id,
    league,
    pa, ab, h, doubles, triples, hr, rbi, sb, cs, bb, ibb, so, hbp, sf, sh, gdp,
    avg, obp, slg, iso, babip, k_rate, bb_rate,
    woba, wraa, wrc, wrc_plus,
    -- Calculated OPS
    CASE WHEN obp IS NOT NULL AND slg IS NOT NULL THEN ROUND((obp + slg)::numeric, 3) ELSE NULL END as ops
FROM stats_with_advanced;

CREATE UNIQUE INDEX idx_season_batting_leaders_pk ON season_batting_leaders(player_id, season);

CREATE INDEX idx_season_batting_leaders_hr ON season_batting_leaders(season, hr DESC, h DESC) WHERE ab >= 300;
CREATE INDEX idx_season_batting_leaders_avg ON season_batting_leaders(season, avg DESC) WHERE ab >= 300;
CREATE INDEX idx_season_batting_leaders_rbi ON season_batting_leaders(season, rbi DESC) WHERE ab >= 300;
CREATE INDEX idx_season_batting_leaders_sb ON season_batting_leaders(season, sb DESC) WHERE ab >= 300;
CREATE INDEX idx_season_batting_leaders_h ON season_batting_leaders(season, h DESC) WHERE ab >= 300;
CREATE INDEX idx_season_batting_leaders_wrc_plus ON season_batting_leaders(season, wrc_plus DESC) WHERE pa >= 502;
CREATE INDEX idx_season_batting_leaders_woba ON season_batting_leaders(season, woba DESC) WHERE pa >= 502;
CREATE INDEX idx_season_batting_leaders_league_hr ON season_batting_leaders(season, league, hr DESC) WHERE ab >= 300;
CREATE INDEX idx_season_batting_leaders_season ON season_batting_leaders(season);
CREATE INDEX idx_season_batting_leaders_player ON season_batting_leaders(player_id, season DESC);

ANALYZE season_batting_leaders;
