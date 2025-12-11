-- Retrosheet Play-by-Play (Plays) Table
-- Source: https://www.retrosheet.org/downloads/plays.html
-- This table contains detailed play-by-play data for each game

DROP TABLE IF EXISTS plays;
CREATE TABLE plays (
    -- Game identification
    gid varchar(12) NOT NULL,           -- Game ID
    event text,                         -- Play as it appears in event file
    inning int,                         -- Inning number
    top_bot int,                        -- Top (0) or bottom (1) of inning
    vis_home int,                       -- Visiting (0) or home (1) team batting
    site varchar(5),                    -- Location (ballpark) of event
    batteam varchar(3),                 -- Batting team
    pitteam varchar(3),                 -- Pitching team
    score_v int,                        -- Visiting team score at start of play
    score_h int,                        -- Home team score at start of play

    -- Players involved
    batter varchar(8),                  -- Batter ID
    pitcher varchar(8),                 -- Pitcher ID
    lp int,                             -- Lineup position of batter
    bat_f int,                          -- Fielding position of batter
    bathand varchar(1),                 -- Batter handedness (B, L, R)
    pithand varchar(1),                 -- Pitcher handedness

    -- Pitch count
    balls int,                          -- Number of balls
    strikes int,                        -- Number of strikes
    count varchar(10),                  -- Pitch count (e.g., "31", "??" if unknown)
    pitches text,                       -- Pitch sequence
    nump int,                           -- Number of pitches

    -- Plate appearance outcome
    pa int,                             -- Plate appearance (1 if yes, 0 if no)
    ab int,                             -- At bat
    single int,                         -- Single
    double int,                         -- Double
    triple int,                         -- Triple
    hr int,                             -- Home run
    sh int,                             -- Sacrifice bunt
    sf int,                             -- Sacrifice fly
    hbp int,                            -- Hit-by-pitch
    walk int,                           -- Walk
    k int,                              -- Strikeout
    xi int,                             -- Reached on interference/obstruction
    roe int,                            -- Reached on error
    fc int,                             -- Fielder's choice
    othout int,                         -- Other batting out
    noout int,                          -- No out (PA not otherwise specified)
    oth int,                            -- Sum of othout and noout (legacy)

    -- Ball in play details
    bip int,                            -- Ball-in-play
    bunt int,                           -- Bunt
    ground int,                         -- Ground ball
    fly int,                            -- Fly ball or pop up
    line int,                           -- Line drive

    -- Special plays
    iw int,                             -- Intentional walk
    gdp int,                            -- Grounded into double play
    othdp int,                          -- Other double play
    tp int,                             -- Triple play
    fle int,                            -- Dropped foul-ball error
    wp int,                             -- Wild pitch
    pb int,                             -- Passed ball
    bk int,                             -- Balk
    oa int,                             -- Out advancing or other advance
    di int,                             -- Defensive indifference

    -- Stolen bases and caught stealing
    sb2 int,                            -- Stolen base of second
    sb3 int,                            -- Stolen base of third
    sbh int,                            -- Stolen base of home
    cs2 int,                            -- Caught stealing second
    cs3 int,                            -- Caught stealing third
    csh int,                            -- Caught stealing home

    -- Pickoffs
    pko1 int,                           -- Pickoff at first
    pko2 int,                           -- Pickoff at second
    pko3 int,                           -- Pickoff at third
    k_safe int,                         -- Strikeout but reached base safely

    -- Errors by position
    e1 int,                             -- Error by pitcher
    e2 int,                             -- Error by catcher
    e3 int,                             -- Error by first baseman
    e4 int,                             -- Error by second baseman
    e5 int,                             -- Error by third baseman
    e6 int,                             -- Error by shortstop
    e7 int,                             -- Error by left fielder
    e8 int,                             -- Error by center fielder
    e9 int,                             -- Error by right fielder

    -- Outs and baserunners before/after play
    outs_pre int,                       -- Outs before the play
    outs_post int,                      -- Outs after the play
    br1_pre varchar(8),                 -- Runner on first before play
    br2_pre varchar(8),                 -- Runner on second before play
    br3_pre varchar(8),                 -- Runner on third before play
    br1_post varchar(8),                -- Runner on first after play
    br2_post varchar(8),                -- Runner on second after play
    br3_post varchar(8),                -- Runner on third after play

    -- Runners left on base
    lob_id1 varchar(8),                 -- Runner 1 left on base after third out
    lob_id2 varchar(8),                 -- Runner 2 left on base after third out
    lob_id3 varchar(8),                 -- Runner 3 left on base after third out

    -- Pitcher responsible for runners before/after
    pr1_pre varchar(8),                 -- Pitcher responsible for runner on first (pre)
    pr2_pre varchar(8),                 -- Pitcher responsible for runner on second (pre)
    pr3_pre varchar(8),                 -- Pitcher responsible for runner on third (pre)
    pr1_post varchar(8),                -- Pitcher responsible for runner on first (post)
    pr2_post varchar(8),                -- Pitcher responsible for runner on second (post)
    pr3_post varchar(8),                -- Pitcher responsible for runner on third (post)

    -- Runs scored
    run_b varchar(8),                   -- Batter if he scored
    run1 varchar(8),                    -- Runner on first if he scored
    run2 varchar(8),                    -- Runner on second if he scored
    run3 varchar(8),                    -- Runner on third if he scored
    prun_b varchar(8),                  -- Pitcher charged with run (batter)
    prun1 varchar(8),                   -- Pitcher charged with run (runner 1)
    prun2 varchar(8),                   -- Pitcher charged with run (runner 2)
    prun3 varchar(8),                   -- Pitcher charged with run (runner 3)

    -- Unearned runs
    ur_b int,                           -- Unearned run scored by batter
    ur1 int,                            -- Unearned run scored by runner 1
    ur2 int,                            -- Unearned run scored by runner 2
    ur3 int,                            -- Unearned run scored by runner 3

    -- RBIs
    rbi_b int,                          -- RBI for batter's run
    rbi1 int,                           -- RBI for runner 1's run
    rbi2 int,                           -- RBI for runner 2's run
    rbi3 int,                           -- RBI for runner 3's run
    runs int,                           -- Total runs scored on the play
    rbi int,                            -- Total RBI credited to batter
    er int,                             -- Total earned runs scored
    tur int,                            -- Team unearned runs

    -- Lineups (batting team)
    l1 varchar(8),                      -- Lineup position 1
    l2 varchar(8),                      -- Lineup position 2
    l3 varchar(8),                      -- Lineup position 3
    l4 varchar(8),                      -- Lineup position 4
    l5 varchar(8),                      -- Lineup position 5
    l6 varchar(8),                      -- Lineup position 6
    l7 varchar(8),                      -- Lineup position 7
    l8 varchar(8),                      -- Lineup position 8
    l9 varchar(8),                      -- Lineup position 9

    -- Fielding positions of lineup (batting team)
    lf1 int,                            -- Fielding position of lineup 1
    lf2 int,                            -- Fielding position of lineup 2
    lf3 int,                            -- Fielding position of lineup 3
    lf4 int,                            -- Fielding position of lineup 4
    lf5 int,                            -- Fielding position of lineup 5
    lf6 int,                            -- Fielding position of lineup 6
    lf7 int,                            -- Fielding position of lineup 7
    lf8 int,                            -- Fielding position of lineup 8
    lf9 int,                            -- Fielding position of lineup 9

    -- Fielding team positions
    f2 varchar(8),                      -- Catcher
    f3 varchar(8),                      -- First baseman
    f4 varchar(8),                      -- Second baseman
    f5 varchar(8),                      -- Third baseman
    f6 varchar(8),                      -- Shortstop
    f7 varchar(8),                      -- Left fielder
    f8 varchar(8),                      -- Center fielder
    f9 varchar(8),                      -- Right fielder

    -- Putouts by position
    po0 int,                            -- Putouts (fielder unknown)
    po1 int,                            -- Putouts by pitcher
    po2 int,                            -- Putouts by catcher
    po3 int,                            -- Putouts by first baseman
    po4 int,                            -- Putouts by second baseman
    po5 int,                            -- Putouts by third baseman
    po6 int,                            -- Putouts by shortstop
    po7 int,                            -- Putouts by left fielder
    po8 int,                            -- Putouts by center fielder
    po9 int,                            -- Putouts by right fielder

    -- Assists by position
    a1 int,                             -- Assists by pitcher
    a2 int,                             -- Assists by catcher
    a3 int,                             -- Assists by first baseman
    a4 int,                             -- Assists by second baseman
    a5 int,                             -- Assists by third baseman
    a6 int,                             -- Assists by shortstop
    a7 int,                             -- Assists by left fielder
    a8 int,                             -- Assists by center fielder
    a9 int,                             -- Assists by right fielder

    -- Fielding sequence and outs
    fseq varchar(20),                   -- Fielding sequence (e.g., "643" for 6-4-3 DP)
    batout1 int,                        -- Position initiating first batting out
    batout2 int,                        -- Position initiating second batting out (DP)
    batout3 int,                        -- Position initiating third batting out (TP)
    brout_b int,                        -- Position initiating baserunning out (batter)
    brout1 int,                         -- Position initiating baserunning out (runner 1)
    brout2 int,                         -- Position initiating baserunning out (runner 2)
    brout3 int,                         -- Position initiating baserunning out (runner 3)

    -- Ball in play details
    firstf int,                         -- First fielder to field ball (1-9, 0 if unknown)
    loc varchar(10),                    -- Location of ball in play
    hittype varchar(5),                 -- Hit type (BG, BP, BL, G, P, F, L)
    dpopp int,                          -- Double play opportunity (runner on first, <2 outs)
    pivot int,                          -- Pivot man on DP opportunity
    pn int,                             -- Play number (sequential within game)

    -- Umpires
    umphome varchar(8),                 -- Home plate umpire
    ump1b varchar(8),                   -- First base umpire
    ump2b varchar(8),                   -- Second base umpire
    ump3b varchar(8),                   -- Third base umpire
    umplf varchar(8),                   -- Left field umpire
    umprf varchar(8),                   -- Right field umpire

    -- Game metadata
    date varchar(8),                    -- Date of game (YYYYMMDD)
    gametype varchar(20),               -- Type of game (regular, etc.)
    pbp varchar(10)                     -- Play-by-play type (deduced or full)
);

CREATE INDEX idx_plays_gid ON plays(gid);
CREATE INDEX idx_plays_batter ON plays(batter);
CREATE INDEX idx_plays_pitcher ON plays(pitcher);
CREATE INDEX idx_plays_date ON plays(date);
CREATE INDEX idx_plays_batteam ON plays(batteam);
CREATE INDEX idx_plays_pitteam ON plays(pitteam);
