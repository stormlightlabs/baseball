-- Calculate park factor for a specific park and season
WITH home_park_runs AS (
    SELECT
        park_id,
        home_team,
        COUNT(*) as games_home,
        SUM(home_score + visiting_score) as runs_home,
        SUM(home_homeruns + visiting_homeruns) as hr_home,
        AVG(home_score + visiting_score) as runs_per_game_home,
        AVG(home_homeruns + visiting_homeruns) as hr_per_game_home
    FROM games
    WHERE park_id = $1
        AND SUBSTRING(date FROM 1 FOR 4)::int = $2
    GROUP BY park_id, home_team
),
road_runs AS (
    SELECT
        visiting_team as team,
        COUNT(*) as games_road,
        AVG(home_score + visiting_score) as runs_per_game_road,
        AVG(home_homeruns + visiting_homeruns) as hr_per_game_road
    FROM games
    WHERE visiting_team IN (SELECT home_team FROM home_park_runs)
        AND SUBSTRING(date FROM 1 FOR 4)::int = $2
    GROUP BY visiting_team
)
SELECT
    h.park_id,
    h.home_team,
    $2 as season,
    h.games_home,
    -- Runs park factor (scaled to 100)
    ROUND(100.0 * (h.runs_per_game_home / NULLIF(r.runs_per_game_road, 0)), 1) as runs_factor,
    -- HR park factor (scaled to 100)
    ROUND(100.0 * (h.hr_per_game_home / NULLIF(r.hr_per_game_road, 0)), 1) as hr_factor
FROM home_park_runs h
JOIN road_runs r ON h.home_team = r.team;
