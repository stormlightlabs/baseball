-- Calculate park factors over a range of seasons
WITH home_park_runs AS (
    SELECT
        park_id,
        home_team,
        SUBSTRING(date FROM 1 FOR 4)::int as season,
        COUNT(*) as games_home,
        AVG(home_score + visiting_score) as runs_per_game_home,
        AVG(home_homeruns + visiting_homeruns) as hr_per_game_home
    FROM games
    WHERE park_id = $1
        AND SUBSTRING(date FROM 1 FOR 4)::int BETWEEN $2 AND $3
    GROUP BY park_id, home_team, SUBSTRING(date FROM 1 FOR 4)::int
),
road_runs AS (
    SELECT
        visiting_team as team,
        SUBSTRING(date FROM 1 FOR 4)::int as season,
        AVG(home_score + visiting_score) as runs_per_game_road,
        AVG(home_homeruns + visiting_homeruns) as hr_per_game_road
    FROM games
    WHERE SUBSTRING(date FROM 1 FOR 4)::int BETWEEN $2 AND $3
    GROUP BY visiting_team, SUBSTRING(date FROM 1 FOR 4)::int
)
SELECT
    h.park_id,
    h.home_team,
    h.season,
    h.games_home,
    ROUND(100.0 * (h.runs_per_game_home / NULLIF(r.runs_per_game_road, 0)), 1) as runs_factor,
    ROUND(100.0 * (h.hr_per_game_home / NULLIF(r.hr_per_game_road, 0)), 1) as hr_factor
FROM home_park_runs h
JOIN road_runs r ON h.home_team = r.team AND h.season = r.season
ORDER BY h.season;
