package core

// TeamCrosswalkFilter controls team crosswalk queries.
type TeamCrosswalkFilter struct {
	Season      *SeasonYear
	AllSeasons  bool
	TeamID      *TeamID
	FranchiseID *FranchiseID
	MLBAMTeamID *int
}

// PlayerCrosswalkFilter controls player crosswalk queries.
type PlayerCrosswalkFilter struct {
	PlayerID *PlayerID
	RetroID  *RetroPlayerID
	MLBAMID  *int
}

// TeamCrosswalkRow represents a season-scoped team/franchise/MLBAM mapping.
type TeamCrosswalkRow struct {
	Season          SeasonYear   `json:"season" swaggertype:"integer"`
	TeamID          *TeamID      `json:"team_id,omitempty" swaggertype:"string"`
	FranchiseID     *FranchiseID `json:"franchise_id,omitempty" swaggertype:"string"`
	TeamName        string       `json:"team_name,omitempty"`
	League          *LeagueID    `json:"league,omitempty" swaggertype:"string"`
	MLBAMTeamID     *int         `json:"mlbam_team_id,omitempty"`
	MLBAbbreviation string       `json:"mlb_abbreviation,omitempty"`
	MLBTeamCode     string       `json:"mlb_team_code,omitempty"`
	MLBFileCode     string       `json:"mlb_file_code,omitempty"`
	MLBTeamName     string       `json:"mlb_team_name,omitempty"`
	MLBFranchise    string       `json:"mlb_franchise_name,omitempty"`
	MLBClubName     string       `json:"mlb_club_name,omitempty"`
	MatchMethod     string       `json:"match_method,omitempty"`
	Confidence      string       `json:"confidence,omitempty"`
	Source          string       `json:"source,omitempty"`
}

// TeamCrosswalkResponse wraps team crosswalk rows.
type TeamCrosswalkResponse struct {
	Season     *SeasonYear        `json:"season,omitempty" swaggertype:"integer"`
	AllSeasons bool               `json:"all_seasons"`
	Total      int                `json:"total"`
	Rows       []TeamCrosswalkRow `json:"rows"`
}

// PlayerCrosswalkRow represents player ID mapping across systems.
type PlayerCrosswalkRow struct {
	PlayerID   *PlayerID      `json:"player_id,omitempty" swaggertype:"string"`
	RetroID    *RetroPlayerID `json:"retro_id,omitempty" swaggertype:"string"`
	MLBAMID    *int           `json:"mlbam_id,omitempty"`
	BBRefID    string         `json:"bbref_id,omitempty"`
	FullName   string         `json:"full_name,omitempty"`
	FirstName  string         `json:"first_name,omitempty"`
	LastName   string         `json:"last_name,omitempty"`
	Source     string         `json:"source,omitempty"`
	Confidence string         `json:"confidence,omitempty"`
}

// PlayerCrosswalkResponse wraps player crosswalk rows.
type PlayerCrosswalkResponse struct {
	Total int                  `json:"total"`
	Rows  []PlayerCrosswalkRow `json:"rows"`
}
