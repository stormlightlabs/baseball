# Baseball API Development Roadmap

This roadmap consolidates the guidelines and endpoint specifications into a prioritized development plan.

## Overview

Three-layer architecture:

1. **Raw data → Postgres** (ETL for Lahman + Retrosheet)
2. **Postgres → Go domain model** (queries, joins, views)
3. **Go → HTTP API** (versioned, documented, cached)

## ETL Foundation

### CLI Tool Setup

- **Framework**: cobra + lipgloss for beautiful CLI interface
- **Commands**:
    - `baseball etl fetch lahman` - Download Lahman CSV archive
    - `baseball etl fetch retrosheet` - Download Retrosheet data
    - `baseball etl load lahman` - Load Lahman data into Postgres
    - `baseball etl load retrosheet` - Load Retrosheet data into Postgres
    - `baseball etl status` - Check data freshness and completeness
    - `baseball db migrate` - Run database migrations
    - `baseball server start` - Start server
    - `baseball server fetch [url] --format=json|table` - cURL like test tool
    - `baseball server health` - Health check

### Database Schema

- Use existing Postgres DDL from `Baseball-PostgreSQL` repo
- **Lahman schema**: `people`, `batting`, `pitching`, `fielding`, `teams`, `parks`, etc.
- **Retrosheet schema**: event-level and game-level tables
- Create views for common joins between Lahman & Retrosheet data

### ETL Implementation

- Use `COPY` for fast CSV ingestion
- Generic `CopyCSV` function for reusability
- Progress bars and status reporting with lipgloss
- Error handling and data validation

## Core API (Lahman-focused)

### API Foundation

- **Tech Stack**:
    - Router: standard library
    - DB layer: hand-rolled
    - Config: Viper
    - Migrations: hand-rolled

### Core Endpoints (v1)

#### Players & People

- `GET /v1/players` - Search/browse players
- `GET /v1/players/{player_id}` - Biographical data + career stats
- `GET /v1/players/{player_id}/seasons` - Season-by-season stats
- `GET /v1/players/{player_id}/awards` - Awards history
- `GET /v1/players/{player_id}/hall-of-fame` - HOF voting data

#### Teams & Franchises

- `GET /v1/teams` - List teams by year/league
- `GET /v1/teams/{team_id}` - Single team-season data
- `GET /v1/franchises` - Franchise information
- `GET /v1/seasons/{year}/teams` - All teams in season

#### Stats & Leaderboards

- `GET /v1/seasons/{year}/leaders/batting` - Season batting leaders
- `GET /v1/seasons/{year}/leaders/pitching` - Season pitching leaders
- `GET /v1/stats/batting` - Generic batting stats query
- `GET /v1/stats/pitching` - Generic pitching stats query

## Game Data Integration (Retrosheet)

### Game Endpoints

- `GET /v1/games` - Search games by date/teams/park
- `GET /v1/games/{game_id}` - Game metadata and score
- `GET /v1/games/{game_id}/boxscore` - Detailed boxscore
- `GET /v1/seasons/{year}/schedule` - Full season schedule

### Enhanced Player Data

- `GET /v1/players/{player_id}/game-logs` - Game-by-game performance
- `GET /v1/players/{player_id}/appearances` - Detailed appearance records

## Advanced Features

### Play-by-Play Data

Requires extension of ETL pipeline to include <https://retrosheet.org/downloads/plays.html>

1. <https://www.retrosheet.org/downloads/plays/2023plays.zip>
2. <https://www.retrosheet.org/downloads/plays/2024plays.zip>
3. <https://www.retrosheet.org/downloads/plays/2025plays.zip>

```text
gid              Game ID
event            play as it appears in our event file
inning           inning
top_bot          top (0) or bottom (1) of inning
vis_home         visiting (0) or home (1) team batting
site	         location (ballpark) of event
                 - if a game was suspended and later resumed at a different site, the value for 'site' will change mid-game,
                   indicating the site of specific events
batteam          batting team
pitteam          pitching team
score_v          visiting team score at start of play
score_h          home team score at start of play
batter           batter
pitcher          pitcher
lp               lineup position of batter
bat_f            fielding position of batter
bathand          batter handedness (B, L, R)
pithand          pitcher handedness
balls            number of balls
strikes          number of strikes
count            pitch count, if known; '??' if unknown
                 - balls, strikes, and count do not include final pitch of play
pitches          pitch sequence, if known; blank if unknown
nump		 number of pitches, if known; blank if unknown
pa               did the play result in a plate appearance? (1 if yes, 0 if no)
ab               at bat
single           single
double           double
triple           triple
hr               home run
sh               sacrifice bunt
sf               sacrifice fly, if awarded; in some seasons, sh and sf were combined in official stats; they are separated here
hbp              hit-by-pitch
walk             walk
k                strikeout
xi               batter reached on interference or obstruction, no at bat
roe              batter reached on an error
fc               batter reached on a fielder's choice
othout           batting out not specified otherwise
noout            plate appearance not otherwise specified
                 - plate appearances will be equal to exactly one of the 14 preceding columns ('single' - 'noout')
oth              sum of 'othout' and 'noout' (legacy from earlier releases of 'plays')
bip              ball-in-play; the hit type for bip is identified in the next four columns, if known
bunt             bunt
ground           ground ball
fly              fly ball (or pop up)
line             line drive
iw               intentional walk - subset of 'walk'
gdp              grounded into double play; 'gdp' is an official stat
othdp            double play that was not a 'gdp'
tp               triple play
fle              dropped foul-ball error (batter remains at bat)
wp               wild pitch
pb               passed ball
bk               balk
oa               non-pa not identified by any other stats: either 'out advancing' or 'other advance'
di               defensive indifference
sb2              stolen base of second base
sb3              stolen base of third base
sbh              stolen base of home
cs2              caught stealing second base
cs3              caught stealing third base
csh              caught stealing home
pko1             pickoff at first base
pko2             pickoff at second base
pko3             pickoff at third base
k_safe           strikeout on which the batter reached base safely - subset of 'k'
e1               error(s) by the pitcher; it is possible to be charged multiple errors on a single play
e2               error(s) by the catcher
e3               error(s) by the first baseman
e4               error(s) by the second baseman
e5               error(s) by the third baseman
e6               error(s) by the shortstop
e7               error(s) by the left fielder
e8               error(s) by the center fielder
e9               error(s) by the right fielder
outs_pre         number of outs prior to the play
outs_post        number of outs at the conclusion of the play
br1_pre          runner on first base, if any, prior to the play
br2_pre          runner on second base, if any, prior to the play
br3_pre          runner on third base, if any, prior to the play
br1_post         runner on first base, if any, at the conclusion of the play
br2_post         runner on second base, if any, at the conclusion of the play
br3_post         runner on third base, if any, at the conclusion of the play
lob_id1          id of baserunner left on base after third out of inning
lob_id2          id of baserunner left on base after third out of inning
lob_id3          id of baserunner left on base after third out of inning
pr1_pre          pitcher responsible for runner on first base prior to the play
pr2_pre          pitcher responsible for runner on second base prior to the play
pr3_pre          pitcher responsible for runner on third base prior to the play
pr1_post         pitcher responsible for runner on first base at the conclusion of the play
pr2_post         pitcher responsible for runner on second base at the conclusion of the play
pr3_post         pitcher responsible for runner on third base at the conclusion of the play
run_b            batter if he scored a run on the play
run1             runner on first base if he scored a run on the play
run2             runner on second base if he scored a run on the play
run3             runner on third base if he scored a run on the play
prun_b		 pitcher charged with the run scored by the batter
prun1            pitcher charged with the run scored by the runner on first base
prun2            pitcher charged with the run scored by the runner on second base
prun3            pitcher charged with the run scored by the runner on third base
ur_b		 unearned run scored by batter
ur1		 unearned run scored by Unearned runner on first base
ur2		 unearned run scored by Unearned runner on second base
ur3		 unearned run scored by Unearned runner on third base
rbi_b		 RBI credited for run scored by batter
rbi1		 RBI credited for run scored by runner on first base
rbi2		 RBI credited for run scored by runner on second base
rbi3		 RBI credited for run scored by runner on third base
runs             total runs scored on the play
rbi              total RBI credited to batter on the play
er               total earned runs scored on the play
tur              runs scored on the play credited as (TUR) - team unearned runs
                 - in modern seasons, relief pitchers do not get the benefit of errors which occurred
                   before they entered the game in calculating pitcher-specific earned runs; team
                   earned runs are calculated taking into account all errors in the inning
l1	 	 batting team lineup position 1
l2		 batting team lineup position 2
l3		 batting team lineup position 3
l4		 batting team lineup position 4
l5		 batting team lineup position 5
l6		 batting team lineup position 6
l7		 batting team lineup position 7
l8		 batting team lineup position 8
l9		 batting team lineup position 9
lf1		 fielding position of batting team lineup position 1
lf2		 fielding position of batting team lineup position 2
lf3		 fielding position of batting team lineup position 3
lf4		 fielding position of batting team lineup position 4
lf5		 fielding position of batting team lineup position 5
lf6		 fielding position of batting team lineup position 6
lf7		 fielding position of batting team lineup position 7
lf8		 fielding position of batting team lineup position 8
lf9	 	 fielding position of batting team lineup position 9
f2               catcher
f3               first baseman
f4               second baseman
f5               third baseman
f6               shortstop
f7               left fielder
f8               center fielder
f9               right fielder
po0              putouts on play for which fielder is not identified (scored as 99 in event files)
po1              putouts on play by pitcher
po2              putouts on play by catcher
po3              putouts on play by first baseman
po4              putouts on play by second baseman
po5              putouts on play by third baseman
po6              putouts on play by shortstop
po7              putouts on play by left fielder
po8              putouts on play by center fielder
po9              putouts on play by right fielder
a1               assists on play by pitcher
a2               assists on play by catcher
a3               assists on play by first baseman
a4               assists on play by second baseman
a5               assists on play by third baseman
a6               assists on play by shortstop
a7               assists on play by left fielder
a8               assists on play by center fielder
a9               assists on play by right fielder
fseq		 fielding sequence for out(s): e.g., 643 for 6-4-3 double play (0 indicates unknown fielder)
batout1          fielding position of player who initiated first batting out
batout2          fielding position of player who initiated second batting out (for double play)
batout3          fielding position of player who initiated third batting out (for triple play)
brout_b          fielding position of player who initiated baserunning out by batter
brout1           fielding position of player who initiated baserunning out by runner on first base
brout2           fielding position of player who initiated baserunning out by runner on second base
brout3           fielding position of player who initiated baserunning out by runner on third base
                 - for assisted plays, the player who 'initiated' an out is the player who earned the first assist
firstf           fielding position of first player to field a ball-in-play (1 - 9; 0 if unknown)
loc              location of ball in play, if known
hittype          hit type of ball in play, if known
                 - BG = bunt ground ball, BP = bunt pop-up, BL = bunt line drive, G = ground ball, P = pop up, F = fly ball, L =line drive
dpopp            equal to one if there was a runner on first base and fewer than two outs
pivot            pivot man on double-play opportunity if one existed
                 - on a ground ball where 'dpopp' = 1, 'pivot' is player who made initial putout if a baserunner was forced out,
                   or fielder of ground ball ('firstf') if no force out
                 - idea behind 'pivot' is to establish fielders who deserve to share credit/blame for converting double plays
pn               play number - sequential order of plays within a particular game
umphome		 home plate umpire
ump1b		 1b umpire
ump2b		 2b umpire
ump3b		 3b umpire
umplf		 lf umpire
umprf		 rf umpire
date		 date of game
                 - if a game was suspended and later resumed on a different date, the value for 'date' will change mid-game,
                   to indicate the resumption of the game; 'date' does not change mid-game for games that are played continuously,
   		   but end after midnight local time
gametype         type of game (e.g., regular-season, exhibition, etc.)
pbp              'deduced' or 'full'
```

- `GET /v1/games/{game_id}/plays` - Full play-by-play
- `GET /v1/games/{game_id}/events` - Raw Retrosheet events
- `GET /v1/players/{player_id}/plate-appearances` - All PA with context

### Additional Entities

- Parks, umpires, managers endpoints
- Ejections and special events
- Awards and All-Star game data
- Postseason series and games

### Search & Discovery

- `GET /v1/search/players` - Fuzzy player search
- `GET /v1/search/teams` - Team search
- `GET /v1/search/games` - Natural language game search

## Advanced Analytics

### Derived Statistics

- Player streaks and splits
- Team run differential analysis
- Win probability calculations
- Advanced metrics integration

### API Enhancements

- OpenAPI documentation
- Caching layer
- Rate limiting
- Performance optimization

## Technical Implementation Notes

### Data Sources

- **Lahman**: Season/career stats from SABR (1871-2024)
- **Retrosheet**: Game-level and play-by-play data
- Both sources provide consistent IDs for joins

### ID Conventions

- `player_id`: Lahman playerID (e.g., "troutmi01")
- `team_id`: Lahman teamID (e.g., "LAA")
- `game_id`: Retrosheet GAME_ID (e.g., "ANA201304010")

### Response Format

```json
{
  "data": [...],
  "page": 1,
  "per_page": 50,
  "total": 1234
}
```

### Attribution Requirements

- **Lahman**: Credit SABR + Lahman database with copyright notice
- **Retrosheet**: "The information used here was obtained free of charge from and is copyrighted by Retrosheet"

## Targets

### Lahman ETL + Basic API

- CLI tool with cobra + lipgloss
- Lahman schema + ingestion
- Core player/team/stats endpoints

### Retrosheet Game Logs

- Retrosheet daily logs ingestion
- Game and schedule endpoints

### Play-by-Play

- Full Retrosheet event parsing
- Advanced play-by-play endpoints

### Deployment

- Documentation, caching, optimization
- Joined Lahman + Retrosheet endpoints
- Performance testing and deployment
