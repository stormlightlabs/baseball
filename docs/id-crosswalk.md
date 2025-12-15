# Dataset ID Crosswalk System

Maps player, team, and park identifiers across Lahman and Retrosheet datasets. Enables seamless joins and dual-ID API lookups.

## Coverage

| Crosswalk           | Source      | Total | Mapped | Coverage | Notes                                     |
| ------------------- | ----------- | ----- | ------ | -------- | ----------------------------------------- |
| Player IDs          | People      | 24K   | 21K    | 88%      | Deduplicates records with same Retro ID   |
| Team/Franchise      | Teams       | 256   | 256    | 100%     | Tracks relocations and name changes       |
| Parks               | Games       | 449   | 142    | 32%      | Missing parks are primarily Negro Leagues |

## Player ID Mapping

Maps Lahman `playerID` to Retrosheet `retroID`. Enables API lookups with either format.

### Views & Functions

| View/Function       | Purpose                                    |
| ------------------- | ------------------------------------------ |
| `player_id_map`     | Materialized view with bidirectional map   |
| `lahman_to_retro()` | Convert Lahman ID to Retrosheet ID         |
| `retro_to_lahman()` | Convert Retrosheet ID to Lahman ID         |

### API Usage

API endpoints automatically accept both ID formats. Responses include both IDs when available.

```sh
GET /v1/players/troutmi01  # Lahman ID
GET /v1/players/troum001   # Retrosheet ID
# Both return same player with both IDs in response
```

## Team/Franchise Mapping

Links team codes to franchises across seasons. Handles relocations and historical name changes.

### Views & Functions

| View/Function              | Purpose                                           |
| -------------------------- | ------------------------------------------------- |
| `team_franchise_map`       | Team-season-franchise relationships               |
| `franchise_current_team()` | Get current team ID for franchise                 |
| `franchise_all_teams()`    | Get all historical teams for franchise            |

## Park Mapping

Cross-references Retrosheet park codes with Lahman park metadata. Enriches game responses with park names and locations.

### Views & Functions

| View/Function              | Purpose                                      |
| -------------------------- | -------------------------------------------- |
| `park_map`                 | All parks with metadata where available      |
| `parks_missing_from_lahman`| High-usage parks needing metadata            |
| `get_park_info()`          | Lookup park details by Retrosheet code       |
| `active_parks()`           | Parks used since specified year              |

### API Enrichment

Game endpoints automatically include park metadata:

```sh
GET /v1/games/NYA202509280
# Returns park_name, park_city, park_state alongside park_id
```

## Implementation Details

All crosswalks use materialized views with unique indexes for concurrent refresh.
Helper functions simplify common lookups without explicit joins.

Player repository accepts either ID format automatically via crosswalk join.
Game repository enriches responses with park metadata from park_map.

Duplicate handling: Player crosswalk deduplicates by preferring records with debut dates.
Park crosswalk deduplicates Parks table using DISTINCT ON parkkey.
