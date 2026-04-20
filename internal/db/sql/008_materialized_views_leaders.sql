-- Fresh migration set: season/career leaderboards materialized views.
-- Views are created WITH NO DATA and refreshed via ETL/db refresh-views.


-- SECTION 036_season_batting_leaders_view.sql

-- Create materialized view for season batting leaders
-- Combines Retrosheet per-game stats (1903-2025) with Lahman pre-1903 data (1871-1902)
-- Pre-aggregates all stats including advanced metrics (wOBA, wRC+)

CREATE MATERIALIZED VIEW IF NOT EXISTS season_batting_leaders AS
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
FROM stats_with_advanced
WITH NO DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_season_batting_leaders_pk ON season_batting_leaders(player_id, season);

CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_hr ON season_batting_leaders(season, hr DESC, h DESC) WHERE ab >= 300;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_avg ON season_batting_leaders(season, avg DESC) WHERE ab >= 300;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_rbi ON season_batting_leaders(season, rbi DESC) WHERE ab >= 300;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_sb ON season_batting_leaders(season, sb DESC) WHERE ab >= 300;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_h ON season_batting_leaders(season, h DESC) WHERE ab >= 300;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_wrc_plus ON season_batting_leaders(season, wrc_plus DESC) WHERE pa >= 502;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_woba ON season_batting_leaders(season, woba DESC) WHERE pa >= 502;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_league_hr ON season_batting_leaders(season, league, hr DESC) WHERE ab >= 300;
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_season ON season_batting_leaders(season);
CREATE INDEX IF NOT EXISTS idx_season_batting_leaders_player ON season_batting_leaders(player_id, season DESC);



-- SECTION 037_season_pitching_leaders_view.sql

-- Create materialized view for season pitching leaders
-- Combines Retrosheet per-game stats (1903-2025) with W/L/SV from games table
-- Includes Lahman pre-1903 data (1871-1902)
-- Pre-aggregates all stats including advanced metrics (FIP, WHIP, K/9)

CREATE MATERIALIZED VIEW IF NOT EXISTS season_pitching_leaders AS
WITH retrosheet_pitching AS (
    SELECT
        player_id,
        season,
        COUNT(*) as g,  -- games appeared
        SUM(ip * 3) as ipouts,  -- convert IP back to outs
        SUM(h) as h,
        SUM(er) as er,
        SUM(hr) as hr,
        SUM(bb) as bb,
        SUM(so) as so,
        SUM(ibb) as ibb,
        SUM(hbp) as hbp,
        SUM(wp) as wp,
        SUM(bk) as bk,
        SUM(pa) as bfp,
        MAX(team_id) as team_id
    FROM player_game_pitching_stats
    GROUP BY player_id, season
),
retrosheet_with_league AS (
    SELECT
        rp.*,
        (
            SELECT DISTINCT home_team_league
            FROM games
            WHERE CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) = rp.season
              AND (home_team = rp.team_id OR visiting_team = rp.team_id)
            LIMIT 1
        ) as league
    FROM retrosheet_pitching rp
),
pitcher_decisions AS (
    SELECT
        season,
        player_id,
        COUNT(*) FILTER (WHERE decision = 'W') as w,
        COUNT(*) FILTER (WHERE decision = 'L') as l,
        COUNT(*) FILTER (WHERE decision = 'SV') as sv,
        COUNT(*) FILTER (WHERE decision = 'GS') as gs
    FROM (
        SELECT
            CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) as season,
            winning_pitcher_id as player_id,
            'W' as decision
        FROM games
        WHERE winning_pitcher_id IS NOT NULL AND winning_pitcher_id != ''

        UNION ALL

        SELECT
            CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) as season,
            losing_pitcher_id as player_id,
            'L' as decision
        FROM games
        WHERE losing_pitcher_id IS NOT NULL AND losing_pitcher_id != ''

        UNION ALL

        SELECT
            CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) as season,
            saving_pitcher_id as player_id,
            'SV' as decision
        FROM games
        WHERE saving_pitcher_id IS NOT NULL AND saving_pitcher_id != ''

        UNION ALL

        SELECT
            CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) as season,
            h_starting_pitcher_id as player_id,
            'GS' as decision
        FROM games
        WHERE h_starting_pitcher_id IS NOT NULL AND h_starting_pitcher_id != ''

        UNION ALL

        SELECT
            CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) as season,
            v_starting_pitcher_id as player_id,
            'GS' as decision
        FROM games
        WHERE v_starting_pitcher_id IS NOT NULL AND v_starting_pitcher_id != ''
    ) decisions
    GROUP BY season, player_id
),
retrosheet_combined AS (
    SELECT
        rp.player_id,
        rp.season,
        COALESCE(pd.w, 0) as w,
        COALESCE(pd.l, 0) as l,
        COALESCE(pd.sv, 0) as sv,
        COALESCE(pd.gs, 0) as gs,
        0 as cg,
        0 as sho,
        rp.g,
        rp.ipouts,
        rp.h,
        rp.er,
        rp.hr,
        rp.bb,
        rp.so,
        rp.ibb,
        rp.hbp,
        rp.wp,
        rp.bk,
        rp.bfp,
        rp.team_id,
        rp.league
    FROM retrosheet_with_league rp
    LEFT JOIN pitcher_decisions pd ON pd.player_id = rp.player_id AND pd.season = rp.season
),
lahman_pitching AS (
    SELECT
        "playerID" as player_id,
        "yearID" as season,
        SUM("W") as w,
        SUM("L") as l,
        SUM("SV") as sv,
        SUM("GS") as gs,
        SUM("CG") as cg,
        SUM("SHO") as sho,
        SUM("G") as g,
        SUM("IPouts") as ipouts,
        SUM("H") as h,
        SUM("ER") as er,
        SUM("HR") as hr,
        SUM("BB") as bb,
        SUM("SO") as so,
        SUM(COALESCE("IBB", 0)) as ibb,
        SUM(COALESCE("HBP", 0)) as hbp,
        SUM(COALESCE("WP", 0)) as wp,
        SUM(COALESCE("BK", 0)) as bk,
        SUM(COALESCE("BFP", 0)) as bfp,
        MAX("teamID") as team_id,
        MAX("lgID") as league
    FROM "Pitching"
    WHERE "yearID" < 1903
    GROUP BY "playerID", "yearID"
),
all_pitching AS (
    SELECT * FROM retrosheet_combined
    UNION ALL
    SELECT * FROM lahman_pitching
)
SELECT
    ap.*,
    ROUND((ap.ipouts::numeric / 3), 1) as ip,
    CASE WHEN ap.ipouts > 0 THEN ROUND((ap.er::numeric * 27.0 / ap.ipouts), 2) ELSE 0 END as era,
    CASE WHEN ap.ipouts > 0 THEN ROUND(((ap.h + ap.bb)::numeric * 3 / ap.ipouts), 2) ELSE 0 END as whip,
    CASE WHEN ap.ipouts > 0 THEN ROUND((ap.so::numeric * 27.0 / ap.ipouts), 2) ELSE 0 END as k_per_9,
    CASE WHEN ap.ipouts > 0 THEN ROUND((ap.bb::numeric * 27.0 / ap.ipouts), 2) ELSE 0 END as bb_per_9,
    CASE WHEN ap.ipouts > 0 THEN ROUND((ap.hr::numeric * 27.0 / ap.ipouts), 2) ELSE 0 END as hr_per_9,
    CASE WHEN ap.ipouts > 0 THEN
        ROUND(
            (((13 * ap.hr + 3 * (ap.bb + ap.hbp) - 2 * ap.so)::numeric / (ap.ipouts / 3.0)) +
             COALESCE(wc.c_fip, 3.2))::numeric,
            2
        )
    ELSE NULL END as fip
FROM all_pitching ap
LEFT JOIN woba_constants wc ON wc.season = ap.season
WITH NO DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_season_pitching_leaders_pk ON season_pitching_leaders(player_id, season);

CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_era ON season_pitching_leaders(season, era ASC) WHERE ipouts >= 450;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_so ON season_pitching_leaders(season, so DESC) WHERE ipouts >= 450;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_w ON season_pitching_leaders(season, w DESC) WHERE ipouts >= 450;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_sv ON season_pitching_leaders(season, sv DESC) WHERE g >= 20;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_fip ON season_pitching_leaders(season, fip ASC) WHERE ipouts >= 450;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_whip ON season_pitching_leaders(season, whip ASC) WHERE ipouts >= 450;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_league_era ON season_pitching_leaders(season, league, era ASC) WHERE ipouts >= 450;
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_season ON season_pitching_leaders(season);
CREATE INDEX IF NOT EXISTS idx_season_pitching_leaders_player ON season_pitching_leaders(player_id, season DESC);



-- SECTION 038_career_batting_leaders_view.sql

-- Create materialized view for career batting leaders
-- Aggregates from season_batting_leaders to get career totals
-- Pre-calculates career rate stats

CREATE MATERIALIZED VIEW IF NOT EXISTS career_batting_leaders AS
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
GROUP BY player_id
WITH NO DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_career_batting_leaders_pk ON career_batting_leaders(player_id);

CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_hr ON career_batting_leaders(total_hr DESC, total_h DESC) WHERE total_ab >= 1000;
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_h ON career_batting_leaders(total_h DESC) WHERE total_ab >= 1000;
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_rbi ON career_batting_leaders(total_rbi DESC) WHERE total_ab >= 1000;
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_avg ON career_batting_leaders(career_avg DESC) WHERE total_ab >= 1000;
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_ops ON career_batting_leaders(career_ops DESC) WHERE total_pa >= 3000;



-- SECTION 039_career_pitching_leaders_view.sql

-- Create materialized view for career pitching leaders
-- Aggregates from season_pitching_leaders to get career totals
-- Pre-calculates career rate stats

CREATE MATERIALIZED VIEW IF NOT EXISTS career_pitching_leaders AS
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
GROUP BY player_id
WITH NO DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_career_pitching_leaders_pk ON career_pitching_leaders(player_id);

CREATE INDEX IF NOT EXISTS idx_career_pitching_leaders_w ON career_pitching_leaders(total_w DESC) WHERE total_ipouts >= 1500;
CREATE INDEX IF NOT EXISTS idx_career_pitching_leaders_so ON career_pitching_leaders(total_so DESC) WHERE total_ipouts >= 1500;
CREATE INDEX IF NOT EXISTS idx_career_pitching_leaders_sv ON career_pitching_leaders(total_sv DESC) WHERE total_g >= 100;
CREATE INDEX IF NOT EXISTS idx_career_pitching_leaders_era ON career_pitching_leaders(career_era ASC) WHERE total_ipouts >= 1500;
CREATE INDEX IF NOT EXISTS idx_career_pitching_leaders_whip ON career_pitching_leaders(career_whip ASC) WHERE total_ipouts >= 1500;

