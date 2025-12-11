package core

import "time"

type Pagination struct {
	Page    int
	PerPage int
}

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

type PlayerFilter struct {
	NameQuery string // partial name
	DebutYear *SeasonYear
	TeamIDs   []TeamID
	Positions []string
	Bats      []string // "R","L","B"
	Throws    []string

	SortBy    string // "name","debut","war" etc.
	SortOrder SortOrder

	Pagination Pagination
}

type TeamFilter struct {
	Year        *SeasonYear
	League      *LeagueID
	FranchiseID *FranchiseID

	SortBy     string // "wins","run_diff"
	SortOrder  SortOrder
	Pagination Pagination
}

type GameFilter struct {
	Season       *SeasonYear
	DateFrom     *time.Time
	DateTo       *time.Time
	HomeTeam     *TeamID
	AwayTeam     *TeamID
	ParkID       *ParkID
	IsPostseason *bool

	SortBy     string // "date"
	SortOrder  SortOrder
	Pagination Pagination
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
	AwardID  *AwardID
	PlayerID *PlayerID
	Year     *SeasonYear
	League   *LeagueID

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
