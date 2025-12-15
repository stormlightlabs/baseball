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

	// BattingGameLogs returns per-game batting statistics for a player
	// Derived from player_game_batting_stats materialized view
	BattingGameLogs(ctx context.Context, id PlayerID, filter PlayerGameLogFilter) ([]PlayerGameBattingLog, error)
	CountBattingGameLogs(ctx context.Context, id PlayerID, filter PlayerGameLogFilter) (int, error)

	// PitchingGameLogs returns per-game pitching statistics for a player
	// Derived from player_game_pitching_stats materialized view
	PitchingGameLogs(ctx context.Context, id PlayerID, filter PlayerGameLogFilter) ([]PlayerGamePitchingLog, error)
	CountPitchingGameLogs(ctx context.Context, id PlayerID, filter PlayerGameLogFilter) (int, error)

	// FieldingGameLogs returns per-game fielding statistics for a player
	// Derived from player_game_fielding_stats materialized view
	// If position filter is set, returns only games at that position
	FieldingGameLogs(ctx context.Context, id PlayerID, filter PlayerGameLogFilter) ([]PlayerGameFieldingLog, error)
	CountFieldingGameLogs(ctx context.Context, id PlayerID, filter PlayerGameLogFilter) (int, error)

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

	// Per-game team statistics for daily performance tracking
	DailyStats(ctx context.Context, filter TeamDailyStatsFilter) ([]TeamDailyStats, error)
	CountDailyStats(ctx context.Context, filter TeamDailyStatsFilter) (int, error)
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

	// SearchGamesNL performs a natural language search for games
	SearchGamesNL(ctx context.Context, query string, limit int) ([]Game, error)

	// ResolveTeamAlias looks up a team ID from a team name alias
	ResolveTeamAlias(ctx context.Context, alias string, season *int) (TeamID, bool)
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

// ParkRepository for ballpark metadata and usage.
type ParkRepository interface {
	GetByID(ctx context.Context, id ParkID) (*Park, error)
	List(ctx context.Context, filter ParkFilter) ([]Park, error)
	Count(ctx context.Context, filter ParkFilter) (int, error)

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

	// WOBAConstants returns wOBA calculation constants for a specific season.
	WOBAConstants(ctx context.Context, season *SeasonYear) ([]WOBAConstant, error)

	// LeagueConstants returns league-specific constants for wRC+ calculations.
	LeagueConstants(ctx context.Context, season *SeasonYear, league *LeagueID) ([]LeagueConstant, error)

	// ParkFactors returns park factors for all parks in a season.
	ParkFactors(ctx context.Context, season *SeasonYear, teamID *TeamID) ([]ParkFactorRow, error)
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

// PitchRepository manages individual pitch data derived from Retrosheet play-by-play sequences.
type PitchRepository interface {
	// List retrieves pitches based on filter criteria, parsing pitch sequences from plays
	List(ctx context.Context, filter PitchFilter) ([]Pitch, error)
	Count(ctx context.Context, filter PitchFilter) (int, error)

	// ListByGame retrieves all pitches for a specific game in order
	ListByGame(ctx context.Context, gameID GameID, p Pagination) ([]Pitch, error)

	// ListByPlay retrieves all pitches from a specific plate appearance
	ListByPlay(ctx context.Context, gameID GameID, playNum int) ([]Pitch, error)
}

// DerivedStatsRepository manages advanced computed analytics from play-by-play data.
type DerivedStatsRepository interface {
	// PlayerStreaks retrieves hitting or scoreless innings streaks for a player
	PlayerStreaks(ctx context.Context, playerID PlayerID, kind StreakKind, season SeasonYear, minLength int) ([]Streak, error)

	// TeamRunDifferential calculates season run differential with rolling windows
	TeamRunDifferential(ctx context.Context, teamID TeamID, season SeasonYear, windows []int) (*RunDifferentialSeries, error)

	// GameWinProbability returns win probability curve for a game
	GameWinProbability(ctx context.Context, gameID GameID) (*WinProbabilityCurve, error)

	// PlayerSplits calculates batting splits for a player by dimension
	PlayerSplits(ctx context.Context, playerID PlayerID, dimension SplitDimension, season SeasonYear) (*SplitResult, error)
}

// AdvancedStatsRepository computes sabermetric stats (wOBA, wRC+, FIP, WAR, etc.).
// These calculations require both Lahman stats and FanGraphs constants (year-specific weights).
type AdvancedStatsRepository interface {
	// PlayerAdvancedBatting computes wOBA, wRC+, ISO, BABIP, etc. for a player
	// Supports splits by home/away, pitcher handedness, month, etc.
	PlayerAdvancedBatting(ctx context.Context, playerID PlayerID, filter AdvancedBattingFilter) (*AdvancedBattingStats, error)

	// PlayerAdvancedBattingSplits returns advanced batting stats split by dimension
	PlayerAdvancedBattingSplits(ctx context.Context, playerID PlayerID, filter AdvancedBattingFilter) ([]AdvancedBattingStats, error)

	// PlayerAdvancedPitching computes FIP, xFIP, ERA+, etc. for a pitcher
	PlayerAdvancedPitching(ctx context.Context, playerID PlayerID, filter AdvancedPitchingFilter) (*AdvancedPitchingStats, error)

	// PlayerBaserunning calculates base running runs (wSB) for a player
	PlayerBaserunning(ctx context.Context, playerID PlayerID, season SeasonYear, teamID *TeamID) (*BaserunningStats, error)

	// PlayerFielding calculates fielding runs for a player using range factor
	PlayerFielding(ctx context.Context, playerID PlayerID, season SeasonYear, teamID *TeamID) (*FieldingStats, error)

	// PlayerWAR computes WAR components and total WAR for a player
	// Provider indicates which formula to use (fangraphs, bbref, internal)
	PlayerWAR(ctx context.Context, playerID PlayerID, filter WARFilter) (*PlayerWARSummary, error)

	// SeasonBattingLeaders returns top N players by advanced batting stat
	SeasonBattingLeaders(ctx context.Context, season SeasonYear, stat string, limit int, filter AdvancedBattingFilter) ([]AdvancedBattingStats, error)

	// SeasonPitchingLeaders returns top N pitchers by advanced pitching stat
	SeasonPitchingLeaders(ctx context.Context, season SeasonYear, stat string, limit int, filter AdvancedPitchingFilter) ([]AdvancedPitchingStats, error)

	// SeasonWARLeaders returns top N players by WAR
	SeasonWARLeaders(ctx context.Context, season SeasonYear, limit int, filter WARFilter) ([]PlayerWARSummary, error)
}

// LeverageRepository computes leverage index and win probability metrics from play-by-play data.
type LeverageRepository interface {
	// GamePlateLeverages returns leverage index for each plate appearance in a game
	GamePlateLeverages(ctx context.Context, gameID GameID, minLI *float64) ([]PlateAppearanceLeverage, error)

	// PlayerLeverageSummary aggregates leverage metrics for a player (aLI, gmLI, etc.)
	// Most commonly used for pitchers to understand usage patterns
	PlayerLeverageSummary(ctx context.Context, playerID PlayerID, season SeasonYear, role string) (*PlayerLeverageSummary, error)

	// PlayerHighLeveragePAs returns high-leverage plate appearances for a player
	PlayerHighLeveragePAs(ctx context.Context, playerID PlayerID, season SeasonYear, minLI float64) ([]PlateAppearanceLeverage, error)

	// GameWinProbabilitySummary returns summary stats for a game's win probability
	GameWinProbabilitySummary(ctx context.Context, gameID GameID) (*GameWinProbabilitySummary, error)
}

// ParkFactorRepository computes and retrieves park factors.
// Park factors measure how hitter/pitcher-friendly a ballpark is relative to league average.
type ParkFactorRepository interface {
	// ParkFactor returns park factors for a specific park and season
	// Includes runs, HR, BB, and hits factors
	ParkFactor(ctx context.Context, parkID ParkID, season SeasonYear) (*ParkFactor, error)

	// ParkFactorSeries returns park factors over a range of seasons
	// Supports multi-year averaging for stability
	ParkFactorSeries(ctx context.Context, parkID ParkID, fromSeason, toSeason SeasonYear) ([]ParkFactor, error)

	// SeasonParkFactors returns all park factors for a given season
	// Optionally filter by factor type (runs, hr, etc.)
	SeasonParkFactors(ctx context.Context, season SeasonYear, factorType *string) ([]ParkFactor, error)

	// MultiYearParkFactor returns a park factor averaged over multiple seasons
	// Used for more stable estimates
	MultiYearParkFactor(ctx context.Context, parkID ParkID, fromSeason, toSeason SeasonYear) (*ParkFactor, error)
}

// WinExpectancyRepository manages historical win expectancy data used for leverage index calculations.
// Win expectancy represents the probability that the home team wins from a given game state,
// based on historical outcomes from similar situations.
type WinExpectancyRepository interface {
	// GetWinExpectancy returns the win probability for a specific game state
	// Uses the most appropriate historical era if era parameters are not specified
	GetWinExpectancy(ctx context.Context, state GameState) (*WinExpectancy, error)

	// GetWinExpectancyForEra returns win probability for a specific game state within a time period
	GetWinExpectancyForEra(ctx context.Context, state GameState, startYear, endYear *int) (*WinExpectancy, error)

	// BatchGetWinExpectancy efficiently retrieves win expectancies for multiple game states
	// Useful for computing leverage index across a full game
	BatchGetWinExpectancy(ctx context.Context, states []GameState) ([]WinExpectancy, error)

	// ListAvailableEras returns all available historical eras in the win expectancy table
	ListAvailableEras(ctx context.Context) ([]WinExpectancyEra, error)

	// UpsertWinExpectancy inserts or updates win expectancy data
	// Used for populating the table from historical analysis
	UpsertWinExpectancy(ctx context.Context, we *WinExpectancy) error

	// BuildFromHistoricalData computes win expectancies from play-by-play data for a given era
	// This is a heavy operation that analyzes all games in the specified year range
	BuildFromHistoricalData(ctx context.Context, startYear, endYear int, minSampleSize int) (int64, error)
}

// NegroLeaguesRepository provides access to Negro Leagues historical data (1903-1962).
// Negro Leagues data is stored in the main games and plays tables with specific game_type markers.
type NegroLeaguesRepository interface {
	// ListGames returns Negro Leagues games with filtering and pagination
	ListGames(ctx context.Context, filter GameFilter) ([]Game, error)
	CountGames(ctx context.Context, filter GameFilter) (int, error)

	// ListTeamSeasons returns teams that played in the Negro Leagues
	ListTeamSeasons(ctx context.Context, filter TeamFilter) ([]TeamSeason, error)
	CountTeamSeasons(ctx context.Context, filter TeamFilter) (int, error)

	// ListPlays returns play-by-play data from Negro Leagues games
	ListPlays(ctx context.Context, filter PlayFilter) ([]Play, error)
	CountPlays(ctx context.Context, filter PlayFilter) (int, error)

	// GetSeasonSchedule returns all Negro Leagues games for a specific season
	GetSeasonSchedule(ctx context.Context, year SeasonYear, league *LeagueID, p Pagination) ([]Game, error)

	// GetTeamGames returns all games for a specific team in a season
	GetTeamGames(ctx context.Context, teamID TeamID, year SeasonYear, p Pagination) ([]Game, error)
}
