-- Extend league columns to accommodate Negro Leagues codes
-- Negro Leagues use 3-character codes (NAL, NNL, NN2, ECL, etc.)
-- while MLB uses 2-character codes (AL, NL)

ALTER TABLE games
ALTER COLUMN home_team_league TYPE varchar(3),
ALTER COLUMN visiting_team_league TYPE varchar(3);
