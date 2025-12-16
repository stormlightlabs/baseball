-- Create materialized view for season pitching leaders
-- Combines Retrosheet per-game stats (1903-2025) with W/L/SV from games table
-- Includes Lahman pre-1903 data (1871-1902)
-- Pre-aggregates all stats including advanced metrics (FIP, WHIP, K/9)

CREATE MATERIALIZED VIEW season_pitching_leaders AS
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
LEFT JOIN woba_constants wc ON wc.season = ap.season;

CREATE UNIQUE INDEX idx_season_pitching_leaders_pk ON season_pitching_leaders(player_id, season);

CREATE INDEX idx_season_pitching_leaders_era ON season_pitching_leaders(season, era ASC) WHERE ipouts >= 450;
CREATE INDEX idx_season_pitching_leaders_so ON season_pitching_leaders(season, so DESC) WHERE ipouts >= 450;
CREATE INDEX idx_season_pitching_leaders_w ON season_pitching_leaders(season, w DESC) WHERE ipouts >= 450;
CREATE INDEX idx_season_pitching_leaders_sv ON season_pitching_leaders(season, sv DESC) WHERE g >= 20;
CREATE INDEX idx_season_pitching_leaders_fip ON season_pitching_leaders(season, fip ASC) WHERE ipouts >= 450;
CREATE INDEX idx_season_pitching_leaders_whip ON season_pitching_leaders(season, whip ASC) WHERE ipouts >= 450;
CREATE INDEX idx_season_pitching_leaders_league_era ON season_pitching_leaders(season, league, era ASC) WHERE ipouts >= 450;
CREATE INDEX idx_season_pitching_leaders_season ON season_pitching_leaders(season);
CREATE INDEX idx_season_pitching_leaders_player ON season_pitching_leaders(player_id, season DESC);

ANALYZE season_pitching_leaders;
