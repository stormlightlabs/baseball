-- 049_denormalize_league_to_plays.sql
-- Add league columns to plays table for better partition pruning
-- This eliminates the need to join with games table for league filtering

-- Step 1: Add league columns to the partitioned plays table
ALTER TABLE plays ADD COLUMN IF NOT EXISTS home_team_league character varying(10);
ALTER TABLE plays ADD COLUMN IF NOT EXISTS visiting_team_league character varying(10);

-- Step 2: Populate league columns from games table
-- This will take a while for 4.75M rows
UPDATE plays p
SET
    home_team_league = g.home_team_league,
    visiting_team_league = g.visiting_team_league
FROM games g
WHERE p.gid = g.game_id;

-- Step 3: Create indexes on league columns for fast filtering
CREATE INDEX idx_plays_home_team_league ON plays(home_team_league);
CREATE INDEX idx_plays_visiting_team_league ON plays(visiting_team_league);

-- Step 4: Create composite index for league + date (optimal for Negro Leagues queries)
CREATE INDEX idx_plays_home_league_date ON plays(home_team_league, date);
CREATE INDEX idx_plays_visiting_league_date ON plays(visiting_team_league, date);

-- Step 5: Analyze table to update statistics
ANALYZE plays;

-- Note: These indexes will be automatically created on all child partitions
-- The league columns can now be used directly in WHERE clauses without joining games table
