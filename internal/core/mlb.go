package core

// MLBOverviewResponse represents the catalog of available MLB proxy routes
type MLBOverviewResponse struct {
	BaseURL string                `json:"base_url"`
	Target  string                `json:"target"`
	Routes  []MLBProxyCatalogItem `json:"routes"`
}

// MLBProxyCatalogItem describes a single MLB proxy route
type MLBProxyCatalogItem struct {
	Route       string `json:"route"`
	Target      string `json:"target"`
	Description string `json:"description"`
}

// MLBPeopleResponse represents the response from the people endpoint
type MLBPeopleResponse struct {
	Copyright string      `json:"copyright"`
	People    []MLBPerson `json:"people"`
}

// MLBPerson represents a player in the MLB system
type MLBPerson struct {
	ID               int            `json:"id"`
	FullName         string         `json:"fullName"`
	Link             string         `json:"link"`
	FirstName        string         `json:"firstName"`
	LastName         string         `json:"lastName"`
	PrimaryNumber    string         `json:"primaryNumber,omitempty"`
	BirthDate        string         `json:"birthDate"`
	CurrentAge       int            `json:"currentAge"`
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
	NameFirstLast    string         `json:"nameFirstLast"`
	NameSlug         string         `json:"nameSlug"`
	FirstLastName    string         `json:"firstLastName"`
	LastFirstName    string         `json:"lastFirstName"`
	LastInitName     string         `json:"lastInitName"`
	InitLastName     string         `json:"initLastName"`
	FullFMLName      string         `json:"fullFMLName"`
	FullLFMName      string         `json:"fullLFMName"`
	StrikeZoneTop    float64        `json:"strikeZoneTop,omitempty"`
	StrikeZoneBottom float64        `json:"strikeZoneBottom,omitempty"`
}

// MLBPosition represents a player position
type MLBPosition struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Abbreviation string `json:"abbreviation"`
}

// MLBHandedness represents batting or pitching hand preference
type MLBHandedness struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// MLBTeamsResponse represents the response from the teams endpoint
type MLBTeamsResponse struct {
	Copyright string    `json:"copyright"`
	Teams     []MLBTeam `json:"teams"`
}

// MLBTeam represents an MLB team
type MLBTeam struct {
	SpringLeague    *MLBLeagueRef   `json:"springLeague,omitempty"`
	AllStarStatus   string          `json:"allStarStatus"`
	ID              int             `json:"id"`
	Name            string          `json:"name"`
	Link            string          `json:"link"`
	Season          int             `json:"season"`
	Venue           *MLBVenueRef    `json:"venue,omitempty"`
	SpringVenue     *MLBVenueRef    `json:"springVenue,omitempty"`
	TeamCode        string          `json:"teamCode"`
	FileCode        string          `json:"fileCode"`
	Abbreviation    string          `json:"abbreviation"`
	TeamName        string          `json:"teamName"`
	LocationName    string          `json:"locationName"`
	FirstYearOfPlay string          `json:"firstYearOfPlay"`
	League          *MLBLeagueRef   `json:"league,omitempty"`
	Division        *MLBDivisionRef `json:"division,omitempty"`
	Sport           *MLBSportRef    `json:"sport,omitempty"`
	ShortName       string          `json:"shortName"`
	FranchiseName   string          `json:"franchiseName,omitempty"`
	ClubName        string          `json:"clubName"`
	Active          bool            `json:"active"`
}

// MLBLeagueRef represents a league reference
type MLBLeagueRef struct {
	ID           int    `json:"id"`
	Name         string `json:"name,omitempty"`
	Link         string `json:"link"`
	Abbreviation string `json:"abbreviation,omitempty"`
}

// MLBDivisionRef represents a division reference
type MLBDivisionRef struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
	Link string `json:"link"`
}

// MLBSportRef represents a sport reference
type MLBSportRef struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
	Link string `json:"link"`
}

// MLBVenueRef represents a venue reference
type MLBVenueRef struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
	Link string `json:"link"`
}

// MLBScheduleResponse represents the response from the schedule endpoint
type MLBScheduleResponse struct {
	Copyright            string            `json:"copyright"`
	TotalItems           int               `json:"totalItems"`
	TotalEvents          int               `json:"totalEvents"`
	TotalGames           int               `json:"totalGames"`
	TotalGamesInProgress int               `json:"totalGamesInProgress"`
	Dates                []MLBScheduleDate `json:"dates"`
}

// MLBScheduleDate represents a date in the schedule with games
type MLBScheduleDate struct {
	Date                 string        `json:"date"`
	TotalItems           int           `json:"totalItems"`
	TotalEvents          int           `json:"totalEvents"`
	TotalGames           int           `json:"totalGames"`
	TotalGamesInProgress int           `json:"totalGamesInProgress"`
	Games                []MLBGame     `json:"games"`
	Events               []interface{} `json:"events"`
}

// MLBGame represents a game in the schedule
type MLBGame struct {
	GamePk   int    `json:"gamePk"`
	Link     string `json:"link"`
	GameType string `json:"gameType"`
	Season   string `json:"season"`
	GameDate string `json:"gameDate"`
	// Additional fields would be added based on actual game data structure
}

// MLBSeasonsResponse represents the response from the seasons endpoint
type MLBSeasonsResponse struct {
	Copyright string      `json:"copyright"`
	Seasons   []MLBSeason `json:"seasons"`
}

// MLBSeason represents a season in MLB
type MLBSeason struct {
	SeasonID                  string  `json:"seasonId"`
	HasWildcard               bool    `json:"hasWildcard"`
	PreSeasonStartDate        string  `json:"preSeasonStartDate"`
	PreSeasonEndDate          string  `json:"preSeasonEndDate"`
	SeasonStartDate           string  `json:"seasonStartDate"`
	SpringStartDate           string  `json:"springStartDate"`
	SpringEndDate             string  `json:"springEndDate"`
	RegularSeasonStartDate    string  `json:"regularSeasonStartDate"`
	LastDate1stHalf           string  `json:"lastDate1stHalf"`
	AllStarDate               string  `json:"allStarDate"`
	FirstDate2ndHalf          string  `json:"firstDate2ndHalf"`
	RegularSeasonEndDate      string  `json:"regularSeasonEndDate"`
	PostSeasonStartDate       string  `json:"postSeasonStartDate"`
	PostSeasonEndDate         string  `json:"postSeasonEndDate"`
	SeasonEndDate             string  `json:"seasonEndDate"`
	OffseasonStartDate        string  `json:"offseasonStartDate"`
	OffSeasonEndDate          string  `json:"offSeasonEndDate"`
	SeasonLevelGamedayType    string  `json:"seasonLevelGamedayType"`
	GameLevelGamedayType      string  `json:"gameLevelGamedayType"`
	QualifierPlateAppearances float64 `json:"qualifierPlateAppearances"`
	QualifierOutsPitched      float64 `json:"qualifierOutsPitched"`
}

// MLBStandingsResponse represents the response from the standings endpoint
type MLBStandingsResponse struct {
	Copyright string               `json:"copyright"`
	Records   []MLBStandingsRecord `json:"records"`
}

// MLBStandingsRecord represents standings for a division or league
type MLBStandingsRecord struct {
	StandingsType string          `json:"standingsType"`
	League        *MLBLeagueRef   `json:"league,omitempty"`
	Division      *MLBDivisionRef `json:"division,omitempty"`
	Sport         *MLBSportRef    `json:"sport,omitempty"`
	RoundRobin    *MLBRoundRobin  `json:"roundRobin,omitempty"`
	LastUpdated   string          `json:"lastUpdated"`
	TeamRecords   []MLBTeamRecord `json:"teamRecords"`
}

// MLBRoundRobin represents round robin status
type MLBRoundRobin struct {
	Status string `json:"status"`
}

// MLBTeamRecord represents a team's record in the standings
type MLBTeamRecord struct {
	Team              *MLBTeamRef       `json:"team"`
	Season            string            `json:"season"`
	Streak            *MLBStreak        `json:"streak,omitempty"`
	ClinchIndicator   string            `json:"clinchIndicator,omitempty"`
	DivisionRank      string            `json:"divisionRank"`
	LeagueRank        string            `json:"leagueRank"`
	SportRank         string            `json:"sportRank"`
	GamesPlayed       int               `json:"gamesPlayed"`
	GamesBack         string            `json:"gamesBack"`
	WildCardGamesBack string            `json:"wildCardGamesBack"`
	LeagueRecord      *MLBWinLossRecord `json:"leagueRecord"`
	Records           *MLBRecordDetails `json:"records,omitempty"`
	LastUpdated       string            `json:"lastUpdated"`
}

// MLBTeamRef represents a team reference in standings
type MLBTeamRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
}

// MLBStreak represents a team's win/loss streak
type MLBStreak struct {
	StreakCode   string `json:"streakCode"`
	StreakType   string `json:"streakType"`
	StreakNumber int    `json:"streakNumber"`
}

// MLBWinLossRecord represents a win-loss record
type MLBWinLossRecord struct {
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
	Ties   int    `json:"ties,omitempty"`
	Pct    string `json:"pct"`
}

// MLBRecordDetails contains detailed split records
type MLBRecordDetails struct {
	SplitRecords    []MLBSplitRecord    `json:"splitRecords,omitempty"`
	DivisionRecords []MLBDivisionRecord `json:"divisionRecords,omitempty"`
}

// MLBSplitRecord represents a record split (home/away, day/night, etc.)
type MLBSplitRecord struct {
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
	Type   string `json:"type"`
	Pct    string `json:"pct"`
}

// MLBDivisionRecord represents a record against a specific division
type MLBDivisionRecord struct {
	Wins     int             `json:"wins"`
	Losses   int             `json:"losses"`
	Pct      string          `json:"pct"`
	Division *MLBDivisionRef `json:"division"`
}

// MLBAwardsResponse represents the response from the awards endpoint
type MLBAwardsResponse struct {
	Copyright string     `json:"copyright"`
	Awards    []MLBAward `json:"awards"`
}

// MLBAward represents an MLB award
type MLBAward struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	SortOrder   int          `json:"sortOrder"`
	Sport       *MLBSportRef `json:"sport,omitempty"`
}

// MLBVenuesResponse represents the response from the venues endpoint
type MLBVenuesResponse struct {
	Copyright string     `json:"copyright"`
	Venues    []MLBVenue `json:"venues"`
}

// MLBVenue represents a ballpark venue
type MLBVenue struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Link   string `json:"link"`
	Active bool   `json:"active"`
	Season string `json:"season,omitempty"`
}
