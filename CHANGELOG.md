# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [2025-12-16]

### Added

- Era-based, League-based date partitioning for plays table
- Materialized views for season and career batting/pitching leaderboards
- Retrosheet player data loading functionality
- Command to list all registered API routes

### Changed

- Renamed populate commands to repopulate for clarity

### Fixed

- League column denormalization eliminates expensive joins with games table

## [2025-12-15]

### Added

- Negro Leagues endpoints with comprehensive data coverage (1935-1949)
- Per-game aggregations for batting, pitching, fielding, and team stats
- Achievements tracking system for baseball milestones
- ID crosswalk mapping between Lahman and Retrosheet player/team IDs
- Weather data loading and game context endpoints
- Database indexes for Negro Leagues queries
- Database recreate command for development workflows

### Fixed

- Pagination parameter handling for endpoint consistency & NULL database value handling
  across repositories

## [2025-12-12]

### Added

- Advanced analytics endpoints (wOBA, wRC+, WAR calculations)
- Derived statistics endpoints (leverage index, win expectancy)
- Natural language search for players and teams
- Cache layer implementation with Redis support
- Meta utility endpoints (API version, data coverage, refresh status)
- MLB Stats API proxy for real-time data
- Ejections data from Retrosheet with ETL pipeline
- All-Star game data endpoints
- Stat splits endpoints (home/away, vs LHP/RHP, etc.)
- Pitch sequencing foundation and wOBA constants
- Win expectancy table with batch query optimization

### Changed

- Restructured ROADMAP into domain-specific overview documents
- Embedded complex SQL queries in separate files for maintainability
- Renamed error helper functions for consistency

### Fixed

- Middleware chain ordering for proper request handling
- Retrosheet data field mappings

## [2025-12-11]

### Added

- Games and schedules endpoints (Retrosheet)
- Play-by-play events and context endpoints
- Player endpoints (Lahman with Retrosheet joins)
- Team endpoints (seasons, rosters, splits, franchise data)
- Stats endpoints (batting, pitching, fielding aggregations)
- Team-level statistics and leaderboards
- Awards endpoints (MVP, Cy Young, Hall of Fame, etc.)
- Manager and umpire endpoints
- Search endpoints for player/team lookup
- Authentication system with OAuth support (GitHub, Codeberg)
- Middleware layer (logging, tracing, metrics)
- Configuration system with TOML and environment variables
- ETL metadata tracking (refreshes table for timestamps/row counts)
- Database seeding and population commands
- Docker development environment

### Changed

- Initial repository structure and API architecture
- Swagger documentation generation with swaggo
