package core

import "time"

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

type Pagination struct {
	Page    int
	PerPage int
}

func NewPagination(p, pp int) *Pagination {
	return &Pagination{Page: p, PerPage: pp}
}

type PlayerGameLogFilter struct {
	Season     *SeasonYear // Filter by season
	DateFrom   *string     // YYYYMMDD format
	DateTo     *string     // YYYYMMDD format
	Position   *int        // Filter by position (1-9, for fielding logs)
	MinHR      *int        // Minimum home runs
	MinH       *int        // Minimum hits
	MinRBI     *int        // Minimum RBI
	MinPA      *int        // Minimum plate appearances
	MinSO      *int        // Minimum strikeouts (pitching)
	MinIP      *float64    // Minimum innings pitched (pitching)
	Pagination Pagination
}

type PlayerFilter struct {
	NameQuery  string // partial name
	DebutYear  *SeasonYear
	TeamIDs    []TeamID
	Positions  []string
	Bats       []string // "R","L","B"
	Throws     []string
	SortBy     string // "name","debut","war" etc.
	SortOrder  SortOrder
	Pagination Pagination
}

type TeamFilter struct {
	NameQuery   string // partial name search
	Year        *SeasonYear
	League      *LeagueID
	Leagues     []LeagueID // multiple leagues filter
	FranchiseID *FranchiseID
	SortBy      string // "wins","run_diff"
	SortOrder   SortOrder
	Pagination  Pagination
}

type ParkFilter struct {
	NameQuery  string // partial name, city, or state search
	Pagination Pagination
}

type GameFilter struct {
	Season       *SeasonYear
	DateFrom     *time.Time
	DateTo       *time.Time
	HomeTeam     *TeamID
	AwayTeam     *TeamID
	ParkID       *ParkID
	League       *LeagueID
	Leagues      []LeagueID // multiple leagues filter
	GameType     *string
	IsPostseason *bool
	SortBy       string // "date"
	SortOrder    SortOrder
	Pagination   Pagination
}

type EventFilter struct {
	GameID     *GameID
	PlayerID   *PlayerID // batter or pitcher
	Season     *SeasonYear
	InningFrom *int
	InningTo   *int
	Pagination Pagination
}

type AwardFilter struct {
	AwardID    *AwardID
	PlayerID   *PlayerID
	Year       *SeasonYear
	League     *LeagueID
	Pagination Pagination
}

type SearchFilter struct {
	Query string
	Limit int
}

type BattingStatsFilter struct {
	PlayerID   *PlayerID
	TeamID     *TeamID
	Season     *SeasonYear
	SeasonFrom *SeasonYear
	SeasonTo   *SeasonYear
	League     *LeagueID
	MinAB      *int
	MinPA      *int
	SortBy     string // "avg", "hr", "rbi", "h", "r", "sb"
	SortOrder  SortOrder
	Pagination Pagination
}

type PitchingStatsFilter struct {
	PlayerID   *PlayerID
	TeamID     *TeamID
	Season     *SeasonYear
	SeasonFrom *SeasonYear
	SeasonTo   *SeasonYear
	League     *LeagueID
	MinIP      *float64
	MinGS      *int
	SortBy     string // "era", "w", "so", "sv", "ip"
	SortOrder  SortOrder
	Pagination Pagination
}

type PlayFilter struct {
	GameID     *GameID
	Batter     *RetroPlayerID
	Pitcher    *RetroPlayerID
	BatTeam    *TeamID
	PitTeam    *TeamID
	League     *LeagueID
	Leagues    []LeagueID
	GameType   *string
	Date       *string // YYYYMMDD format
	DateFrom   *string // YYYYMMDD format
	DateTo     *string // YYYYMMDD format
	Inning     *int
	HomeRuns   *bool  // if true, only return plays with HR=1
	Walks      *bool  // if true, only return plays with walk=1
	K          *bool  // if true, only return plays with k=1
	SortBy     string // "pn" (play number), "inning", "date"
	SortOrder  SortOrder
	Pagination Pagination
}

type FieldingStatsFilter struct {
	PlayerID   *PlayerID
	TeamID     *TeamID
	Season     *SeasonYear
	SeasonFrom *SeasonYear
	SeasonTo   *SeasonYear
	League     *LeagueID
	Position   *string // Filter by position
	MinG       *int    // Minimum games threshold
	SortBy     string  // "po", "a", "e", "dp", "fpct"
	SortOrder  SortOrder
	Pagination Pagination
}

type TeamStatsFilter struct {
	TeamID     *TeamID
	Season     *SeasonYear
	SeasonFrom *SeasonYear
	SeasonTo   *SeasonYear
	League     *LeagueID
	SortBy     string // stat-specific: "hr", "avg", "era", "so", "po", "e", etc.
	SortOrder  SortOrder
	Pagination Pagination
}

type TeamDailyStatsFilter struct {
	TeamID     *TeamID
	Season     *SeasonYear
	DateFrom   *string // YYYYMMDD format
	DateTo     *string // YYYYMMDD format
	Result     *string // Filter by result: W, L, or T
	SortBy     string  // "date", "runs", "runs_allowed"
	SortOrder  SortOrder
	Pagination Pagination
}

type EjectionFilter struct {
	Season     *SeasonYear
	PlayerID   *RetroPlayerID
	UmpireID   *UmpireID
	TeamID     *TeamID
	Role       *string // P, M, C
	DateFrom   *string // YYYYMMDD format
	DateTo     *string // YYYYMMDD format
	SortBy     string  // "date", "inning"
	SortOrder  SortOrder
	Pagination Pagination
}

type PitchFilter struct {
	GameID      *GameID
	Batter      *RetroPlayerID
	Pitcher     *RetroPlayerID
	BatTeam     *TeamID
	PitTeam     *TeamID
	Date        *string // YYYYMMDD format
	DateFrom    *string // YYYYMMDD format
	DateTo      *string // YYYYMMDD format
	Inning      *int
	TopBot      *int    // Filter by top (0) or bottom (1) of inning
	PitchType   *string // Single character: B, C, F, S, X, etc.
	BallCount   *int    // Filter by ball count (0-3)
	StrikeCount *int    // Filter by strike count (0-2)
	IsInPlay    *bool   // Filter for pitches in play (X)
	IsStrike    *bool   // Filter for strikes
	IsBall      *bool   // Filter for balls
	SortBy      string  // "seq", "date"
	SortOrder   SortOrder
	Pagination  Pagination
}

type AdvancedBattingFilter struct {
	Season     *SeasonYear
	SeasonFrom *SeasonYear
	SeasonTo   *SeasonYear
	TeamID     *TeamID
	League     *LeagueID
	MinPA      *int // Minimum plate appearances
	Split      *SplitDimension
	Provider   *StatProvider // fangraphs, bbref, internal
	ParkAdjust bool          // Apply park adjustments
	Pagination Pagination
}

type AdvancedPitchingFilter struct {
	Season     *SeasonYear
	SeasonFrom *SeasonYear
	SeasonTo   *SeasonYear
	TeamID     *TeamID
	League     *LeagueID
	MinIP      *float64 // Minimum innings pitched
	Split      *SplitDimension
	Provider   *StatProvider
	ParkAdjust bool
	Pagination Pagination
}

type WARFilter struct {
	Season     *SeasonYear
	SeasonFrom *SeasonYear
	SeasonTo   *SeasonYear
	TeamID     *TeamID
	League     *LeagueID
	MinPA      *int          // For position players
	MinIP      *float64      // For pitchers
	Provider   *StatProvider // fangraphs, bbref, internal
	Pagination Pagination
}

type AchievementFilter struct {
	Season     *SeasonYear
	SeasonFrom *SeasonYear // Filter from season (inclusive)
	SeasonTo   *SeasonYear // Filter to season (inclusive)
	TeamID     *TeamID
	PlayerID   *string //  Retrosheet player ID
	DateFrom   *time.Time
	DateTo     *time.Time
	ParkID     *ParkID
	MinHR      *int // Minimum home runs for multi-HR games
	MinInnings *int // Minimum innings for extra inning games
	Pagination Pagination
}
