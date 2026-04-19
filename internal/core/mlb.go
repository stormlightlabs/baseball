package core

import "encoding/json"

// MLBOverviewResponse represents the catalog of available MLB proxy routes.
type MLBOverviewResponse struct {
	BaseURL string                `json:"base_url"`
	Target  string                `json:"target"`
	Routes  []MLBProxyCatalogItem `json:"routes"`
}

// MLBProxyCatalogItem describes a single MLB proxy route.
type MLBProxyCatalogItem struct {
	Route       string `json:"route"`
	Target      string `json:"target"`
	Description string `json:"description"`
}

// MLBPeopleResponse represents the response from the people endpoint.
type MLBPeopleResponse struct {
	Copyright string      `json:"copyright"`
	People    []MLBPerson `json:"people"`
}

// MLBPerson represents a player in the MLB system.
type MLBPerson struct {
	ID               int            `json:"id"`
	FullName         string         `json:"fullName"`
	Link             string         `json:"link"`
	FirstName        string         `json:"firstName"`
	LastName         string         `json:"lastName"`
	PrimaryNumber    string         `json:"primaryNumber,omitempty"`
	BirthDate        string         `json:"birthDate,omitempty"`
	CurrentAge       int            `json:"currentAge,omitempty"`
	BirthCity        string         `json:"birthCity,omitempty"`
	BirthCountry     string         `json:"birthCountry,omitempty"`
	Height           string         `json:"height,omitempty"`
	Weight           int            `json:"weight,omitempty"`
	Active           bool           `json:"active"`
	PrimaryPosition  *MLBPosition   `json:"primaryPosition,omitempty"`
	UseName          string         `json:"useName,omitempty"`
	UseLastName      string         `json:"useLastName,omitempty"`
	BoxscoreName     string         `json:"boxscoreName,omitempty"`
	NickName         string         `json:"nickName,omitempty"`
	Gender           string         `json:"gender,omitempty"`
	IsPlayer         bool           `json:"isPlayer"`
	IsVerified       bool           `json:"isVerified"`
	Pronunciation    string         `json:"pronunciation,omitempty"`
	MLBDebutDate     string         `json:"mlbDebutDate,omitempty"`
	BatSide          *MLBHandedness `json:"batSide,omitempty"`
	PitchHand        *MLBHandedness `json:"pitchHand,omitempty"`
	NameFirstLast    string         `json:"nameFirstLast,omitempty"`
	NameSlug         string         `json:"nameSlug,omitempty"`
	FirstLastName    string         `json:"firstLastName,omitempty"`
	LastFirstName    string         `json:"lastFirstName,omitempty"`
	LastInitName     string         `json:"lastInitName,omitempty"`
	InitLastName     string         `json:"initLastName,omitempty"`
	FullFMLName      string         `json:"fullFMLName,omitempty"`
	FullLFMName      string         `json:"fullLFMName,omitempty"`
	StrikeZoneTop    float64        `json:"strikeZoneTop,omitempty"`
	StrikeZoneBottom float64        `json:"strikeZoneBottom,omitempty"`
	CurrentTeam      *MLBTeamRef    `json:"currentTeam,omitempty"`
}

// MLBPosition represents a player position.
type MLBPosition struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Abbreviation string `json:"abbreviation"`
}

// MLBHandedness represents batting or pitching hand preference.
type MLBHandedness struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// MLBTeamsResponse represents the response from the teams endpoint.
type MLBTeamsResponse struct {
	Copyright string    `json:"copyright"`
	Teams     []MLBTeam `json:"teams"`
}

// MLBTeam represents an MLB team.
type MLBTeam struct {
	SpringLeague    *MLBLeagueRef   `json:"springLeague,omitempty"`
	AllStarStatus   string          `json:"allStarStatus,omitempty"`
	ID              int             `json:"id"`
	Name            string          `json:"name"`
	Link            string          `json:"link"`
	Season          int             `json:"season,omitempty"`
	Venue           *MLBVenueRef    `json:"venue,omitempty"`
	SpringVenue     *MLBVenueRef    `json:"springVenue,omitempty"`
	TeamCode        string          `json:"teamCode,omitempty"`
	FileCode        string          `json:"fileCode,omitempty"`
	Abbreviation    string          `json:"abbreviation,omitempty"`
	TeamName        string          `json:"teamName,omitempty"`
	LocationName    string          `json:"locationName,omitempty"`
	FirstYearOfPlay string          `json:"firstYearOfPlay,omitempty"`
	League          *MLBLeagueRef   `json:"league,omitempty"`
	Division        *MLBDivisionRef `json:"division,omitempty"`
	Sport           *MLBSportRef    `json:"sport,omitempty"`
	ShortName       string          `json:"shortName,omitempty"`
	FranchiseName   string          `json:"franchiseName,omitempty"`
	ClubName        string          `json:"clubName,omitempty"`
	Active          bool            `json:"active"`
}

// MLBTeamCrosswalkResponse describes mappings between MLB Stats teams and local team/franchise IDs.
type MLBTeamCrosswalkResponse struct {
	RequestedSeason int                   `json:"requested_season"`
	LocalSeason     int                   `json:"local_season"`
	Matched         int                   `json:"matched"`
	Unmatched       int                   `json:"unmatched"`
	Rows            []MLBTeamCrosswalkRow `json:"rows"`
}

// MLBTeamCrosswalkRow is one MLB team to local-ID crosswalk result.
type MLBTeamCrosswalkRow struct {
	MLBTeamID        int          `json:"mlb_team_id"`
	MLBAbbreviation  string       `json:"mlb_abbreviation,omitempty"`
	MLBTeamCode      string       `json:"mlb_team_code,omitempty"`
	MLBFileCode      string       `json:"mlb_file_code,omitempty"`
	MLBTeamName      string       `json:"mlb_team_name,omitempty"`
	MLBFranchiseName string       `json:"mlb_franchise_name,omitempty"`
	MLBClubName      string       `json:"mlb_club_name,omitempty"`
	LocalTeamID      *TeamID      `json:"local_team_id,omitempty" swaggertype:"string"`
	LocalFranchiseID *FranchiseID `json:"local_franchise_id,omitempty" swaggertype:"string"`
	LocalTeamName    string       `json:"local_team_name,omitempty"`
	LocalLeague      *LeagueID    `json:"local_league,omitempty" swaggertype:"string"`
	MatchMethod      string       `json:"match_method,omitempty"`
	Confidence       string       `json:"confidence,omitempty"`
}

// MLBLeagueRef represents a league reference.
type MLBLeagueRef struct {
	ID           int    `json:"id"`
	Name         string `json:"name,omitempty"`
	Link         string `json:"link"`
	Abbreviation string `json:"abbreviation,omitempty"`
}

// MLBDivisionRef represents a division reference.
type MLBDivisionRef struct {
	ID           int    `json:"id"`
	Name         string `json:"name,omitempty"`
	Link         string `json:"link"`
	Abbreviation string `json:"abbreviation,omitempty"`
}

// MLBSportRef represents a sport reference.
type MLBSportRef struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
	Link string `json:"link"`
}

// MLBVenueRef represents a venue reference.
type MLBVenueRef struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
	Link string `json:"link"`
}

// MLBNamedResource is a generic typed resource shape returned by MLB Stats API.
type MLBNamedResource struct {
	ID           int    `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Link         string `json:"link,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Abbreviation string `json:"abbreviation,omitempty"`
}

// MLBScheduleResponse represents the response from the schedule endpoint.
type MLBScheduleResponse struct {
	Copyright            string            `json:"copyright"`
	TotalItems           int               `json:"totalItems"`
	TotalEvents          int               `json:"totalEvents"`
	TotalGames           int               `json:"totalGames"`
	TotalGamesInProgress int               `json:"totalGamesInProgress"`
	Dates                []MLBScheduleDate `json:"dates"`
}

// MLBScheduleDate represents a date in the schedule with games.
type MLBScheduleDate struct {
	Date                 string            `json:"date"`
	TotalItems           int               `json:"totalItems"`
	TotalEvents          int               `json:"totalEvents"`
	TotalGames           int               `json:"totalGames"`
	TotalGamesInProgress int               `json:"totalGamesInProgress"`
	Games                []MLBGame         `json:"games"`
	Events               []json.RawMessage `json:"events,omitempty"`
}

// MLBGame represents a game in the schedule payload.
type MLBGame struct {
	GamePk            int               `json:"gamePk"`
	Link              string            `json:"link"`
	GameType          string            `json:"gameType,omitempty"`
	Season            string            `json:"season,omitempty"`
	GameDate          string            `json:"gameDate,omitempty"`
	OfficialDate      string            `json:"officialDate,omitempty"`
	Status            *MLBGameStatus    `json:"status,omitempty"`
	Teams             *MLBGameTeams     `json:"teams,omitempty"`
	Venue             *MLBVenueRef      `json:"venue,omitempty"`
	Content           *MLBNamedResource `json:"content,omitempty"`
	GameNumber        int               `json:"gameNumber,omitempty"`
	DoubleHeader      string            `json:"doubleHeader,omitempty"`
	GamedayType       string            `json:"gamedayType,omitempty"`
	SeriesDescription string            `json:"seriesDescription,omitempty"`
	ScheduledInnings  int               `json:"scheduledInnings,omitempty"`
	GamesInSeries     int               `json:"gamesInSeries,omitempty"`
	SeriesGameNumber  int               `json:"seriesGameNumber,omitempty"`
	DayNight          string            `json:"dayNight,omitempty"`
	Linescore         *MLBLineScore     `json:"linescore,omitempty"`
}

// MLBGameStatus represents game state and labels.
type MLBGameStatus struct {
	AbstractGameState string `json:"abstractGameState,omitempty"`
	CodedGameState    string `json:"codedGameState,omitempty"`
	DetailedState     string `json:"detailedState,omitempty"`
	StatusCode        string `json:"statusCode,omitempty"`
	StartTimeTBD      bool   `json:"startTimeTBD,omitempty"`
	AbstractGameCode  string `json:"abstractGameCode,omitempty"`
}

// MLBGameTeams describes away/home team sections for a game.
type MLBGameTeams struct {
	Away *MLBGameTeamWrapper `json:"away,omitempty"`
	Home *MLBGameTeamWrapper `json:"home,omitempty"`
}

// MLBGameTeamWrapper wraps score state and team details.
type MLBGameTeamWrapper struct {
	LeagueRecord *MLBWinLossRecord `json:"leagueRecord,omitempty"`
	Score        int               `json:"score,omitempty"`
	Team         *MLBTeamRef       `json:"team,omitempty"`
	IsWinner     bool              `json:"isWinner,omitempty"`
	SplitSquad   bool              `json:"splitSquad,omitempty"`
	SeriesNumber int               `json:"seriesNumber,omitempty"`
}

// MLBLineScore represents a live linescore section.
type MLBLineScore struct {
	CurrentInning        int                  `json:"currentInning,omitempty"`
	CurrentInningOrdinal string               `json:"currentInningOrdinal,omitempty"`
	InningState          string               `json:"inningState,omitempty"`
	InningHalf           string               `json:"inningHalf,omitempty"`
	IsTopInning          bool                 `json:"isTopInning,omitempty"`
	ScheduledInnings     int                  `json:"scheduledInnings,omitempty"`
	Innings              []MLBLineScoreInning `json:"innings,omitempty"`
	Teams                *MLBLineScoreTeams   `json:"teams,omitempty"`
	Offense              *MLBLineScoreOffense `json:"offense,omitempty"`
	Defense              *MLBLineScoreDefense `json:"defense,omitempty"`
	Balls                int                  `json:"balls,omitempty"`
	Strikes              int                  `json:"strikes,omitempty"`
	Outs                 int                  `json:"outs,omitempty"`
}

// MLBLineScoreInning is one inning slice from linescore.
type MLBLineScoreInning struct {
	Num        int                     `json:"num,omitempty"`
	OrdinalNum string                  `json:"ordinalNum,omitempty"`
	Home       *MLBLineScoreInningHalf `json:"home,omitempty"`
	Away       *MLBLineScoreInningHalf `json:"away,omitempty"`
}

// MLBLineScoreInningHalf is home/away inning line details.
type MLBLineScoreInningHalf struct {
	Runs   int `json:"runs,omitempty"`
	Hits   int `json:"hits,omitempty"`
	Errors int `json:"errors,omitempty"`
}

// MLBLineScoreTeams aggregates total linescore per side.
type MLBLineScoreTeams struct {
	Home *MLBLineScoreInningHalf `json:"home,omitempty"`
	Away *MLBLineScoreInningHalf `json:"away,omitempty"`
}

// MLBLineScoreOffense represents current offense base runners.
type MLBLineScoreOffense struct {
	Batter  *MLBPlayerRef `json:"batter,omitempty"`
	OnDeck  *MLBPlayerRef `json:"onDeck,omitempty"`
	InHole  *MLBPlayerRef `json:"inHole,omitempty"`
	Pitcher *MLBPlayerRef `json:"pitcher,omitempty"`
	First   *MLBPlayerRef `json:"first,omitempty"`
	Second  *MLBPlayerRef `json:"second,omitempty"`
	Third   *MLBPlayerRef `json:"third,omitempty"`
	Team    *MLBTeamRef   `json:"team,omitempty"`
}

// MLBLineScoreDefense represents defensive alignment anchors.
type MLBLineScoreDefense struct {
	Pitcher   *MLBPlayerRef `json:"pitcher,omitempty"`
	Catcher   *MLBPlayerRef `json:"catcher,omitempty"`
	First     *MLBPlayerRef `json:"first,omitempty"`
	Second    *MLBPlayerRef `json:"second,omitempty"`
	Third     *MLBPlayerRef `json:"third,omitempty"`
	Shortstop *MLBPlayerRef `json:"shortstop,omitempty"`
	Left      *MLBPlayerRef `json:"left,omitempty"`
	Center    *MLBPlayerRef `json:"center,omitempty"`
	Right     *MLBPlayerRef `json:"right,omitempty"`
	Team      *MLBTeamRef   `json:"team,omitempty"`
}

// MLBPlayerRef is a lightweight player link payload.
type MLBPlayerRef struct {
	ID       int    `json:"id"`
	FullName string `json:"fullName,omitempty"`
	Link     string `json:"link,omitempty"`
}

// MLBSeasonsResponse represents the response from the seasons endpoint.
type MLBSeasonsResponse struct {
	Copyright string      `json:"copyright"`
	Seasons   []MLBSeason `json:"seasons"`
}

// MLBSeason represents a season in MLB.
type MLBSeason struct {
	SeasonID                  string  `json:"seasonId"`
	HasWildcard               bool    `json:"hasWildcard"`
	PreSeasonStartDate        string  `json:"preSeasonStartDate,omitempty"`
	PreSeasonEndDate          string  `json:"preSeasonEndDate,omitempty"`
	SeasonStartDate           string  `json:"seasonStartDate,omitempty"`
	SpringStartDate           string  `json:"springStartDate,omitempty"`
	SpringEndDate             string  `json:"springEndDate,omitempty"`
	RegularSeasonStartDate    string  `json:"regularSeasonStartDate,omitempty"`
	LastDate1stHalf           string  `json:"lastDate1stHalf,omitempty"`
	AllStarDate               string  `json:"allStarDate,omitempty"`
	FirstDate2ndHalf          string  `json:"firstDate2ndHalf,omitempty"`
	RegularSeasonEndDate      string  `json:"regularSeasonEndDate,omitempty"`
	PostSeasonStartDate       string  `json:"postSeasonStartDate,omitempty"`
	PostSeasonEndDate         string  `json:"postSeasonEndDate,omitempty"`
	SeasonEndDate             string  `json:"seasonEndDate,omitempty"`
	OffseasonStartDate        string  `json:"offseasonStartDate,omitempty"`
	OffSeasonEndDate          string  `json:"offSeasonEndDate,omitempty"`
	SeasonLevelGamedayType    string  `json:"seasonLevelGamedayType,omitempty"`
	GameLevelGamedayType      string  `json:"gameLevelGamedayType,omitempty"`
	QualifierPlateAppearances float64 `json:"qualifierPlateAppearances,omitempty"`
	QualifierOutsPitched      float64 `json:"qualifierOutsPitched,omitempty"`
}

// MLBStatsResponse represents the top-level /v1/stats payload.
type MLBStatsResponse struct {
	Copyright string          `json:"copyright"`
	Stats     []MLBStatResult `json:"stats"`
}

// MLBStatResult is one stat result group in /v1/stats.
type MLBStatResult struct {
	Type        *MLBNamedResource  `json:"type,omitempty"`
	Group       *MLBNamedResource  `json:"group,omitempty"`
	DisplayName string             `json:"displayName,omitempty"`
	SortBy      string             `json:"sortBy,omitempty"`
	Season      string             `json:"season,omitempty"`
	TotalSplits int                `json:"totalSplits,omitempty"`
	Splits      []MLBStatSplit     `json:"splits,omitempty"`
	Exemptions  []MLBStatExemption `json:"exemptions,omitempty"`
}

// MLBCommonStatGroup captures keys shared across multiple stats groups.
type MLBCommonStatGroup struct {
	GamesPlayed         json.RawMessage `json:"gamesPlayed,omitempty"`
	GamesStarted        json.RawMessage `json:"gamesStarted,omitempty"`
	PlateAppearances    json.RawMessage `json:"plateAppearances,omitempty"`
	AtBats              json.RawMessage `json:"atBats,omitempty"`
	Runs                json.RawMessage `json:"runs,omitempty"`
	Hits                json.RawMessage `json:"hits,omitempty"`
	Doubles             json.RawMessage `json:"doubles,omitempty"`
	Triples             json.RawMessage `json:"triples,omitempty"`
	HomeRuns            json.RawMessage `json:"homeRuns,omitempty"`
	RBI                 json.RawMessage `json:"rbi,omitempty"`
	StolenBases         json.RawMessage `json:"stolenBases,omitempty"`
	CaughtStealing      json.RawMessage `json:"caughtStealing,omitempty"`
	BaseOnBalls         json.RawMessage `json:"baseOnBalls,omitempty"`
	IntentionalWalks    json.RawMessage `json:"intentionalWalks,omitempty"`
	StrikeOuts          json.RawMessage `json:"strikeOuts,omitempty"`
	HitByPitch          json.RawMessage `json:"hitByPitch,omitempty"`
	SacBunts            json.RawMessage `json:"sacBunts,omitempty"`
	SacFlies            json.RawMessage `json:"sacFlies,omitempty"`
	Avg                 json.RawMessage `json:"avg,omitempty"`
	OBP                 json.RawMessage `json:"obp,omitempty"`
	SLG                 json.RawMessage `json:"slg,omitempty"`
	OPS                 json.RawMessage `json:"ops,omitempty"`
	TotalBases          json.RawMessage `json:"totalBases,omitempty"`
	BABIP               json.RawMessage `json:"babip,omitempty"`
	GroundOuts          json.RawMessage `json:"groundOuts,omitempty"`
	AirOuts             json.RawMessage `json:"airOuts,omitempty"`
	GroundOutsToAirouts json.RawMessage `json:"groundOutsToAirouts,omitempty"`
	NumberOfPitches     json.RawMessage `json:"numberOfPitches,omitempty"`
	StrikePercentage    json.RawMessage `json:"strikePercentage,omitempty"`
	InningsPitched      json.RawMessage `json:"inningsPitched,omitempty"`
	Wins                json.RawMessage `json:"wins,omitempty"`
	Losses              json.RawMessage `json:"losses,omitempty"`
	EarnedRuns          json.RawMessage `json:"earnedRuns,omitempty"`
	ERA                 json.RawMessage `json:"era,omitempty"`
	WHIP                json.RawMessage `json:"whip,omitempty"`
	Saves               json.RawMessage `json:"saves,omitempty"`
	SaveOpportunities   json.RawMessage `json:"saveOpportunities,omitempty"`
	BlownSaves          json.RawMessage `json:"blownSaves,omitempty"`
	Holds               json.RawMessage `json:"holds,omitempty"`
	HitBatsmen          json.RawMessage `json:"hitBatsmen,omitempty"`
	WildPitches         json.RawMessage `json:"wildPitches,omitempty"`
	Balks               json.RawMessage `json:"balks,omitempty"`
	QualityStarts       json.RawMessage `json:"qualityStarts,omitempty"`
}

// MLBHittingStatGroup captures hitting-specific keys from /v1/stats.
type MLBHittingStatGroup struct {
	Singles             json.RawMessage `json:"singles,omitempty"`
	ExtraBaseHits       json.RawMessage `json:"extraBaseHits,omitempty"`
	WalkOffs            json.RawMessage `json:"walkOffs,omitempty"`
	AtBatsPerHomeRun    json.RawMessage `json:"atBatsPerHomeRun,omitempty"`
	LeftOnBase          json.RawMessage `json:"leftOnBase,omitempty"`
	GIDP                json.RawMessage `json:"gidp,omitempty"`
	AtBatsPerStrikeout  json.RawMessage `json:"atBatsPerStrikeout,omitempty"`
	PitchesPerPlateAppr json.RawMessage `json:"pitchesPerPlateAppearance,omitempty"`
}

// MLBPitchingStatGroup captures pitching-specific keys from /v1/stats.
type MLBPitchingStatGroup struct {
	GamesFinished                json.RawMessage `json:"gamesFinished,omitempty"`
	Shutouts                     json.RawMessage `json:"shutouts,omitempty"`
	CompleteGames                json.RawMessage `json:"completeGames,omitempty"`
	StrikeoutsPer9Inn            json.RawMessage `json:"strikeoutsPer9Inn,omitempty"`
	WalksPer9Inn                 json.RawMessage `json:"walksPer9Inn,omitempty"`
	HitsPer9Inn                  json.RawMessage `json:"hitsPer9Inn,omitempty"`
	HomeRunsPer9                 json.RawMessage `json:"homeRunsPer9,omitempty"`
	SavePercentage               json.RawMessage `json:"savePercentage,omitempty"`
	InheritedRunners             json.RawMessage `json:"inheritedRunners,omitempty"`
	InheritedRunnersScored       json.RawMessage `json:"inheritedRunnersScored,omitempty"`
	WinPercentage                json.RawMessage `json:"winPercentage,omitempty"`
	PitchesPerInning             json.RawMessage `json:"pitchesPerInning,omitempty"`
	WalksAndHitsPerInningPitched json.RawMessage `json:"walksAndHitsPerInningPitched,omitempty"`
}

// MLBFieldingStatGroup captures fielding-specific keys from /v1/stats.
type MLBFieldingStatGroup struct {
	Assists            json.RawMessage `json:"assists,omitempty"`
	PutOuts            json.RawMessage `json:"putOuts,omitempty"`
	Errors             json.RawMessage `json:"errors,omitempty"`
	Chances            json.RawMessage `json:"chances,omitempty"`
	Fielding           json.RawMessage `json:"fielding,omitempty"`
	PassedBall         json.RawMessage `json:"passedBall,omitempty"`
	StolenBasesAllowed json.RawMessage `json:"stolenBasesAllowed,omitempty"`
	CaughtStealingPct  json.RawMessage `json:"caughtStealingPct,omitempty"`
	RangeFactorPerGame json.RawMessage `json:"rangeFactorPerGame,omitempty"`
	RangeFactorPer9Inn json.RawMessage `json:"rangeFactorPer9Inn,omitempty"`
	DoublePlays        json.RawMessage `json:"doublePlays,omitempty"`
	TriplePlays        json.RawMessage `json:"triplePlays,omitempty"`
	ThrowingErrors     json.RawMessage `json:"throwingErrors,omitempty"`
}

// MLBStatGroup is a typed union of MLB stat group fields while preserving unknown keys.
type MLBStatGroup struct {
	MLBCommonStatGroup
	MLBHittingStatGroup
	MLBPitchingStatGroup
	MLBFieldingStatGroup
	raw map[string]json.RawMessage
}

type mlbStatGroupKnown struct {
	MLBCommonStatGroup
	MLBHittingStatGroup
	MLBPitchingStatGroup
	MLBFieldingStatGroup
}

// UnmarshalJSON stores the full raw stat object and decodes known stat-group keys.
func (g *MLBStatGroup) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*g = MLBStatGroup{}
		return nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var known mlbStatGroupKnown
	if err := json.Unmarshal(data, &known); err != nil {
		return err
	}

	g.MLBCommonStatGroup = known.MLBCommonStatGroup
	g.MLBHittingStatGroup = known.MLBHittingStatGroup
	g.MLBPitchingStatGroup = known.MLBPitchingStatGroup
	g.MLBFieldingStatGroup = known.MLBFieldingStatGroup
	g.raw = raw

	return nil
}

// MarshalJSON emits the original stat object to preserve upstream fidelity.
func (g MLBStatGroup) MarshalJSON() ([]byte, error) {
	if len(g.raw) > 0 {
		return json.Marshal(g.raw)
	}

	known := mlbStatGroupKnown{
		MLBCommonStatGroup:   g.MLBCommonStatGroup,
		MLBHittingStatGroup:  g.MLBHittingStatGroup,
		MLBPitchingStatGroup: g.MLBPitchingStatGroup,
		MLBFieldingStatGroup: g.MLBFieldingStatGroup,
	}
	return json.Marshal(known)
}

// MLBStatSplit is a single split row in a stat result.
type MLBStatSplit struct {
	Season   string          `json:"season,omitempty"`
	Stat     MLBStatGroup    `json:"stat"`
	Team     *MLBTeamRef     `json:"team,omitempty"`
	Player   *MLBPlayerRef   `json:"player,omitempty"`
	League   *MLBLeagueRef   `json:"league,omitempty"`
	Division *MLBDivisionRef `json:"division,omitempty"`
	Sport    *MLBSportRef    `json:"sport,omitempty"`
	Position *MLBPosition    `json:"position,omitempty"`
	GameType string          `json:"gameType,omitempty"`
	Rank     int             `json:"rank,omitempty"`
}

// MLBStatExemption represents one exemption row in stat payloads.
type MLBStatExemption struct {
	Type   string `json:"type,omitempty"`
	Code   string `json:"code,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// MLBStandingsResponse represents the response from the standings endpoint.
type MLBStandingsResponse struct {
	Copyright string               `json:"copyright"`
	Records   []MLBStandingsRecord `json:"records"`
}

// MLBStandingsRecord represents standings for a division or league.
type MLBStandingsRecord struct {
	StandingsType string          `json:"standingsType,omitempty"`
	League        *MLBLeagueRef   `json:"league,omitempty"`
	Division      *MLBDivisionRef `json:"division,omitempty"`
	Sport         *MLBSportRef    `json:"sport,omitempty"`
	RoundRobin    *MLBRoundRobin  `json:"roundRobin,omitempty"`
	LastUpdated   string          `json:"lastUpdated,omitempty"`
	Season        string          `json:"season,omitempty"`
	TeamRecords   []MLBTeamRecord `json:"teamRecords"`
}

// MLBRoundRobin represents round robin status.
type MLBRoundRobin struct {
	Status string `json:"status,omitempty"`
}

// MLBTeamRecord represents a team's record in standings.
type MLBTeamRecord struct {
	Team                      *MLBTeamRef       `json:"team"`
	Season                    string            `json:"season,omitempty"`
	Streak                    *MLBStreak        `json:"streak,omitempty"`
	ClinchIndicator           string            `json:"clinchIndicator,omitempty"`
	DivisionRank              string            `json:"divisionRank,omitempty"`
	LeagueRank                string            `json:"leagueRank,omitempty"`
	SportRank                 string            `json:"sportRank,omitempty"`
	WildCardRank              string            `json:"wildCardRank,omitempty"`
	LeagueGamesBack           string            `json:"leagueGamesBack,omitempty"`
	SportGamesBack            string            `json:"sportGamesBack,omitempty"`
	DivisionGamesBack         string            `json:"divisionGamesBack,omitempty"`
	ConferenceGamesBack       string            `json:"conferenceGamesBack,omitempty"`
	GamesPlayed               int               `json:"gamesPlayed,omitempty"`
	Wins                      int               `json:"wins,omitempty"`
	Losses                    int               `json:"losses,omitempty"`
	WinningPercentage         string            `json:"winningPercentage,omitempty"`
	GamesBack                 string            `json:"gamesBack,omitempty"`
	WildCardGamesBack         string            `json:"wildCardGamesBack,omitempty"`
	LeagueRecord              *MLBWinLossRecord `json:"leagueRecord,omitempty"`
	Records                   *MLBRecordDetails `json:"records,omitempty"`
	RunsAllowed               int               `json:"runsAllowed,omitempty"`
	RunsScored                int               `json:"runsScored,omitempty"`
	RunDifferential           string            `json:"runDifferential,omitempty"`
	DivisionChamp             bool              `json:"divisionChamp,omitempty"`
	DivisionLeader            bool              `json:"divisionLeader,omitempty"`
	HasWildcard               bool              `json:"hasWildcard,omitempty"`
	Clinched                  bool              `json:"clinched,omitempty"`
	MagicNumber               string            `json:"magicNumber,omitempty"`
	EliminationNumber         string            `json:"eliminationNumber,omitempty"`
	WildCardEliminationNumber string            `json:"wildCardEliminationNumber,omitempty"`
	LastUpdated               string            `json:"lastUpdated,omitempty"`
	LastTenRecords            []MLBSplitRecord  `json:"lastTenRecords,omitempty"`
}

// MLBTeamRef represents a team reference in standings and related payloads.
type MLBTeamRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
}

// MLBStreak represents a team's win/loss streak.
type MLBStreak struct {
	StreakCode   string `json:"streakCode,omitempty"`
	StreakType   string `json:"streakType,omitempty"`
	StreakNumber int    `json:"streakNumber,omitempty"`
}

// MLBWinLossRecord represents a win-loss record.
type MLBWinLossRecord struct {
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
	Ties   int    `json:"ties,omitempty"`
	Pct    string `json:"pct,omitempty"`
}

// MLBRecordDetails contains detailed split records.
type MLBRecordDetails struct {
	SplitRecords    []MLBSplitRecord    `json:"splitRecords,omitempty"`
	DivisionRecords []MLBDivisionRecord `json:"divisionRecords,omitempty"`
	LeagueRecords   []MLBLeagueRecord   `json:"leagueRecords,omitempty"`
	ExpectedRecords []MLBSplitRecord    `json:"expectedRecords,omitempty"`
}

// MLBSplitRecord represents a record split (home/away, day/night, etc.).
type MLBSplitRecord struct {
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
	Type   string `json:"type,omitempty"`
	Pct    string `json:"pct,omitempty"`
}

// MLBDivisionRecord represents a record against a specific division.
type MLBDivisionRecord struct {
	Wins     int             `json:"wins"`
	Losses   int             `json:"losses"`
	Pct      string          `json:"pct,omitempty"`
	Division *MLBDivisionRef `json:"division,omitempty"`
}

// MLBLeagueRecord represents a record against a specific league.
type MLBLeagueRecord struct {
	Wins   int           `json:"wins"`
	Losses int           `json:"losses"`
	Pct    string        `json:"pct,omitempty"`
	League *MLBLeagueRef `json:"league,omitempty"`
}

// MLBAwardsResponse represents the response from the awards endpoint.
type MLBAwardsResponse struct {
	Copyright string     `json:"copyright"`
	Awards    []MLBAward `json:"awards"`
}

// MLBAward represents an MLB award.
type MLBAward struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Link            string              `json:"link,omitempty"`
	Description     string              `json:"description,omitempty"`
	SortOrder       int                 `json:"sortOrder,omitempty"`
	Sport           *MLBSportRef        `json:"sport,omitempty"`
	AwardRecipients []MLBAwardRecipient `json:"awardRecipients,omitempty"`
}

// MLBAwardRecipient represents one recipient entry under an award payload.
type MLBAwardRecipient struct {
	ID     int         `json:"id,omitempty"`
	Name   string      `json:"name,omitempty"`
	Link   string      `json:"link,omitempty"`
	Notes  string      `json:"notes,omitempty"`
	Season string      `json:"season,omitempty"`
	Date   string      `json:"date,omitempty"`
	Player *MLBPerson  `json:"player,omitempty"`
	Team   *MLBTeamRef `json:"team,omitempty"`
}

// MLBVenuesResponse represents the response from the venues endpoint.
type MLBVenuesResponse struct {
	Copyright string     `json:"copyright"`
	Venues    []MLBVenue `json:"venues"`
}

// MLBVenue represents a ballpark venue.
type MLBVenue struct {
	ID        int               `json:"id"`
	Link      string            `json:"link"`
	Name      string            `json:"name"`
	Active    bool              `json:"active"`
	Location  *MLBVenueLocation `json:"location,omitempty"`
	TimeZone  *MLBTimeZone      `json:"timeZone,omitempty"`
	FieldInfo *MLBFieldInfo     `json:"fieldInfo,omitempty"`
	Season    string            `json:"season,omitempty"`
	VenueType string            `json:"venueType,omitempty"`
	RoofType  string            `json:"roofType,omitempty"`
	TurfType  string            `json:"turfType,omitempty"`
	Capacity  int               `json:"capacity,omitempty"`
}

// MLBVenueLocation represents venue geography metadata.
type MLBVenueLocation struct {
	Address1           string               `json:"address1,omitempty"`
	Address2           string               `json:"address2,omitempty"`
	City               string               `json:"city,omitempty"`
	State              string               `json:"state,omitempty"`
	StateAbbrev        string               `json:"stateAbbrev,omitempty"`
	PostalCode         string               `json:"postalCode,omitempty"`
	Country            string               `json:"country,omitempty"`
	Phone              string               `json:"phone,omitempty"`
	DefaultCoordinates *MLBVenueCoordinates `json:"defaultCoordinates,omitempty"`
}

// MLBVenueCoordinates represents latitude/longitude coordinate pairs for venue location.
type MLBVenueCoordinates struct {
	Latitude  string `json:"latitude,omitempty"`
	Longitude string `json:"longitude,omitempty"`
}

// MLBTimeZone represents venue time zone metadata.
type MLBTimeZone struct {
	ID     string `json:"id,omitempty"`
	Offset int    `json:"offset,omitempty"`
	TZ     string `json:"tz,omitempty"`
}

// MLBFieldInfo represents venue dimensions and surface metadata.
type MLBFieldInfo struct {
	Capacity    int    `json:"capacity,omitempty"`
	TurfType    string `json:"turfType,omitempty"`
	RoofType    string `json:"roofType,omitempty"`
	LeftLine    int    `json:"leftLine,omitempty"`
	Left        int    `json:"left,omitempty"`
	LeftCenter  int    `json:"leftCenter,omitempty"`
	Center      int    `json:"center,omitempty"`
	RightCenter int    `json:"rightCenter,omitempty"`
	Right       int    `json:"right,omitempty"`
	RightLine   int    `json:"rightLine,omitempty"`
}
