SELECT
    u.retro_id,
    u.first_name,
    u.last_name,
    u.first_game,
    u.last_game
FROM umpires u
WHERE u.retro_id = $1
