-- Fix season batting leaders duplication from overlapping park-map windows.
--
-- Some teams (e.g., CLE in 2021) have overlapping entries in fangraphs_team_park_map.
-- The prior LEFT JOIN could multiply rows for a single (player_id, season), causing
-- unique index build failures when refreshing season_batting_leaders.
--
-- This migration rewrites season_batting_leaders to use a deterministic single-row
-- park map lookup (LATERAL ... LIMIT 1), then recreates dependent career_batting_leaders.
-- Both views remain WITH NO DATA and are populated by explicit refresh steps.

DROP MATERIALIZED VIEW IF EXISTS career_batting_leaders;
DROP MATERIALIZED VIEW IF EXISTS season_batting_leaders;

CREATE MATERIALIZED VIEW season_batting_leaders AS
WITH retrosheet_batting AS (
    SELECT
        player_id,
        season,
        SUM(pa) AS pa,
        SUM(ab) AS ab,
        SUM(h) AS h,
        SUM(doubles) AS doubles,
        SUM(triples) AS triples,
        SUM(hr) AS hr,
        SUM(rbi) AS rbi,
        SUM(sb) AS sb,
        SUM(cs) AS cs,
        SUM(bb) AS bb,
        SUM(ibb) AS ibb,
        SUM(so) AS so,
        SUM(hbp) AS hbp,
        SUM(sf) AS sf,
        SUM(sh) AS sh,
        SUM(gdp) AS gdp,
        MAX(team_id) AS team_id
    FROM player_game_batting_stats
    GROUP BY player_id, season
),
retrosheet_with_league AS (
    SELECT
        rb.*,
        COALESCE(
            (
                SELECT DISTINCT home_team_league
                FROM games
                WHERE CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) = rb.season
                  AND home_team = rb.team_id
                LIMIT 1
            ),
            (
                SELECT DISTINCT visiting_team_league
                FROM games
                WHERE CAST(SUBSTRING(game_id, 4, 4) AS INTEGER) = rb.season
                  AND visiting_team = rb.team_id
                LIMIT 1
            )
        ) AS league
    FROM retrosheet_batting rb
),
lahman_batting AS (
    SELECT
        "playerID" AS player_id,
        "yearID" AS season,
        SUM("AB" + "BB" + COALESCE("HBP", 0) + COALESCE("SF", 0)) AS pa,
        SUM("AB") AS ab,
        SUM("H") AS h,
        SUM("2B") AS doubles,
        SUM("3B") AS triples,
        SUM("HR") AS hr,
        SUM("RBI") AS rbi,
        SUM(COALESCE("SB", 0)) AS sb,
        SUM(COALESCE("CS", 0)) AS cs,
        SUM("BB") AS bb,
        SUM(COALESCE("IBB", 0)) AS ibb,
        SUM(COALESCE("SO", 0)) AS so,
        SUM(COALESCE("HBP", 0)) AS hbp,
        SUM(COALESCE("SF", 0)) AS sf,
        SUM(COALESCE("SH", 0)) AS sh,
        SUM(COALESCE("GIDP", 0)) AS gdp,
        MAX("teamID") AS team_id,
        MAX("lgID") AS league
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
        CASE WHEN ab.ab > 0 THEN ROUND((ab.h::numeric / ab.ab), 3) ELSE 0 END AS avg,
        CASE WHEN ab.pa > 0 THEN ROUND(((ab.h + ab.bb + ab.hbp)::numeric / ab.pa), 3) ELSE 0 END AS obp,
        CASE WHEN ab.ab > 0 THEN ROUND(((ab.h + ab.doubles + 2 * ab.triples + 3 * ab.hr)::numeric / ab.ab), 3) ELSE 0 END AS slg,
        CASE WHEN ab.ab > 0 THEN ROUND(((ab.h + ab.doubles + 2 * ab.triples + 3 * ab.hr)::numeric / ab.ab - ab.h::numeric / ab.ab), 3) ELSE 0 END AS iso,
        CASE WHEN (ab.ab - ab.so - ab.hr + ab.sf) > 0 THEN ROUND(((ab.h - ab.hr)::numeric / (ab.ab - ab.so - ab.hr + ab.sf)), 3) ELSE 0 END AS babip,
        CASE WHEN ab.pa > 0 THEN ROUND((ab.so::numeric / ab.pa), 3) ELSE 0 END AS k_rate,
        CASE WHEN ab.pa > 0 THEN ROUND((ab.bb::numeric / ab.pa), 3) ELSE 0 END AS bb_rate,
        CASE
            WHEN wc.season IS NOT NULL AND (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) > 0
            THEN ROUND(
                (
                    wc.w_bb * (ab.bb - ab.ibb) + wc.w_hbp * ab.hbp +
                    wc.w_1b * (ab.h - ab.doubles - ab.triples - ab.hr) +
                    wc.w_2b * ab.doubles + wc.w_3b * ab.triples + wc.w_hr * ab.hr
                )::numeric /
                (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp),
                3
            )
            ELSE NULL
        END AS woba,
        CASE
            WHEN wc.season IS NOT NULL AND (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) > 0
            THEN ROUND(
                (
                    (
                        (
                            wc.w_bb * (ab.bb - ab.ibb) + wc.w_hbp * ab.hbp +
                            wc.w_1b * (ab.h - ab.doubles - ab.triples - ab.hr) +
                            wc.w_2b * ab.doubles + wc.w_3b * ab.triples + wc.w_hr * ab.hr
                        )::numeric /
                        (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) - wc.woba
                    ) /
                    wc.woba_scale * ab.pa
                )::numeric,
                2
            )
            ELSE NULL
        END AS wraa,
        CASE
            WHEN wc.season IS NOT NULL AND lc.season IS NOT NULL AND (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) > 0
            THEN ROUND(
                (
                    (
                        (
                            wc.w_bb * (ab.bb - ab.ibb) + wc.w_hbp * ab.hbp +
                            wc.w_1b * (ab.h - ab.doubles - ab.triples - ab.hr) +
                            wc.w_2b * ab.doubles + wc.w_3b * ab.triples + wc.w_hr * ab.hr
                        )::numeric /
                        (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) - wc.woba
                    ) /
                    wc.woba_scale * ab.pa + (lc.wrc_per_pa * ab.pa)
                )::numeric,
                2
            )
            ELSE NULL
        END AS wrc,
        CASE
            WHEN lc.season IS NOT NULL AND lc.wrc_per_pa > 0 AND (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) > 0
            THEN ROUND(
                (
                    (
                        (
                            (
                                wc.w_bb * (ab.bb - ab.ibb) + wc.w_hbp * ab.hbp +
                                wc.w_1b * (ab.h - ab.doubles - ab.triples - ab.hr) +
                                wc.w_2b * ab.doubles + wc.w_3b * ab.triples + wc.w_hr * ab.hr
                            )::numeric /
                            (ab.ab + ab.bb - ab.ibb + ab.sf + ab.hbp) - wc.woba
                        ) /
                        wc.woba_scale + lc.wrc_per_pa
                    ) +
                    (
                        lc.wrc_per_pa - COALESCE(pf.basic_5yr, 100) / 100.0 * lc.wrc_per_pa
                    )
                ) /
                lc.wrc_per_pa * 100,
                0
            )::int
            ELSE NULL
        END AS wrc_plus
    FROM all_batting ab
    LEFT JOIN woba_constants wc ON wc.season = ab.season
    LEFT JOIN league_constants lc ON lc.season = ab.season AND lc.league = ab.league
    LEFT JOIN LATERAL (
        SELECT pm.primary_park_id
        FROM fangraphs_team_park_map pm
        WHERE pm.retrosheet_team_id = ab.team_id
          AND ab.season BETWEEN COALESCE(pm.start_year, 1871) AND COALESCE(pm.end_year, 2100)
        ORDER BY
            (COALESCE(pm.end_year, 2100) - COALESCE(pm.start_year, 1871)) ASC,
            COALESCE(pm.end_year, 2100) DESC,
            COALESCE(pm.start_year, 1871) DESC,
            pm.primary_park_id ASC
        LIMIT 1
    ) pm ON TRUE
    LEFT JOIN park_factors pf ON pf.park_id = pm.primary_park_id AND pf.season = ab.season
)
SELECT
    player_id,
    season,
    team_id,
    league,
    pa,
    ab,
    h,
    doubles,
    triples,
    hr,
    rbi,
    sb,
    cs,
    bb,
    ibb,
    so,
    hbp,
    sf,
    sh,
    gdp,
    avg,
    obp,
    slg,
    iso,
    babip,
    k_rate,
    bb_rate,
    woba,
    wraa,
    wrc,
    wrc_plus,
    CASE WHEN obp IS NOT NULL AND slg IS NOT NULL THEN ROUND((obp + slg)::numeric, 3) ELSE NULL END AS ops
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

CREATE MATERIALIZED VIEW career_batting_leaders AS
SELECT
    player_id,
    MAX(season) AS last_season,
    SUM(pa) AS total_pa,
    SUM(ab) AS total_ab,
    SUM(h) AS total_h,
    SUM(doubles) AS total_doubles,
    SUM(triples) AS total_triples,
    SUM(hr) AS total_hr,
    SUM(rbi) AS total_rbi,
    SUM(sb) AS total_sb,
    SUM(cs) AS total_cs,
    SUM(bb) AS total_bb,
    SUM(ibb) AS total_ibb,
    SUM(so) AS total_so,
    SUM(hbp) AS total_hbp,
    SUM(sf) AS total_sf,
    SUM(sh) AS total_sh,
    SUM(gdp) AS total_gdp,
    CASE WHEN SUM(ab) > 0 THEN ROUND((SUM(h)::numeric / SUM(ab)), 3) ELSE 0 END AS career_avg,
    CASE WHEN SUM(pa) > 0 THEN ROUND(((SUM(h) + SUM(bb) + SUM(hbp))::numeric / SUM(pa)), 3) ELSE 0 END AS career_obp,
    CASE WHEN SUM(ab) > 0 THEN ROUND(((SUM(h) + SUM(doubles) + 2 * SUM(triples) + 3 * SUM(hr))::numeric / SUM(ab)), 3) ELSE 0 END AS career_slg,
    CASE
        WHEN SUM(pa) > 0 AND SUM(ab) > 0
        THEN ROUND(
            (
                ((SUM(h) + SUM(bb) + SUM(hbp))::numeric / SUM(pa)) +
                ((SUM(h) + SUM(doubles) + 2 * SUM(triples) + 3 * SUM(hr))::numeric / SUM(ab))
            ),
            3
        )
        ELSE 0
    END AS career_ops,
    COUNT(*) AS seasons_played
FROM season_batting_leaders
GROUP BY player_id
WITH NO DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_career_batting_leaders_pk ON career_batting_leaders(player_id);
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_hr ON career_batting_leaders(total_hr DESC, total_h DESC) WHERE total_ab >= 1000;
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_h ON career_batting_leaders(total_h DESC) WHERE total_ab >= 1000;
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_rbi ON career_batting_leaders(total_rbi DESC) WHERE total_ab >= 1000;
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_avg ON career_batting_leaders(career_avg DESC) WHERE total_ab >= 1000;
CREATE INDEX IF NOT EXISTS idx_career_batting_leaders_ops ON career_batting_leaders(career_ops DESC) WHERE total_pa >= 3000;
