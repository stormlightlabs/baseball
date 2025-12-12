package core

import (
	"context"
	"time"
)

// PlayerRepository encapsulates all "player-centric" access: bio, seasons, per-player stats, game logs.
type PlayerRepository interface {
	GetByID(ctx context.Context, id PlayerID) (*Player, error)
	List(ctx context.Context, filter PlayerFilter) ([]Player, error)
	Count(ctx context.Context, filter PlayerFilter) (int, error)

	BattingSeasons(ctx context.Context, id PlayerID) ([]PlayerBattingSeason, error)
	PitchingSeasons(ctx context.Context, id PlayerID) ([]PlayerPitchingSeason, error)
	FieldingSeasons(ctx context.Context, id PlayerID) ([]PlayerFieldingSeason, error)

	// Game-level view for a player (from Retrosheet logs / day-by-day).
	GameLogs(ctx context.Context, id PlayerID, filter GameFilter) ([]Game, error)

	// Appearance records by position for a player
	Appearances(ctx context.Context, id PlayerID) ([]PlayerAppearance, error)

	// Team history (season-by-season)
	Teams(ctx context.Context, id PlayerID) ([]PlayerTeamSeason, error)

	// Salary history (Lahman Salaries table)
	Salaries(ctx context.Context, id PlayerID) ([]PlayerSalary, error)
}

// TeamRepository handles team & franchise views.
type TeamRepository interface {
	GetTeamSeason(ctx context.Context, teamID TeamID, year SeasonYear) (*TeamSeason, error)
	ListTeamSeasons(ctx context.Context, filter TeamFilter) ([]TeamSeason, error)
	CountTeamSeasons(ctx context.Context, filter TeamFilter) (int, error)

	GetFranchise(ctx context.Context, id FranchiseID) (*Franchise, error)
	ListFranchises(ctx context.Context, onlyActive bool) ([]Franchise, error)

	// ListSeasons returns all available seasons with league and team counts
	ListSeasons(ctx context.Context) ([]Season, error)

	// Roster for a given team & season; built from batting/fielding/pitching joins.
	Roster(ctx context.Context, year SeasonYear, teamID TeamID) ([]RosterPlayer, error)

	// Team aggregate stats with optional per-player splits
	BattingStats(ctx context.Context, year SeasonYear, teamID TeamID, includePlayers bool) (*TeamBattingStats, error)
	PitchingStats(ctx context.Context, year SeasonYear, teamID TeamID, includePlayers bool) (*TeamPitchingStats, error)
	FieldingStats(ctx context.Context, year SeasonYear, teamID TeamID, includePlayers bool) (*TeamFieldingStats, error)
}

// GameRepository manages game-log level data.
type GameRepository interface {
	GetByID(ctx context.Context, id GameID) (*Game, error)
	List(ctx context.Context, filter GameFilter) ([]Game, error)
	Count(ctx context.Context, filter GameFilter) (int, error)

	// Convenience helpers for common views:
	ListByDate(ctx context.Context, date time.Time) ([]Game, error)
	ListByTeamSeason(ctx context.Context, teamID TeamID, year SeasonYear, p Pagination) ([]Game, error)

	// Get detailed boxscore for a game
	GetBoxscore(ctx context.Context, id GameID) (*Boxscore, error)
}

// PlayRepository manages play-by-play data from Retrosheet.
type PlayRepository interface {
	// List retrieves plays based on filter criteria
	List(ctx context.Context, filter PlayFilter) ([]Play, error)
	Count(ctx context.Context, filter PlayFilter) (int, error)

	// ListByGame retrieves all plays for a specific game in order
	ListByGame(ctx context.Context, gameID GameID, p Pagination) ([]Play, error)

	// ListByPlayer retrieves plays involving a specific player (as batter or pitcher)
	ListByPlayer(ctx context.Context, playerID RetroPlayerID, p Pagination) ([]Play, error)

	// CountByPlayer returns the number of plays where the player was batter or pitcher.
	CountByPlayer(ctx context.Context, playerID RetroPlayerID) (int, error)
}

// EventRepository manages play-by-play events.
// Deprecated: Use PlayRepository for new code.
type EventRepository interface {
	ListByGame(ctx context.Context, gameID GameID, p Pagination) ([]GameEvent, error)
	List(ctx context.Context, filter EventFilter) ([]GameEvent, error)
	Count(ctx context.Context, filter EventFilter) (int, error)
}

// ParkRepository for ballpark metadata and usage.
type ParkRepository interface {
	GetByID(ctx context.Context, id ParkID) (*Park, error)
	List(ctx context.Context, filter ParkFilter) ([]Park, error)

	GamesAtPark(ctx context.Context, id ParkID, filter GameFilter) ([]Game, error)
}

// ManagerRepository / UmpireRepository for people in those roles.
type ManagerRepository interface {
	GetByID(ctx context.Context, id ManagerID) (*Manager, error)
	List(ctx context.Context, p Pagination) ([]Manager, error)
	SeasonRecords(ctx context.Context, id ManagerID) ([]ManagerSeasonRecord, error)
}

type UmpireRepository interface {
	GetByID(ctx context.Context, id UmpireID) (*Umpire, error)
	List(ctx context.Context, p Pagination) ([]Umpire, error)
	GamesForUmpire(ctx context.Context, id UmpireID, filter GameFilter) ([]Game, error)
}

// AwardRepository for awards and Hall of Fame.
type AwardRepository interface {
	ListAwards(ctx context.Context) ([]Award, error)

	ListAwardResults(ctx context.Context, filter AwardFilter) ([]AwardResult, error)
	CountAwardResults(ctx context.Context, filter AwardFilter) (int, error)

	HallOfFameByPlayer(ctx context.Context, id PlayerID) ([]HallOfFameRecord, error)

	// All-star game methods
	ListAllStarGames(ctx context.Context, year *SeasonYear) ([]AllStarGame, error)
	GetAllStarGame(ctx context.Context, gameID string) (*AllStarGame, error)
}

// PostseasonRepository for postseason series data.
type PostseasonRepository interface {
	ListSeries(ctx context.Context, year SeasonYear) ([]PostseasonSeries, error)
}

// StatsRepository for season/career leaderboards and arbitrary stat queries.
// Backed by views or materialized views (batting, pitching, fielding).
type StatsRepository interface {
	SeasonBattingLeaders(ctx context.Context, year SeasonYear, stat string, limit, offset int, league *LeagueID) ([]PlayerBattingSeason, error)
	CareerBattingLeaders(ctx context.Context, stat string, limit, offset int) ([]PlayerBattingSeason, error)

	SeasonPitchingLeaders(ctx context.Context, year SeasonYear, stat string, limit, offset int, league *LeagueID) ([]PlayerPitchingSeason, error)
	CareerPitchingLeaders(ctx context.Context, stat string, limit, offset int) ([]PlayerPitchingSeason, error)

	TeamSeasonStats(ctx context.Context, filter TeamFilter) ([]TeamSeason, error)

	QueryBattingStats(ctx context.Context, filter BattingStatsFilter) ([]PlayerBattingSeason, error)
	QueryBattingStatsCount(ctx context.Context, filter BattingStatsFilter) (int, error)

	QueryPitchingStats(ctx context.Context, filter PitchingStatsFilter) ([]PlayerPitchingSeason, error)
	QueryPitchingStatsCount(ctx context.Context, filter PitchingStatsFilter) (int, error)

	QueryFieldingStats(ctx context.Context, filter FieldingStatsFilter) ([]PlayerFieldingSeason, error)
	QueryFieldingStatsCount(ctx context.Context, filter FieldingStatsFilter) (int, error)

	TeamBattingStats(ctx context.Context, filter TeamStatsFilter) ([]TeamBattingStats, error)
	TeamBattingStatsCount(ctx context.Context, filter TeamStatsFilter) (int, error)

	TeamPitchingStats(ctx context.Context, filter TeamStatsFilter) ([]TeamPitchingStats, error)
	TeamPitchingStatsCount(ctx context.Context, filter TeamStatsFilter) (int, error)

	TeamFieldingStats(ctx context.Context, filter TeamStatsFilter) ([]TeamFieldingStats, error)
	TeamFieldingStatsCount(ctx context.Context, filter TeamStatsFilter) (int, error)
}

// MetaRepository for API/dataset metadata.
type MetaRepository interface {
	// Returns min/max seasons available from Lahman and Retrosheet.
	SeasonCoverage(ctx context.Context) (minLahman, maxLahman, minRetrosheet, maxRetrosheet SeasonYear, err error)

	// When each dataset was last refreshed.
	LastUpdated(ctx context.Context) (lahman time.Time, retrosheet time.Time, err error)

	// DatasetStatuses surfaces ETL and coverage metadata per dataset.
	DatasetStatuses(ctx context.Context) ([]DatasetStatus, error)

	// SchemaHashes returns hashes grouped by migration family/dataset.
	SchemaHashes(ctx context.Context) (map[string]string, error)
}

// SearchRepository lets you do fuzzy lookups across entities.
type SearchRepository interface {
	SearchPlayers(ctx context.Context, filter SearchFilter) ([]Player, error)
	SearchTeams(ctx context.Context, filter SearchFilter) ([]TeamSeason, error)
	SearchGames(ctx context.Context, filter GameFilter) ([]Game, error)
	SearchParks(ctx context.Context, filter SearchFilter) ([]Park, error)
}

// UserRepository handles user authentication and management.
type UserRepository interface {
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id UserID) (*User, error)

	// GetByEmail retrieves a user by email address
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Create creates a new user account
	Create(ctx context.Context, email string, name *string, avatarURL *string) (*User, error)

	// Update updates user information
	Update(ctx context.Context, user *User) error

	// UpdateLastLogin updates the last login timestamp
	UpdateLastLogin(ctx context.Context, id UserID) error

	// List retrieves all users (for admin purposes)
	List(ctx context.Context, p Pagination) ([]User, error)
}

// APIKeyRepository handles API key generation and validation.
type APIKeyRepository interface {
	// Create generates a new API key for a user
	Create(ctx context.Context, userID UserID, name *string, expiresAt *time.Time) (*APIKey, string, error)

	// GetByKey retrieves an API key by its value (for validation)
	GetByKey(ctx context.Context, key string) (*APIKey, error)

	// GetByID retrieves an API key by ID
	GetByID(ctx context.Context, id string) (*APIKey, error)

	// ListByUser retrieves all API keys for a user
	ListByUser(ctx context.Context, userID UserID) ([]APIKey, error)

	// Revoke deactivates an API key
	Revoke(ctx context.Context, id string) error

	// UpdateLastUsed updates the last used timestamp
	UpdateLastUsed(ctx context.Context, id string) error
}

// OAuthTokenRepository handles OAuth2 token storage and validation.
type OAuthTokenRepository interface {
	// Create stores a new OAuth token
	Create(ctx context.Context, userID UserID, accessToken string, refreshToken *string, expiresAt time.Time) (*OAuthToken, error)

	// GetByAccessToken retrieves a token by access token value
	GetByAccessToken(ctx context.Context, accessToken string) (*OAuthToken, error)

	// GetByUserID retrieves the active token for a user
	GetByUserID(ctx context.Context, userID UserID) (*OAuthToken, error)

	// Delete removes a token (for logout)
	Delete(ctx context.Context, id string) error

	// DeleteExpired removes all expired tokens
	DeleteExpired(ctx context.Context) (int64, error)
}

// UsageRepository tracks API usage for rate limiting and analytics.
type UsageRepository interface {
	// Record logs an API request
	Record(ctx context.Context, userID *UserID, apiKeyID *string, endpoint string, method string, statusCode int, responseTimeMs *int) error

	// GetUserUsage retrieves usage stats for a user
	GetUserUsage(ctx context.Context, userID UserID, since time.Time) ([]APIUsage, error)

	// GetAPIKeyUsage retrieves usage stats for an API key
	GetAPIKeyUsage(ctx context.Context, apiKeyID string, since time.Time) ([]APIUsage, error)
}

// EjectionRepository manages ejection data from Retrosheet.
type EjectionRepository interface {
	// List retrieves ejections based on filter criteria
	List(ctx context.Context, filter EjectionFilter) ([]Ejection, error)
	Count(ctx context.Context, filter EjectionFilter) (int, error)

	// ListBySeason retrieves all ejections for a specific season
	ListBySeason(ctx context.Context, year SeasonYear, p Pagination) ([]Ejection, error)
	CountBySeason(ctx context.Context, year SeasonYear) (int, error)
}
