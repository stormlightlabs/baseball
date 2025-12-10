// TODO: complete models
package core

import (
	"time"
)

// SeasonYear represents a baseball season year
// @Description A baseball season year
type SeasonYear int

// PlayerID is the Lahman player ID (e.g., "troutmi01")
// @Description Lahman player identifier
type PlayerID string

// RetroPlayerID is the Retrosheet player ID
// @Description Retrosheet player identifier
type RetroPlayerID string

// TeamID is the Lahman team ID (e.g., "LAA")
// @Description Lahman team identifier
type TeamID string

// FranchiseID is the Lahman franchise ID (e.g., "ANA")
// @Description Lahman franchise identifier
type FranchiseID string

// GameID is the Retrosheet game ID (e.g., "ANA201304010")
// @Description Retrosheet game identifier
type GameID string

// ParkID is the Retrosheet/Lahman park code
// @Description Park identifier
type ParkID string

// ManagerID is the manager ID in Lahman/Retrosheet
// @Description Manager identifier
type ManagerID string

// UmpireID is the Retrosheet umpire ID
// @Description Umpire identifier
type UmpireID string

// AwardID is the Lahman award ID
// @Description Award identifier
type AwardID string

// LeagueID is the league identifier (e.g., "AL", "NL")
// @Description League identifier (AL, NL, etc.)
type LeagueID string

// Player is a person row + a few commonly needed derived fields.
// Based primarily on Lahman People table. :contentReference[oaicite:1]{index=1}
type Player struct {
	ID        PlayerID       `json:"id" swaggertype:"string" example:"ruthba01"`
	RetroID   *RetroPlayerID `json:"retro_id,omitempty" swaggertype:"string" example:"ruthb101"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	GivenName string         `json:"given_name"` // e.g., "Michael Nelson"
	Nickname  *string        `json:"nickname,omitempty"`

	BirthYear    *int    `json:"birth_year,omitempty"`
	BirthMonth   *int    `json:"birth_month,omitempty"`
	BirthDay     *int    `json:"birth_day,omitempty"`
	BirthCity    *string `json:"birth_city,omitempty"`
	BirthState   *string `json:"birth_state,omitempty"`
	BirthCountry *string `json:"birth_country,omitempty"`

	DeathYear    *int    `json:"death_year,omitempty"`
	DeathMonth   *int    `json:"death_month,omitempty"`
	DeathDay     *int    `json:"death_day,omitempty"`
	DeathCity    *string `json:"death_city,omitempty"`
	DeathState   *string `json:"death_state,omitempty"`
	DeathCountry *string `json:"death_country,omitempty"`

	Bats   *string `json:"bats,omitempty"`   // "R", "L", "B"
	Throws *string `json:"throws,omitempty"` // "R", "L"

	HeightInches *int `json:"height_inches,omitempty"`
	WeightLbs    *int `json:"weight_lbs,omitempty"`

	Debut     *time.Time `json:"debut,omitempty"`
	FinalGame *time.Time `json:"final_game,omitempty"`
}

// PlayerBattingSeason is a single season/team batting line.
// Loosely mapped from Lahman Batting table. :contentReference[oaicite:2]{index=2}
type PlayerBattingSeason struct {
	PlayerID PlayerID   `json:"player_id" swaggertype:"string"`
	Year     SeasonYear `json:"year" swaggertype:"integer"`
	TeamID   TeamID     `json:"team_id" swaggertype:"string"`
	League   LeagueID   `json:"league" swaggertype:"string"`

	G       int `json:"g"`
	PA      int `json:"pa"`
	AB      int `json:"ab"`
	R       int `json:"r"`
	H       int `json:"h"`
	Doubles int `json:"doubles"`
	Triples int `json:"triples"`
	HR      int `json:"hr"`
	RBI     int `json:"rbi"`
	SB      int `json:"sb"`
	CS      int `json:"cs"`
	BB      int `json:"bb"`
	SO      int `json:"so"`
	HBP     int `json:"hbp"`
	SF      int `json:"sf"`

	// Precomputed rate stats (optional but convenient)
	AVG float64 `json:"avg"`
	OBP float64 `json:"obp"`
	SLG float64 `json:"slg"`
	OPS float64 `json:"ops"`
}

// PlayerPitchingSeason mapped from Lahman Pitching.
type PlayerPitchingSeason struct {
	PlayerID PlayerID   `json:"player_id" swaggertype:"string"`
	Year     SeasonYear `json:"year" swaggertype:"integer"`
	TeamID   TeamID     `json:"team_id" swaggertype:"string"`
	League   LeagueID   `json:"league" swaggertype:"string"`

	W      int `json:"w"`
	L      int `json:"l"`
	G      int `json:"g"`
	GS     int `json:"gs"`
	CG     int `json:"cg"`
	SHO    int `json:"sho"`
	SV     int `json:"sv"`
	IPOuts int `json:"ip_outs"` // 3 * IP
	H      int `json:"h"`
	ER     int `json:"er"`
	HR     int `json:"hr"`
	BB     int `json:"bb"`
	SO     int `json:"so"`
	HBP    int `json:"hbp"`
	BK     int `json:"bk"`
	WP     int `json:"wp"`

	ERA    float64  `json:"era"`
	WHIP   float64  `json:"whip"`
	KPer9  float64  `json:"k_per_9"`
	BBPer9 float64  `json:"bb_per_9"`
	HRPer9 float64  `json:"hr_per_9"`
	FIP    *float64 `json:"fip,omitempty"` // if you compute it
}

// PlayerFieldingSeason from Lahman Fielding. :contentReference[oaicite:3]{index=3}
type PlayerFieldingSeason struct {
	PlayerID PlayerID   `json:"player_id"`
	Year     SeasonYear `json:"year"`
	TeamID   TeamID     `json:"team_id"`
	League   LeagueID   `json:"league"`
	Position string     `json:"position"`

	G   int `json:"g"`
	GS  int `json:"gs"`
	Inn int `json:"inn_outs"` // or innings * 3

	PO  int     `json:"po"`
	A   int     `json:"a"`
	E   int     `json:"e"`
	DP  int     `json:"dp"`
	PB  int     `json:"pb"` // catchers
	SB  int     `json:"sb"`
	CS  int     `json:"cs"`
	RF9 float64 `json:"rf9"` // range factor per 9 (if computed)
}

// TeamSeason corresponds to a row in Lahman Teams. :contentReference[oaicite:4]{index=4}
type TeamSeason struct {
	TeamID      TeamID      `json:"team_id" swaggertype:"string"`
	Year        SeasonYear  `json:"year" swaggertype:"integer"`
	FranchiseID FranchiseID `json:"franchise_id" swaggertype:"string"`
	League      LeagueID    `json:"league" swaggertype:"string"`

	Name     string  `json:"name"`
	ParkID   ParkID  `json:"park_id" swaggertype:"string"`
	Games    int     `json:"games"`
	Wins     int     `json:"wins"`
	Losses   int     `json:"losses"`
	Ties     int     `json:"ties"`
	Division *string `json:"division,omitempty"`

	RunsScored  int `json:"runs_scored"`
	RunsAllowed int `json:"runs_allowed"`

	Attendance *int `json:"attendance,omitempty"`
}

// Franchise from Lahman TeamsFranchises.
type Franchise struct {
	ID         FranchiseID `json:"id" swaggertype:"string"`
	Name       string      `json:"name"`
	Active     bool        `json:"active"`
	ActiveFrom *SeasonYear `json:"active_from,omitempty" swaggertype:"integer"`
	ActiveTo   *SeasonYear `json:"active_to,omitempty" swaggertype:"integer"`
}

// Game combines Retrosheet’s game-log level data. :contentReference[oaicite:5]{index=5}
type Game struct {
	ID        GameID     `json:"id"`
	Season    SeasonYear `json:"season"`
	Date      time.Time  `json:"date"`
	DayOfWeek string     `json:"day_of_week"`

	HomeTeam   TeamID   `json:"home_team"`
	AwayTeam   TeamID   `json:"away_team"`
	HomeLeague LeagueID `json:"home_league"`
	AwayLeague LeagueID `json:"away_league"`

	HomeScore int `json:"home_score"`
	AwayScore int `json:"away_score"`
	Innings   int `json:"innings"`

	Attendance  *int `json:"attendance,omitempty"`
	DurationMin *int `json:"duration_min,omitempty"`

	ParkID    ParkID    `json:"park_id"`
	UmpHome   *UmpireID `json:"ump_home,omitempty"`
	UmpFirst  *UmpireID `json:"ump_first,omitempty"`
	UmpSecond *UmpireID `json:"ump_second,omitempty"`
	UmpThird  *UmpireID `json:"ump_third,omitempty"`

	// Postseason flags etc.
	IsPostseason bool    `json:"is_postseason"`
	SeriesID     *string `json:"series_id,omitempty"`
	GameInSeries *int    `json:"game_in_series,omitempty"`
}

// GameEvent corresponds to a row in Retrosheet plays.csv (simplified). :contentReference[oaicite:6]{index=6}
type GameEvent struct {
	GameID      GameID `json:"game_id"`
	EventID     int    `json:"event_id"` // sequential within game
	Inning      int    `json:"inning"`
	TopOfInning bool   `json:"top_of_inning"`

	BattingTeam  TeamID   `json:"batting_team"`
	FieldingTeam TeamID   `json:"fielding_team"`
	Batter       PlayerID `json:"batter"`
	Pitcher      PlayerID `json:"pitcher"`
	// Optional: catcher, runner IDs as needed.

	OutsBefore   int `json:"outs_before"`
	OutsAfter    int `json:"outs_after"`
	RunsScored   int `json:"runs_scored"`
	RunsBattedIn int `json:"runs_batted_in"`

	// Base/out state; there are many ways to encode this; here’s a simple one.
	RunnerOnFirst  *PlayerID `json:"runner_on_first,omitempty"`
	RunnerOnSecond *PlayerID `json:"runner_on_second,omitempty"`
	RunnerOnThird  *PlayerID `json:"runner_on_third,omitempty"`

	EventText   string  `json:"event_text"`            // e.g. original Retrosheet event text
	Description *string `json:"description,omitempty"` // cleaned description
}

// Park / Ballpark from Lahman & Retrosheet park tables. :contentReference[oaicite:7]{index=7}
type Park struct {
	ID        ParkID `json:"id"`
	Name      string `json:"name"`
	City      string `json:"city"`
	State     string `json:"state"`
	Country   string `json:"country"`
	StartYear *int   `json:"start_year,omitempty"`
	EndYear   *int   `json:"end_year,omitempty"`
}

// Manager (Lahman Managers).
type Manager struct {
	ID        ManagerID `json:"id"`
	PlayerID  *PlayerID `json:"player_id,omitempty"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

// ManagerSeasonRecord is a season-level line for a manager.
type ManagerSeasonRecord struct {
	ManagerID ManagerID  `json:"manager_id"`
	TeamID    TeamID     `json:"team_id"`
	Year      SeasonYear `json:"year"`
	G         int        `json:"g"`
	W         int        `json:"w"`
	L         int        `json:"l"`
	Rank      *int       `json:"rank,omitempty"`
}

// Umpire basic info.
type Umpire struct {
	ID        UmpireID `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
}

// Award and award results from Lahman Awards*. :contentReference[oaicite:8]{index=8}
type Award struct {
	ID          AwardID `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
}

type AwardResult struct {
	AwardID    AwardID    `json:"award_id" swaggertype:"string"`
	PlayerID   PlayerID   `json:"player_id" swaggertype:"string"`
	Year       SeasonYear `json:"year" swaggertype:"integer"`
	League     *LeagueID  `json:"league,omitempty" swaggertype:"string"`
	VotesFirst *int       `json:"votes_first,omitempty"`
	Points     *int       `json:"points,omitempty"`
	Rank       *int       `json:"rank,omitempty"`
}

// HallOfFameRecord based on Lahman HallOfFame table.
type HallOfFameRecord struct {
	PlayerID PlayerID   `json:"player_id" swaggertype:"string"`
	Year     SeasonYear `json:"year" swaggertype:"integer"`
	VotedBy  string     `json:"voted_by"`
	Votes    *int       `json:"votes,omitempty"`
	Ballots  *int       `json:"ballots,omitempty"`
	Needed   *int       `json:"needed,omitempty"`
	Inducted bool       `json:"inducted"`
}
