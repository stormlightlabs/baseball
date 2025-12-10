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
