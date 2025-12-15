-- Calculate season run differential with per-game details
-- Parameters: $1 = team_id, $2 = season year (YYYY format as int)

WITH team_games AS (
    SELECT
        game_id,
        date,
        game_number,
        CASE WHEN home_team = $1 THEN 1 ELSE 0 END as is_home,
        CASE
            WHEN home_team = $1 THEN visiting_team
            ELSE home_team
        END as opponent_id,
        CASE
            WHEN home_team = $1 THEN home_score
            ELSE visiting_score
        END as runs_scored,
        CASE
            WHEN home_team = $1 THEN visiting_score
            ELSE home_score
        END as runs_allowed,
        ROW_NUMBER() OVER (ORDER BY date, game_number) as game_num
    FROM games
    WHERE (home_team = $1 OR visiting_team = $1)
      AND EXTRACT(YEAR FROM TO_DATE(date, 'YYYYMMDD')) = $2
      AND game_type = 'regular'
)
SELECT
    game_id,
    date,
    opponent_id,
    is_home,
    runs_scored,
    runs_allowed,
    runs_scored - runs_allowed as differential,
    SUM(runs_scored - runs_allowed) OVER (
        ORDER BY date, game_number
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) as cumulative_diff,
    game_num
FROM team_games
ORDER BY date, game_number
