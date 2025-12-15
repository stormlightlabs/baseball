-- Search games with series game number support
-- For playoff series, ranks games chronologically and filters by game number in series
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
        -- Rank games within series by date for playoff series game numbering
        ROW_NUMBER() OVER (
            PARTITION BY
                CASE WHEN game_type IN ('worldseries', 'lcs', 'divisionseries', 'wildcard')
                THEN game_type || '_' || EXTRACT(YEAR FROM TO_DATE(date, 'YYYYMMDD'))::text
                ELSE NULL
                END
            ORDER BY date, game_number
        ) as series_game_num,
        -- Ranking score for text search
        CASE
            WHEN $1 = '' OR $1 IS NULL THEN 0
            ELSE ts_rank(search_tsv, plainto_tsquery('english', $1))
        END as text_rank
    FROM games
    WHERE
        -- Season filter
        ($2::int IS NULL OR EXTRACT(YEAR FROM TO_DATE(date, 'YYYYMMDD'))::int = $2) AND
        -- Team filter (allows matching either home or away)
        (
            ($3::varchar IS NULL AND $4::varchar IS NULL) OR
            ($3::varchar IS NOT NULL AND (home_team = $3 OR visiting_team = $3)) OR
            ($4::varchar IS NOT NULL AND (home_team = $4 OR visiting_team = $4)) OR
            ($3::varchar IS NOT NULL AND $4::varchar IS NOT NULL AND
                ((home_team = $3 AND visiting_team = $4) OR (home_team = $4 AND visiting_team = $3)))
        ) AND
        -- Game type filter
        ($5::varchar IS NULL OR
            ($5 = 'postseason' AND game_type IN ('worldseries', 'lcs', 'divisionseries', 'wildcard')) OR
            ($5 = 'worldseries' AND game_type = 'worldseries') OR
            game_type = $5
        ) AND
        -- Text search (optional when structured filters present)
        (
            $1 = '' OR $1 IS NULL OR
            -- If we have structured filters, make text search optional
            ($2::int IS NOT NULL OR $3::varchar IS NOT NULL OR $4::varchar IS NOT NULL OR $5::varchar IS NOT NULL OR $6::int IS NOT NULL) OR
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
WHERE
    -- Series game number filter (for playoff games)
    ($6::int IS NULL OR series_game_num = $6)
ORDER BY
    -- Prioritize structured filter matches over text matches
    CASE WHEN $2 IS NOT NULL OR $3 IS NOT NULL OR $4 IS NOT NULL OR $5 IS NOT NULL THEN 1 ELSE 2 END,
    text_rank DESC,
    date DESC
LIMIT $7;
