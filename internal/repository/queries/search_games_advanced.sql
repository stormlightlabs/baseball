-- Advanced game search function with structured filters and natural language fallback
-- Parameters: query text, season, home_team, away_team, series_id, game_number, limit
WITH filtered_games AS (
    SELECT
        game_id,
        date,
        home_team,
        visiting_team,
        home_score,
        visiting_score,
        game_type,
        park_id,
        search_text,
        search_tsv,
        -- Ranking score for text search
        CASE
            WHEN $1 = '' OR $1 IS NULL THEN 0
            ELSE ts_rank(search_tsv, plainto_tsquery('english', $1))
        END as text_rank
    FROM games
    WHERE
        -- Season filter
        ($2::int IS NULL OR EXTRACT(YEAR FROM TO_DATE(date, 'YYYYMMDD'))::int = $2) AND
        -- Home team filter (allows matching either home or away)
        ($3::varchar IS NULL OR home_team = $3 OR visiting_team = $3) AND
        -- Away team filter
        ($4::varchar IS NULL OR visiting_team = $4 OR home_team = $4) AND
        -- Game type/series filter
        ($5::varchar IS NULL OR game_type = $5) AND
        -- Game number filter (for postseason series)
        ($6::int IS NULL OR game_number = $6) AND
        -- Text search (full-text or fuzzy fallback)
        (
            $1 = '' OR $1 IS NULL OR
            search_tsv @@ plainto_tsquery('english', $1) OR
            search_text ILIKE '%' || $1 || '%'
        )
)
SELECT
    game_id,
    date,
    home_team,
    visiting_team,
    home_score,
    visiting_score,
    game_type,
    park_id
FROM filtered_games
ORDER BY
    -- Prioritize structured filter matches over text matches
    CASE WHEN $2 IS NOT NULL OR $3 IS NOT NULL OR $4 IS NOT NULL THEN 1 ELSE 2 END,
    text_rank DESC,
    date DESC
LIMIT $7;
