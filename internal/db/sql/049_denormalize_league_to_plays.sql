-- Add league columns to plays table for better partition pruning

ALTER TABLE plays ADD COLUMN IF NOT EXISTS home_team_league character varying(10);
ALTER TABLE plays ADD COLUMN IF NOT EXISTS visiting_team_league character varying(10);

UPDATE plays p
SET
    home_team_league = g.home_team_league,
    visiting_team_league = g.visiting_team_league
FROM games g
WHERE p.gid = g.game_id;

CREATE INDEX idx_plays_home_team_league ON plays(home_team_league);
CREATE INDEX idx_plays_visiting_team_league ON plays(visiting_team_league);

CREATE INDEX idx_plays_home_league_date ON plays(home_team_league, date);
CREATE INDEX idx_plays_visiting_league_date ON plays(visiting_team_league, date);

ANALYZE plays;
