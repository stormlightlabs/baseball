-- Add primary key to games table
-- A game is uniquely identified by date, home team, and game number (for doubleheaders)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'games'::regclass
            AND contype = 'p'
    ) THEN
        ALTER TABLE games ADD PRIMARY KEY (date, home_team, game_number);
    END IF;
END $$;

-- Add primary key to plays table
-- A play is uniquely identified by game ID and play number
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'plays'::regclass
            AND contype = 'p'
    ) THEN
        ALTER TABLE plays ADD PRIMARY KEY (gid, pn);
    END IF;
END $$;
