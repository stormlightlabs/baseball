package core

// StatProvider identifies whose formulas / constants you followed.
type StatProvider string

const (
	StatProviderUnknown   StatProvider = "unknown"
	StatProviderFanGraphs StatProvider = "fangraphs"
	StatProviderBBRef     StatProvider = "baseball_reference"
	StatProviderInternal  StatProvider = "internal" // your own
)

// StatContext anchors a stat line to season/league/park environment.
type StatContext struct {
	Season      SeasonYear   `json:"season"`           // 0 if multi-year or career
	League      *LeagueID    `json:"league,omitempty"` // "AL","NL", etc.
	Provider    StatProvider `json:"provider"`         // fangraphs, bbref, internal
	ParkNeutral bool         `json:"park_neutral"`     // already park-adjusted?
	RegSeason   bool         `json:"regular_season"`   // true = reg season, false = postseason/mix
}

// AdvancedBattingStats is a snapshot of derived hitting metrics for a player
// in a specific context (season, split, career, game).
type AdvancedBattingStats struct {
	PlayerID PlayerID    `json:"player_id"`
	TeamID   *TeamID     `json:"team_id,omitempty"` // nil for multi-team lines
	Context  StatContext `json:"context"`

	// Underlying counting stats for transparency.
	PA      int `json:"pa"`
	AB      int `json:"ab"`
	H       int `json:"h"`
	Doubles int `json:"doubles"`
	Triples int `json:"triples"`
	HR      int `json:"hr"`
	BB      int `json:"bb"`
	IBB     int `json:"ibb"`
	HBP     int `json:"hbp"`
	SF      int `json:"sf"`
	SH      int `json:"sh"`
	SO      int `json:"so"`

	FB *int `json:"fb,omitempty"` // fly balls
	GB *int `json:"gb,omitempty"` // ground balls
	LD *int `json:"ld,omitempty"` // line drives
	PU *int `json:"pu,omitempty"` // pop-ups

	AVG float64 `json:"avg"` // BA
	OBP float64 `json:"obp"`
	SLG float64 `json:"slg"`
	OPS float64 `json:"ops"`
	ISO float64 `json:"iso"` // ISO = SLG - AVG

	KRate  float64  `json:"k_rate"`          // K%
	BBRate float64  `json:"bb_rate"`         // BB%
	HRFB   *float64 `json:"hr_fb,omitempty"` // HR/FB, if FB known
	BABIP  float64  `json:"babip"`

	WOBA    float64 `json:"woba"`
	WRAA    float64 `json:"wraa"`     // weighted runs above average
	WRC     float64 `json:"wrc"`      // weighted runs created
	WRCPlus int     `json:"wrc_plus"` // wRC+ (100 = league avg), park/league adjusted

	// OPS+ if you choose to compute it (BBRef-style).
	OPSPlus *int `json:"ops_plus,omitempty"`

	// WAR ascribed to offense (offensive component only; not full WAR).
	OffWAR *float64 `json:"off_war,omitempty"`
}

// AdvancedPitchingStats is a snapshot of derived pitching metrics for a player
// in a given context (season, split, career, game).
type AdvancedPitchingStats struct {
	PlayerID PlayerID    `json:"player_id"`
	TeamID   *TeamID     `json:"team_id,omitempty"`
	Context  StatContext `json:"context"`

	// Underlying counting stats.
	IPOuts int `json:"ip_outs"` // outs recorded (IP * 3)
	BF     int `json:"bf"`      // batters faced
	H      int `json:"h"`
	R      int `json:"r"`
	ER     int `json:"er"`
	HR     int `json:"hr"`
	BB     int `json:"bb"`
	IBB    int `json:"ibb"`
	HBP    int `json:"hbp"`
	SO     int `json:"so"`

	// Basic rate stats.
	ERA    float64 `json:"era"`  // 9 * ER/IP
	WHIP   float64 `json:"whip"` // (BB+H)/IP
	KPer9  float64 `json:"k_per_9"`
	BBPer9 float64 `json:"bb_per_9"`
	HRPer9 float64 `json:"hr_per_9"`

	// Fielding-independent metrics.
	FIP  float64  `json:"fip"`
	XFIP *float64 `json:"xfip,omitempty"`

	// Context-normalized rate stats.
	ERAPlus *int `json:"era_plus,omitempty"` // ERA+ (100 = league avg)
	// FIP- / xFIP- (FanGraphs style: lower is better, 100 = league avg), if you want.
	FIPMinus  *int `json:"fip_minus,omitempty"`
	XFIPMinus *int `json:"xfip_minus,omitempty"`

	// WAR components for pitchers.
	// Total WAR from chosen provider (FG, BR, etc.).
	WAR *float64 `json:"war,omitempty"`

	// Optional breakdowns if you compute them.
	RA9WAR          *float64 `json:"ra9_war,omitempty"` // BBRef-style
	FIPWAR          *float64 `json:"fip_war,omitempty"` // FG-style
	ReplacementRuns *float64 `json:"replacement_runs,omitempty"`
}

// PlayerWARSummary aggregates WAR components for a player in a given context.
type PlayerWARSummary struct {
	PlayerID PlayerID    `json:"player_id"`
	TeamID   *TeamID     `json:"team_id,omitempty"`
	Context  StatContext `json:"context"`

	// Total WAR (position-player + pitcher if two-way).
	WAR float64 `json:"war"`

	// Components for position players.
	BattingRuns      *float64 `json:"batting_runs,omitempty"` // wRAA-like
	BaseRunningRuns  *float64 `json:"baserunning_runs,omitempty"`
	FieldingRuns     *float64 `json:"fielding_runs,omitempty"`
	PositionalRuns   *float64 `json:"positional_runs,omitempty"`
	LeagueAdjustment *float64 `json:"league_adjustment,omitempty"`
	ReplacementRuns  *float64 `json:"replacement_runs,omitempty"`

	// Alternate view for pitchers.
	PitchingRunsAboveReplacement *float64 `json:"pitching_rar,omitempty"`
}

// PlateAppearanceLeverage captures leverage and win-prob data
// for a single plate appearance (or batter-pitcher confrontation).
type PlateAppearanceLeverage struct {
	GameID      GameID   `json:"game_id"`
	EventID     int      `json:"event_id"` // index within game
	BatterID    PlayerID `json:"batter_id"`
	PitcherID   PlayerID `json:"pitcher_id"`
	Inning      int      `json:"inning"`
	TopOfInning bool     `json:"top_of_inning"`

	// State before the PA.
	HomeScoreBefore int    `json:"home_score_before"`
	AwayScoreBefore int    `json:"away_score_before"`
	OutsBefore      int    `json:"outs_before"`
	BasesBefore     string `json:"bases_before"` // "100", "011", etc.

	// Win expectancy and leverage index before/after the PA.
	WinExpectancyBefore float64 `json:"we_before"` // home team WE
	WinExpectancyAfter  float64 `json:"we_after"`
	LeverageIndex       float64 `json:"li"`        // LI at start of PA
	WEChange            float64 `json:"we_change"` // signed WEAfter - WEBefore

	// Result summary.
	Description string `json:"description"` // e.g. "HR to LF", "BB", "K swinging"
}

// PlayerLeverageSummary aggregates leverage metrics for a player
// (most commonly for pitchers, but can be used for hitters too).
type PlayerLeverageSummary struct {
	PlayerID PlayerID    `json:"player_id"`
	TeamID   *TeamID     `json:"team_id,omitempty"`
	Context  StatContext `json:"context"`

	// Average LI of all plate appearances faced (aLI).
	AvgLeverageIndex float64 `json:"avg_li"`

	// Average LI at the moment the player entered games (gmLI).
	GameEntryLeverageIndex *float64 `json:"gm_li,omitempty"`
	LowLeveragePA          int      `json:"low_leverage_pa"`    // LI < 0.85
	MediumLeveragePA       int      `json:"medium_leverage_pa"` // 0.85 <= LI <= 2.0
	HighLeveragePA         int      `json:"high_leverage_pa"`   // LI > 2.0
	LowLeverageIPOuts      int      `json:"low_leverage_ip_outs"`
	MediumLeverageIPOuts   int      `json:"medium_leverage_ip_outs"`
	HighLeverageIPOuts     int      `json:"high_leverage_ip_outs"`

	// Aggregate WPA (win probability added) if you calculate it.
	WinProbabilityAdded *float64 `json:"wpa,omitempty"`
}

// GameWinProbabilitySummary gives high-level win-prob info for a game.
type GameWinProbabilitySummary struct {
	GameID   GameID `json:"game_id"`
	Season   int    `json:"season"`
	HomeTeam TeamID `json:"home_team"`
	AwayTeam TeamID `json:"away_team"`

	HomeWinProbStart float64 `json:"home_win_prob_start"`
	HomeWinProbEnd   float64 `json:"home_win_prob_end"`

	BiggestPositiveSwing *PlateAppearanceLeverage `json:"biggest_positive_swing,omitempty"`
	BiggestNegativeSwing *PlateAppearanceLeverage `json:"biggest_negative_swing,omitempty"`
}

type ParkFactor struct {
	ParkID string `json:"park_id"` // ParkID
	Season int    `json:"season"`  // season year

	RunsFactor float64 `json:"runs_factor"` // overall runs (100 = neutral)
	HRFactor   float64 `json:"hr_factor"`   // home run factor
	BBFactor   float64 `json:"bb_factor"`   // walks factor (optional)
	HFactor    float64 `json:"h_factor"`    // hits factor (optional)

	RunsFactorLHB *float64 `json:"runs_factor_lhb,omitempty"`
	RunsFactorRHB *float64 `json:"runs_factor_rhb,omitempty"`

	GamesSampled int    `json:"games_sampled"` // # of games used
	Provider     string `json:"provider"`      // "internal", "fangraphs-like", etc.
	MultiYear    bool   `json:"multi_year"`    // if you averaged multiple seasons
}

// WOBAConstant represents season-specific weights and constants for wOBA calculation.
type WOBAConstant struct {
	Season int `json:"season"`

	// wOBA weights for each event type
	WBB  float64 `json:"w_bb"`  // unintentional walk
	WHBP float64 `json:"w_hbp"` // hit by pitch
	W1B  float64 `json:"w_1b"`  // single
	W2B  float64 `json:"w_2b"`  // double
	W3B  float64 `json:"w_3b"`  // triple
	WHR  float64 `json:"w_hr"`  // home run

	// wOBA scale and league average
	WOBAScale float64 `json:"woba_scale"` // scaling factor
	WOBA      float64 `json:"woba"`       // league average wOBA

	// Base running run values
	RunSB float64 `json:"run_sb"` // stolen base runs
	RunCS float64 `json:"run_cs"` // caught stealing runs

	// League context
	RPA float64 `json:"r_pa"` // runs per plate appearance
	RW  float64 `json:"r_w"`  // runs per win

	CFIP float64 `json:"c_fip"` // FIP constant
}

// LeagueConstant represents league-specific constants for park/league adjustments.
type LeagueConstant struct {
	Season int      `json:"season"`
	League LeagueID `json:"league"` // AL or NL

	WOBAAvg       *float64 `json:"woba_avg,omitempty"`       // league average wOBA
	WRCPerPA      *float64 `json:"wrc_per_pa,omitempty"`     // wRC per PA (excluding pitchers)
	RunsPerWin    *float64 `json:"runs_per_win,omitempty"`   // league-specific runs per win
	ReplacementPA *float64 `json:"replacement_pa,omitempty"` // replacement level runs per PA

	TotalPA   *int64 `json:"total_pa,omitempty"`   // total plate appearances
	TotalRuns *int64 `json:"total_runs,omitempty"` // total runs scored
}

// ParkFactorRow represents detailed park factors from FanGraphs for a specific park/season.
type ParkFactorRow struct {
	ParkID string  `json:"park_id"`
	Season int     `json:"season"`
	TeamID *TeamID `json:"team_id,omitempty"` // team playing at this park

	// Overall park factors (100 = neutral, >100 = hitter-friendly)
	Basic5yr *int `json:"basic_5yr,omitempty"` // 5-year regressed (most stable)
	Basic3yr *int `json:"basic_3yr,omitempty"` // 3-year regressed
	Basic1yr *int `json:"basic_1yr,omitempty"` // single year (most volatile)

	// Component park factors
	Factor1B   *int `json:"factor_1b,omitempty"`   // singles
	Factor2B   *int `json:"factor_2b,omitempty"`   // doubles
	Factor3B   *int `json:"factor_3b,omitempty"`   // triples
	FactorHR   *int `json:"factor_hr,omitempty"`   // home runs
	FactorSO   *int `json:"factor_so,omitempty"`   // strikeouts
	FactorBB   *int `json:"factor_bb,omitempty"`   // walks
	FactorGB   *int `json:"factor_gb,omitempty"`   // ground balls
	FactorFB   *int `json:"factor_fb,omitempty"`   // fly balls
	FactorLD   *int `json:"factor_ld,omitempty"`   // line drives
	FactorIFFB *int `json:"factor_iffb,omitempty"` // infield fly balls
	FactorFIP  *int `json:"factor_fip,omitempty"`  // FIP
}
