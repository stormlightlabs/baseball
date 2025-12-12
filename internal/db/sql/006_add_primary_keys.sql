-- Add primary key to games table
-- A game is uniquely identified by date, home team, and game number (for doubleheaders)
ALTER TABLE games ADD PRIMARY KEY (date, home_team, game_number);

-- Add primary key to plays table
-- A play is uniquely identified by game ID and play number
ALTER TABLE plays ADD PRIMARY KEY (gid, pn);
