// Note that Retrosheet's `plays` CSV exposes 160+ fields (teams, lineup slots, ball-in-play classification, baserunner state, umpire crew, etc.).
// All ingestion follows the Retrosheet parsed [play-by-play] specification
//
// [play-by-play]: https://retrosheet.org/downloads/plays.html
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
// Based primarily on Lahman People table.
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
// Loosely mapped from Lahman Batting table.
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
	FIP    *float64 `json:"fip,omitempty"`
}

// PlayerFieldingSeason from Lahman Fielding.
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
	RF9 float64 `json:"rf9"`
}

// TeamSeason corresponds to a row in Lahman Teams.
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

// Game combines Retrosheet’s game-log level data.
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

	IsPostseason bool    `json:"is_postseason"`
	SeriesID     *string `json:"series_id,omitempty"`
	GameInSeries *int    `json:"game_in_series,omitempty"`
}

// Boxscore contains detailed team and player statistics for a game
type Boxscore struct {
	GameID    GameID    `json:"game_id"`
	Date      time.Time `json:"date"`
	HomeTeam  TeamID    `json:"home_team"`
	AwayTeam  TeamID    `json:"away_team"`
	HomeScore int       `json:"home_score"`
	AwayScore int       `json:"away_score"`

	HomeStats  TeamGameStats  `json:"home_stats"`
	AwayStats  TeamGameStats  `json:"away_stats"`
	HomeLineup []LineupPlayer `json:"home_lineup,omitempty"`
	AwayLineup []LineupPlayer `json:"away_lineup,omitempty"`
}

// TeamGameStats contains aggregate statistics for a team in a game
type TeamGameStats struct {
	AB      int `json:"ab"`
	H       int `json:"h"`
	R       int `json:"r"`
	Doubles int `json:"doubles"`
	Triples int `json:"triples"`
	HR      int `json:"hr"`
	RBI     int `json:"rbi"`
	SH      int `json:"sh"`
	SF      int `json:"sf"`
	HBP     int `json:"hbp"`
	BB      int `json:"bb"`
	IBB     int `json:"ibb"`
	SO      int `json:"so"`
	SB      int `json:"sb"`
	CS      int `json:"cs"`
	GDP     int `json:"gdp"`
	LOB     int `json:"lob"`

	PitchersUsed int `json:"pitchers_used"`
	ER           int `json:"er"`
	WP           int `json:"wp"`
	Balks        int `json:"balks"`

	PO int `json:"po"`
	A  int `json:"a"`
	E  int `json:"e"`
	PB int `json:"pb"`
	DP int `json:"dp"`
	TP int `json:"tp"`
}

// LineupPlayer represents a starting lineup player with position
type LineupPlayer struct {
	PlayerID PlayerID `json:"player_id"`
	Name     string   `json:"name,omitempty"`
	Position int      `json:"position"`
}

// PlayerAppearance represents a player's appearance data for a season
type PlayerAppearance struct {
	PlayerID PlayerID   `json:"player_id"`
	Year     SeasonYear `json:"year"`
	TeamID   TeamID     `json:"team_id"`
	League   *LeagueID  `json:"league,omitempty"`

	GamesAll     int `json:"g_all"`
	GamesStarted int `json:"gs"`
	GBatting     int `json:"g_batting"`
	GDefense     int `json:"g_defense"`

	GP  int `json:"g_p"`  // Pitcher
	GC  int `json:"g_c"`  // Catcher
	G1B int `json:"g_1b"` // First base
	G2B int `json:"g_2b"` // Second base
	G3B int `json:"g_3b"` // Third base
	GSS int `json:"g_ss"` // Shortstop
	GLF int `json:"g_lf"` // Left field
	GCF int `json:"g_cf"` // Center field
	GRF int `json:"g_rf"` // Right field
	GOF int `json:"g_of"` // Outfield
	GDH int `json:"g_dh"` // Designated hitter
	GPH int `json:"g_ph"` // Pinch hitter
	GPR int `json:"g_pr"` // Pinch runner
}

// PlayerTeamSeason represents the teams and seasons a player suited up for.
type PlayerTeamSeason struct {
	PlayerID     PlayerID   `json:"player_id"`
	Year         SeasonYear `json:"year"`
	TeamID       TeamID     `json:"team_id"`
	TeamName     *string    `json:"team_name,omitempty"`
	League       *LeagueID  `json:"league,omitempty"`
	Games        int        `json:"games"`
	GamesStarted int        `json:"games_started"`
}

// PlayerSalary captures a single salary record for a player.
type PlayerSalary struct {
	PlayerID PlayerID   `json:"player_id"`
	Year     SeasonYear `json:"year"`
	TeamID   TeamID     `json:"team_id"`
	League   *LeagueID  `json:"league,omitempty"`
	Salary   int64      `json:"salary"`
}

// Play represents a single play from Retrosheet play-by-play data.
// This is the core model with essential fields for most use cases.
type Play struct {
	// Game identification
	GameID   GameID `json:"game_id"`
	PlayNum  int    `json:"play_num"`
	Inning   int    `json:"inning"`
	TopBot   int    `json:"top_bot"` // 0=top, 1=bottom
	BatTeam  TeamID `json:"bat_team"`
	PitTeam  TeamID `json:"pit_team"`
	Date     string `json:"date"`
	GameType string `json:"game_type"`

	Batter  RetroPlayerID `json:"batter"`
	Pitcher RetroPlayerID `json:"pitcher"`
	BatHand *string       `json:"bat_hand,omitempty"`
	PitHand *string       `json:"pit_hand,omitempty"`

	ScoreVis  int `json:"score_vis"`
	ScoreHome int `json:"score_home"`
	OutsPre   int `json:"outs_pre"`
	OutsPost  int `json:"outs_post"`

	Balls   *int    `json:"balls,omitempty"`
	Strikes *int    `json:"strikes,omitempty"`
	Pitches *string `json:"pitches,omitempty"`

	Event string `json:"event"` // The play description

	PA     *int `json:"pa,omitempty"` // Plate appearances
	AB     *int `json:"ab,omitempty"`
	Single *int `json:"single,omitempty"`
	Double *int `json:"double,omitempty"`
	Triple *int `json:"triple,omitempty"`
	HR     *int `json:"hr,omitempty"`
	Walk   *int `json:"walk,omitempty"`
	K      *int `json:"k,omitempty"`
	HBP    *int `json:"hbp,omitempty"`

	Runner1Pre *RetroPlayerID `json:"runner1_pre,omitempty"`
	Runner2Pre *RetroPlayerID `json:"runner2_pre,omitempty"`
	Runner3Pre *RetroPlayerID `json:"runner3_pre,omitempty"`

	Runs *int `json:"runs,omitempty"`
	RBI  *int `json:"rbi,omitempty"`
}

// Park / Ballpark from Lahman & Retrosheet park tables.
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

// Award and award results from Lahman Awards*.
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

// AllStarAppearance represents a player's appearance in an all-star game.
// Based on Lahman AllstarFull table.
type AllStarAppearance struct {
	PlayerID    PlayerID   `json:"player_id" swaggertype:"string"`
	Year        SeasonYear `json:"year" swaggertype:"integer"`
	GameNum     int        `json:"game_num"`
	GameID      string     `json:"game_id"`
	TeamID      *TeamID    `json:"team_id,omitempty" swaggertype:"string"`
	League      *LeagueID  `json:"league,omitempty" swaggertype:"string"`
	GP          *int       `json:"gp,omitempty"`
	StartingPos *int       `json:"starting_pos,omitempty"`
}

// AllStarGame aggregates all-star game information for a given year.
type AllStarGame struct {
	Year           SeasonYear          `json:"year" swaggertype:"integer"`
	GameNum        int                 `json:"game_num"`
	GameID         string              `json:"game_id"`
	Date           *time.Time          `json:"date,omitempty"`
	Venue          *ParkID             `json:"venue,omitempty" swaggertype:"string"`
	RetrosheetGame *Game               `json:"retrosheet_game,omitempty"`
	Participants   []AllStarAppearance `json:"participants,omitempty"`
}

// PostseasonSeries represents a postseason series from Lahman SeriesPost table.
type PostseasonSeries struct {
	Year         SeasonYear `json:"year" swaggertype:"integer"`
	Round        string     `json:"round"`
	WinnerTeam   *TeamID    `json:"winner_team,omitempty" swaggertype:"string"`
	WinnerLeague *LeagueID  `json:"winner_league,omitempty" swaggertype:"string"`
	LoserTeam    *TeamID    `json:"loser_team,omitempty" swaggertype:"string"`
	LoserLeague  *LeagueID  `json:"loser_league,omitempty" swaggertype:"string"`
	Wins         *int       `json:"wins,omitempty"`
	Losses       *int       `json:"losses,omitempty"`
	Ties         *int       `json:"ties,omitempty"`
}

// DatasetStatus surfaces ETL/coverage metadata for a dataset.
type DatasetStatus struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Source       string           `json:"source"`
	CoverageFrom *SeasonYear      `json:"coverage_from,omitempty"`
	CoverageTo   *SeasonYear      `json:"coverage_to,omitempty"`
	LastLoadedAt *time.Time       `json:"last_loaded_at,omitempty"`
	RowCount     int64            `json:"row_count"`
	Tables       map[string]int64 `json:"tables,omitempty"`
}

// Season represents summary information about a season
type Season struct {
	Year      SeasonYear `json:"year" swaggertype:"integer"`
	Leagues   []LeagueID `json:"leagues" swaggertype:"array,string"`
	TeamCount int        `json:"team_count"`
	GameCount *int       `json:"game_count,omitempty"`
}

// RosterPlayer represents a player on a team roster with basic stats
type RosterPlayer struct {
	PlayerID  PlayerID `json:"player_id" swaggertype:"string"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Position  *string  `json:"position,omitempty"`

	BattingG *int     `json:"batting_g,omitempty"`
	AB       *int     `json:"ab,omitempty"`
	H        *int     `json:"h,omitempty"`
	HR       *int     `json:"hr,omitempty"`
	RBI      *int     `json:"rbi,omitempty"`
	AVG      *float64 `json:"avg,omitempty"`

	PitchingG *int     `json:"pitching_g,omitempty"`
	W         *int     `json:"w,omitempty"`
	L         *int     `json:"l,omitempty"`
	ERA       *float64 `json:"era,omitempty"`
	SO        *int     `json:"so,omitempty"`
}

// TeamBattingStats represents aggregated batting statistics for a team
type TeamBattingStats struct {
	TeamID TeamID     `json:"team_id" swaggertype:"string"`
	Year   SeasonYear `json:"year" swaggertype:"integer"`
	League LeagueID   `json:"league" swaggertype:"string"`

	G       int     `json:"g"`
	AB      int     `json:"ab"`
	R       int     `json:"r"`
	H       int     `json:"h"`
	Doubles int     `json:"doubles"`
	Triples int     `json:"triples"`
	HR      int     `json:"hr"`
	RBI     int     `json:"rbi"`
	SB      int     `json:"sb"`
	CS      int     `json:"cs"`
	BB      int     `json:"bb"`
	SO      int     `json:"so"`
	HBP     int     `json:"hbp"`
	SF      int     `json:"sf"`
	AVG     float64 `json:"avg"`
	OBP     float64 `json:"obp"`
	SLG     float64 `json:"slg"`
	OPS     float64 `json:"ops"`

	Players []PlayerBattingSeason `json:"players,omitempty"`
}

// TeamPitchingStats represents aggregated pitching statistics for a team
type TeamPitchingStats struct {
	TeamID TeamID     `json:"team_id" swaggertype:"string"`
	Year   SeasonYear `json:"year" swaggertype:"integer"`
	League LeagueID   `json:"league" swaggertype:"string"`

	W      int     `json:"w"`
	L      int     `json:"l"`
	G      int     `json:"g"`
	GS     int     `json:"gs"`
	CG     int     `json:"cg"`
	SHO    int     `json:"sho"`
	SV     int     `json:"sv"`
	IPOuts int     `json:"ip_outs"`
	H      int     `json:"h"`
	ER     int     `json:"er"`
	HR     int     `json:"hr"`
	BB     int     `json:"bb"`
	SO     int     `json:"so"`
	ERA    float64 `json:"era"`
	WHIP   float64 `json:"whip"`

	Players []PlayerPitchingSeason `json:"players,omitempty"`
}

// TeamFieldingStats represents aggregated fielding statistics for a team
type TeamFieldingStats struct {
	TeamID TeamID     `json:"team_id" swaggertype:"string"`
	Year   SeasonYear `json:"year" swaggertype:"integer"`
	League LeagueID   `json:"league" swaggertype:"string"`

	G    int     `json:"g"`
	PO   int     `json:"po"`
	A    int     `json:"a"`
	E    int     `json:"e"`
	DP   int     `json:"dp"`
	PB   int     `json:"pb"`
	SB   int     `json:"sb"`
	CS   int     `json:"cs"`
	FPct float64 `json:"fpct"` // Fielding percentage

	Players []PlayerFieldingSeason `json:"players,omitempty"`
}

// UserID is a unique identifier for a user
type UserID string

// User represents an authenticated user in the system
type User struct {
	ID          UserID     `json:"id"`
	Email       string     `json:"email"`
	Name        *string    `json:"name,omitempty"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	IsActive    bool       `json:"is_active"`
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID         string     `json:"id"`
	UserID     UserID     `json:"user_id"`
	KeyPrefix  string     `json:"key_prefix"`
	Name       *string    `json:"name,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	IsActive   bool       `json:"is_active"`
}

// OAuthToken represents an OAuth2 token for session management
type OAuthToken struct {
	ID           string    `json:"id"`
	UserID       UserID    `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken *string   `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// APIUsage represents a single API request for usage tracking
type APIUsage struct {
	ID             int64     `json:"id"`
	UserID         *UserID   `json:"user_id,omitempty"`
	APIKeyID       *string   `json:"api_key_id,omitempty"`
	Endpoint       string    `json:"endpoint"`
	Method         string    `json:"method"`
	StatusCode     int       `json:"status_code"`
	ResponseTimeMs *int      `json:"response_time_ms,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// Ejection represents an ejection event from Retrosheet data.
type Ejection struct {
	GameID      GameID        `json:"game_id" swaggertype:"string"`
	Date        string        `json:"date"`
	GameNumber  *int          `json:"game_number,omitempty"`
	EjecteeID   RetroPlayerID `json:"ejectee_id" swaggertype:"string"`
	EjecteeName string        `json:"ejectee_name"`
	Team        *TeamID       `json:"team,omitempty" swaggertype:"string"`
	Role        string        `json:"role" enums:"P,M,C" example:"P"` // P=Player, M=Manager, C=Coach
	UmpireID    *UmpireID     `json:"umpire_id,omitempty" swaggertype:"string"`
	UmpireName  *string       `json:"umpire_name,omitempty"`
	Inning      *int          `json:"inning,omitempty"` // -1 if unknown
	Reason      *string       `json:"reason,omitempty"`
}

// Pitch represents a single pitch derived from Retrosheet play-by-play pitch sequences.
// Each pitch is extracted from the [Play.Pitches] field and annotated with context.
type Pitch struct {
	GameID  GameID `json:"game_id" swaggertype:"string"`
	PlayNum int    `json:"play_num"`
	Inning  int    `json:"inning"`
	// TODO: add this as an enum to api docs
	TopBot  int           `json:"top_bot"` // 0=top, 1=bottom
	BatTeam TeamID        `json:"bat_team" swaggertype:"string"`
	PitTeam TeamID        `json:"pit_team" swaggertype:"string"`
	Date    string        `json:"date"`
	Batter  RetroPlayerID `json:"batter" swaggertype:"string"`
	Pitcher RetroPlayerID `json:"pitcher" swaggertype:"string"`
	BatHand *string       `json:"bat_hand,omitempty"`
	PitHand *string       `json:"pit_hand,omitempty"`
	OutsPre int           `json:"outs_pre"`

	// Pitch-specific fields
	SeqNum      int     `json:"seq_num"`                // Sequence number within the plate appearance (1-indexed)
	PitchType   string  `json:"pitch_type" example:"C"` // Single character: B, C, F, S, X, etc.
	BallCount   int     `json:"ball_count"`             // Balls in count before this pitch
	StrikeCount int     `json:"strike_count"`           // Strikes in count before this pitch
	IsInPlay    bool    `json:"is_in_play"`             // True if pitch resulted in ball in play (X)
	IsStrike    bool    `json:"is_strike"`              // True if C, S, F, L, M, O, T, V
	IsBall      bool    `json:"is_ball"`                // True if B, P, I, H
	Description string  `json:"description,omitempty"`  // Human-readable pitch type description
	Event       *string `json:"event,omitempty"`        // Play result if this was the final pitch
}

// GameState represents a specific game situation used for win expectancy lookups.
// Encodes inning, outs, runners on base, and score differential.
type GameState struct {
	Inning      int    `json:"inning"`                     // Inning number (1-9, extras treated as 9)
	IsBottom    bool   `json:"is_bottom"`                  // true=bottom of inning, false=top
	Outs        int    `json:"outs"`                       // Number of outs (0-2)
	RunnersCode string `json:"runners_code" example:"1_3"` // Base state: ___ = empty, 1__ = runner on first, etc.
	ScoreDiff   int    `json:"score_diff"`                 // Score differential from batting team perspective (capped at ±11)
	Year        *int   `json:"year,omitempty"`             // Optional year for era-specific lookups
}

// WinExpectancy represents the historical win probability for a specific game state.
// This is the probability that the home team wins from the given situation.
type WinExpectancy struct {
	ID             int       `json:"id"`
	Inning         int       `json:"inning"`
	IsBottom       bool      `json:"is_bottom"`
	Outs           int       `json:"outs"`
	RunnersState   string    `json:"runners_state"`
	ScoreDiff      int       `json:"score_diff"`
	WinProbability float64   `json:"win_probability"` // 0.0 to 1.0
	SampleSize     int       `json:"sample_size"`     // Number of historical games used
	StartYear      *int      `json:"start_year,omitempty"`
	EndYear        *int      `json:"end_year,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// WinExpectancyEra represents a historical period for which win expectancy data is available.
type WinExpectancyEra struct {
	StartYear   int    `json:"start_year"`
	EndYear     int    `json:"end_year"`
	Label       string `json:"label"`        // e.g., "Modern Era", "Steroid Era"
	StateCount  int    `json:"state_count"`  // Number of unique game states in this era
	TotalSample int64  `json:"total_sample"` // Total sample size across all states
}
