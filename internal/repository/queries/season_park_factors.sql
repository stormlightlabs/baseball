-- Calculate all park factors for a given season
WITH home_park_runs AS (
    SELECT
        park_id,
        home_team,
        COUNT(*) as games_home,
        AVG(home_score + visiting_score) as runs_per_game_home,
        AVG(home_homeruns + visiting_homeruns) as hr_per_game_home
    FROM games
    WHERE SUBSTRING(date FROM 1 FOR 4)::int = $1
    GROUP BY park_id, home_team
),
road_runs AS (
    SELECT
        visiting_team as team,
        AVG(home_score + visiting_score) as runs_per_game_road,
        AVG(home_homeruns + visiting_homeruns) as hr_per_game_road
    FROM games
    WHERE SUBSTRING(date FROM 1 FOR 4)::int = $1
    GROUP BY visiting_team
)
SELECT
    h.park_id,
    h.home_team,
    SUBSTRING(date FROM 1 FOR 4)::int as season,
    h.games_home,
    ROUND(100.0 * (h.runs_per_game_home / NULLIF(r.runs_per_game_road, 0)), 1) as runs_factor,
    ROUND(100.0 * (h.hr_per_game_home / NULLIF(r.hr_per_game_road, 0)), 1) as hr_factor
FROM home_park_runs h
JOIN road_runs r ON h.home_team = r.team
JOIN games g ON g.park_id = h.park_id
WHERE h.games_home >= 40  -- Only parks with sufficient sample size
    AND SUBSTRING(g.date FROM 1 FOR 4)::int = $1
GROUP BY h.park_id, h.home_team, SUBSTRING(g.date FROM 1 FOR 4)::int, h.games_home, h.runs_per_game_home, r.runs_per_game_road, h.hr_per_game_home, r.hr_per_game_road
ORDER BY runs_factor DESC;
