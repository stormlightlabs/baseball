# API

* **(L)** = primarily from Lahman / Baseball-Databank. ([SABR][3])
* **(R)** = primarily from Retrosheet. ([Retrosheet][4])

Base URL: `https://baseball.stormlightlabs.org/api/v1/...`

## 0. Conventions

**IDs**:

* `player_id` → Lahman `playerID` (e.g. `"troutmi01"`). ([Cdalzell][5])
* `team_id` → Lahman `teamID` for season-level (e.g. `"LAA"`), plus `franchID` for franchise-level endpoints. ([Cdalzell][5])
* `game_id` → Retrosheet `GAME_ID` (e.g. `"ANA201304010"`). ([Kaggle][6])

**Common query params**:

* `page`, `per_page` for pagination; `sort`, `order`.
* `from`, `to` dates use `YYYY-MM-DD`.
* Stats filters use `min_pa`, `min_ip`, `min_g`, etc.

**Response shape (suggested)**:

Wrap lists in an envelope:

```json
{
  "data": [ ... ],
  "page": 1,
  "per_page": 50,
  "total": 1234
}
```

## 1. Meta / Utility

| Verb | Path                | Dataset | Description                                                          |
| ---- | ------------------- | ------- | -------------------------------------------------------------------- |
| GET  | `/v1/health`        | –       | Basic health check.                                                  |
| GET  | `/v1/meta`          | L+R     | API version, last data refresh date, Lahman and Retrosheet versions. |
| GET  | `/v1/meta/datasets` | L+R     | Which seasons, leagues, and tables are loaded and available.         |

These help clients know what they can safely query (e.g. Lahman up to 2024, Retrosheet plays up to a given year). ([SABR][3])

## 2. Players (People & careers) – **(L)** w/ some **(R)** joins

### Core "people" endpoints

| Verb | Path                                   | Description                                                                |
| ---- | -------------------------------------- | -------------------------------------------------------------------------- |
| GET  | `/v1/players`                          | Search/browse players by name, debut year, team, position, handedness.     |
| GET  | `/v1/players/{player_id}`              | Biographical data + summary career stats across batting/pitching/fielding. |
| GET  | `/v1/players/{player_id}/seasons`      | Season-by-season lines (one per team/season).                              |
| GET  | `/v1/players/{player_id}/teams`        | All teams a player has appeared for by year.                               |
| GET  | `/v1/players/{player_id}/awards`       | Awards from Lahman’s awards tables.                                        |
| GET  | `/v1/players/{player_id}/hall-of-fame` | Hall of Fame voting/induction data.                                        |
| GET  | `/v1/players/{player_id}/salaries`     | Salary history from Lahman `Salaries`.                                     |

Lahman’s `People`, `Batting`, `Pitching`, `Fielding`, `Salaries`, `Awards*`, `HallOfFame` tables all support these. ([Bookdown][1])

### Game & play-by-play views for a player – **(R)**

| Verb | Path                                  | Description                                                                                                                      |
| ---- | ------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| GET  | `/v1/players/{player_id}/game-logs`   | Game-by-game stats from Retrosheet game logs.                                                                                    |
| GET  | `/v1/players/{player_id}/appearances` | Detailed appearance records (positions, pinch hits, etc.) by game.                                                               |
| GET  | `/v1/players/{player_id}/events`      | All plate appearances / plays for this player (batting & baserunning), optionally filtered by year, vs_team, leverage flag, etc. |

These use Retrosheet CSVs: game logs, daily logs, and parsed play-by-play data. ([Retrosheet][7])

## 3. Teams, Franchises & Seasons – **(L)** + **(R)**

### Team/franchise reference

| Verb | Path                         | Description                                                              |
| ---- | ---------------------------- | ------------------------------------------------------------------------ |
| GET  | `/v1/teams`                  | List of team entries by year + league (from `Teams`).                    |
| GET  | `/v1/teams/{team_id}`        | Single team-season row (team, league, wins, losses, run diff, etc.).     |
| GET  | `/v1/franchises`             | List of franchises (from `TeamsFranchises`), including historical names. |
| GET  | `/v1/franchises/{franch_id}` | Franchise info + all historical team IDs/locations.                      |
| GET  | `/v1/seasons`                | High-level list of available seasons, with min/max year, leagues.        |

### Team stats & rosters

| Verb | Path                                          | Description                                          |
| ---- | --------------------------------------------- | ---------------------------------------------------- |
| GET  | `/v1/seasons/{year}/teams`                    | All teams in a season with team-level stats.         |
| GET  | `/v1/seasons/{year}/teams/{team_id}/roster`   | Player list with basic stats & positions.            |
| GET  | `/v1/seasons/{year}/teams/{team_id}/batting`  | Aggregated team batting stats plus splits by player. |
| GET  | `/v1/seasons/{year}/teams/{team_id}/pitching` | Same for pitching.                                   |
| GET  | `/v1/seasons/{year}/teams/{team_id}/fielding` | Same for fielding.                                   |

You’re mostly exposing Lahman `Teams` and related tables here. ([Bookdown][1])

### Retrosheet team schedule & logs – **(R)**

| Verb | Path                                            | Description                                                             |
| ---- | ----------------------------------------------- | ----------------------------------------------------------------------- |
| GET  | `/v1/seasons/{year}/teams/{team_id}/schedule`   | Team’s schedule from Retrosheet schedules or game logs.                 |
| GET  | `/v1/seasons/{year}/teams/{team_id}/games`      | All games played by a team in that season with scores, home/away flags. |
| GET  | `/v1/seasons/{year}/teams/{team_id}/daily-logs` | Team daily logs (team-level performance by date).                       |

Backed by Retrosheet game logs / schedule CSVs. ([Retrosheet][7])

## 4. Games & Schedules – **(R)**

### Core game endpoints

| Verb | Path                           | Description                                                            |
| ---- | ------------------------------ | ---------------------------------------------------------------------- |
| GET  | `/v1/games`                    | Search/browse games by date range, season, teams, park, etc.           |
| GET  | `/v1/games/{game_id}`          | Single game: metadata, score, linescore, starting lineups.             |
| GET  | `/v1/games/{game_id}/boxscore` | Expanded boxscore (per-player lines) assembled from Retrosheet fields. |
| GET  | `/v1/games/{game_id}/plays`    | Ordered list of plays / plate appearances (play-by-play).              |
| GET  | `/v1/games/{game_id}/summary`  | Optional: summary object – winning pitcher, save, key events, etc.     |

These endpoints sit directly on top of Retrosheet parsed play-by-play / game logs. ([Retrosheet][8])

### Season-level schedules

| Verb | Path                                       | Description                             |
| ---- | ------------------------------------------ | --------------------------------------- |
| GET  | `/v1/seasons/{year}/schedule`              | Full schedule for a season (all games). |
| GET  | `/v1/seasons/{year}/dates/{date}/games`    | All games played on a given date.       |
| GET  | `/v1/seasons/{year}/parks/{park_id}/games` | All games played in a given park.       |

Retrosheet CSV contents include schedules, which make this straightforward. ([Retrosheet][2])

## 5. Play-by-Play Events & Context – **(R)**

For advanced consumers who want full detail.

| Verb | Path                                        | Description                                                                   |
| ---- | ------------------------------------------- | ----------------------------------------------------------------------------- |
| GET  | `/v1/games/{game_id}/events`                | Raw events/plays in Retrosheet format plus parsed fields.                     |
| GET  | `/v1/games/{game_id}/events/{event_seq}`    | Single event by sequence number.                                              |
| GET  | `/v1/players/{player_id}/plate-appearances` | All plate appearances with filters (count, base/out state, vs_pitcher, etc.). |
| GET  | `/v1/pitches`                               | *Optional later*: if you derive pitch-level data from other sources.          |

Underlying data: parsed Retrosheet "plays" CSV (fully parsed PBP with 160+ columns). ([Retrosheet][9])

## 6. Parks, Umpires, Managers & Other Entities – **(L+R)**

Lahman and Retrosheet both expose **parks**, **umpires**, **managers**, **coaches**, etc. ([Retrosheet][4])

### Parks / Ballparks

| Verb | Path                  | Description                                                                |
| ---- | --------------------- | -------------------------------------------------------------------------- |
| GET  | `/v1/parks`           | List ballparks with location, open/close years (from Lahman + Retrosheet). |
| GET  | `/v1/parks/{park_id}` | Single park + aggregated info (games hosted, teams, years active).         |

### Managers & umpires

| Verb | Path                        | Description                                          |
| ---- | --------------------------- | ---------------------------------------------------- |
| GET  | `/v1/managers`              | Manager list with career wins/losses, teams managed. |
| GET  | `/v1/managers/{manager_id}` | Per-manager record + season breakdown.               |
| GET  | `/v1/umpires`               | Umpire list.                                         |
| GET  | `/v1/umpires/{umpire_id}`   | Umpire details + games officiated.                   |

These use Lahman `Managers` and Retrosheet’s biographical and game metadata. ([Bookdown][1])

### Ejections & special events – **(R)**

| Verb | Path                           | Description                                               |
| ---- | ------------------------------ | --------------------------------------------------------- |
| GET  | `/v1/ejections`                | All ejection events with filters by player, umpire, year. |
| GET  | `/v1/seasons/{year}/ejections` | All ejections in a given season.                          |

Retrosheet provides an ejections CSV, which is ideal for this. ([Retrosheet][2])

## 7. Stats & Leaderboards – **(L)** (with optional **(R)** joins)

These endpoints live on top of views/materialized views aggregating Lahman stats. ([Bookdown][1])

### Career & season stats

| Verb | Path                                     | Description                                                        |
| ---- | ---------------------------------------- | ------------------------------------------------------------------ |
| GET  | `/v1/stats/batting`                      | Generic batting stats query (filters: year, team, league, min_pa). |
| GET  | `/v1/stats/pitching`                     | Same for pitching (min_ip, role filter).                           |
| GET  | `/v1/stats/fielding`                     | Same for fielding.                                                 |
| GET  | `/v1/players/{player_id}/stats/batting`  | Player-specific (career + optional split by year).                 |
| GET  | `/v1/players/{player_id}/stats/pitching` | Same for pitching.                                                 |

### Seasonal leaders

| Verb | Path                                  | Description                                               |
| ---- | ------------------------------------- | --------------------------------------------------------- |
| GET  | `/v1/seasons/{year}/leaders/batting`  | Leaders for a given stat (`stat=HR`, `AVG`, `OBP`, etc.). |
| GET  | `/v1/seasons/{year}/leaders/pitching` | Leaders for ERA, SO, IP, saves, etc.                      |
| GET  | `/v1/leaders/batting/career`          | Career leaders (HR, RBI, WAR if you compute it).          |
| GET  | `/v1/leaders/pitching/career`         | Career pitching leaders (SO, wins, innings, etc.).        |

### Team-level stats

| Verb | Path                       | Description                                           |
| ---- | -------------------------- | ----------------------------------------------------- |
| GET  | `/v1/stats/teams/batting`  | Team batting in a given season, filterable by league. |
| GET  | `/v1/stats/teams/pitching` | Team pitching.                                        |
| GET  | `/v1/stats/teams/fielding` | Team fielding.                                        |

You might back these with precomputed views or materialized views computed during ETL to keep runtime queries simple.

## 8. Awards, All-Star Games, Postseason – **(L)**

Lahman has tables for awards, All-Star games, postseason series and games. ([Bookdown][1])

| Verb | Path                                   | Description                                                    |
| ---- | -------------------------------------- | -------------------------------------------------------------- |
| GET  | `/v1/awards`                           | Browse awards data (MVP, Cy Young, ROY, etc.).                 |
| GET  | `/v1/awards/{award_id}`                | Detail view for an award (participants and years).             |
| GET  | `/v1/seasons/{year}/awards`            | All awards for a given season.                                 |
| GET  | `/v1/seasons/{year}/postseason/series` | Postseason series list (LCS, WS, etc.).                        |
| GET  | `/v1/seasons/{year}/postseason/games`  | All postseason games with series and game number.              |
| GET  | `/v1/allstar/games`                    | All-Star Game list across seasons.                             |
| GET  | `/v1/allstar/games/{game_id}`          | All-Star game specific box + events (if joined to Retrosheet). |

## 9. Search & Lookup Utilities – **(L+R)**

Make it easy to "jump" to IDs from approximate input.

| Verb | Path                 | Description                                                           |
| ---- | -------------------- | --------------------------------------------------------------------- |
| GET  | `/v1/search/players` | Fuzzy search by name, optional filters (era, league, team).           |
| GET  | `/v1/search/teams`   | Search by name, city, franchise.                                      |
| GET  | `/v1/search/games`   | Search by natural language query ("Yankees vs Red Sox 2003 ALCS G7"). |
| GET  | `/v1/search/parks`   | Search ballparks by name/city.                                        |

Under the hood, you can build these on top of simple trigram/fuzzy matching in Postgres, plus straightforward filters.

## 10. Derived Endpoints

These aren’t directly "available" as raw tables, but they’re natural uses of the data and might be worth planning for:

* `/v1/players/{player_id}/streaks` – hitting streaks, scoreless inning streaks.
* `/v1/players/{player_id}/splits` – home/away, LHP/RHP, month, batting order.
* `/v1/teams/{team_id}/run-differential` – per season and rolling 30-game windows.
* `/v1/games/{game_id}/win-probability` – if you compute win-probability graphs from play-by-play.

[1]: https://bookdown.org/vankrevelen/BaseballAnalyticsGuide/lahman.html "Chapter 2 Lahman | Fundamentals of Collecting and ..."
[2]: https://www.retrosheet.org/downloads/csvcontents.html "Contents of CSV Download Files"
[3]: https://sabr.org/lahman-database/ "Lahman Baseball Database"
[4]: https://www.retrosheet.org/game.htm "Retrosheet Event Files"
[5]: https://cdalzell.github.io/Lahman/ "Sean Lahman Baseball Database"
[6]: https://www.kaggle.com/datasets/freshrenzo/lahmanbaseballdatabase "Lahman Baseball Database"
[7]: https://www.retrosheet.org/downloads/csvdownloads.html "Daily Logs (CSV Files) Available for Download"
[8]: https://www.retrosheet.org/eventfile.htm "Retrosheet Event Files"
[9]: https://retrosheet.org/downloads/plays.html "Parsed Play-by-Play Data"
