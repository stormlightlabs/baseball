package core

// StreakKind identifies what is being tracked.
type StreakKind string

const (
	// Hitting streak: consecutive games with at least one hit.
	StreakKindHitting StreakKind = "hitting"

	// Scoreless innings streak for a pitcher or team staff.
	StreakKindScorelessInnings StreakKind = "scoreless_innings"
)

// StreakEntityType clarifies what owns the streak.
type StreakEntityType string

const (
	StreakEntityPlayer StreakEntityType = "player"
	StreakEntityTeam   StreakEntityType = "team"
)

// Streak represents a contiguous run of games or innings.
type Streak struct {
	ID         string           `json:"id"`          // internal identifier for the streak
	Kind       StreakKind       `json:"kind"`        // hitting, scoreless_innings
	EntityType StreakEntityType `json:"entity_type"` // player or team
	EntityID   string           `json:"entity_id"`   // PlayerID or TeamID (as string)
	Label      string           `json:"label"`       // human-readable label

	Season int `json:"season"` // season year (e.g. 2024)

	// Length of the streak:
	// - For hitting: number of consecutive games with ≥1 hit.
	// - For scoreless innings: total consecutive innings without allowing a run.
	Length int `json:"length"`

	// Game/inning bounds of the streak.
	StartGameID GameID `json:"start_game_id"`
	EndGameID   GameID `json:"end_game_id"`

	// ISO-8601 dates (YYYY-MM-DD) to keep JSON lightweight and frontend-friendly.
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`

	// Timeline for charts or detailed views.
	Timeline []StreakPoint `json:"timeline,omitempty"`
}

// StreakPoint captures per-game / per-appearance detail within a streak.
type StreakPoint struct {
	GameID GameID `json:"game_id"`
	Date   string `json:"date"` // YYYY-MM-DD

	// Index within the streak (1..Length).
	Index int `json:"index"`

	// For hitting streaks.
	PlateAppearances int `json:"pa,omitempty"`
	AtBats           int `json:"ab,omitempty"`
	Hits             int `json:"h,omitempty"`

	// For scoreless innings streaks.
	InningsPitched float64 `json:"ip,omitempty"` // e.g. 7.2
	RunsAllowed    int     `json:"runs_allowed,omitempty"`
}

// RunDifferentialSeries represents run differential across a season.
type RunDifferentialSeries struct {
	EntityType string `json:"entity_type"` // usually "team"
	EntityID   string `json:"entity_id"`   // TeamID
	Season     int    `json:"season"`

	// Season totals.
	GamesPlayed     int `json:"games_played"`
	RunsScored      int `json:"runs_scored"`
	RunsAllowed     int `json:"runs_allowed"`
	RunDifferential int `json:"run_differential"` // RS - RA

	// Per-game series, ordered chronologically.
	Games []RunDifferentialGamePoint `json:"games"`

	// Rolling-window aggregations (e.g. last 10, 20, 30 games).
	Rolling []RunDifferentialWindow `json:"rolling,omitempty"`
}

// RunDifferentialGamePoint: one row per game.
type RunDifferentialGamePoint struct {
	GameID GameID `json:"game_id"`
	Date   string `json:"date"` // YYYY-MM-DD

	OpponentID TeamID `json:"opponent_id"`
	Home       bool   `json:"home"`

	RunsScored     int `json:"runs_scored"`
	RunsAllowed    int `json:"runs_allowed"`
	Differential   int `json:"differential"`    // single-game RS - RA
	CumulativeDiff int `json:"cumulative_diff"` // running season sum
}

// RunDifferentialWindow: precomputed rolling-window stat.
type RunDifferentialWindow struct {
	WindowSize int    `json:"window_size"` // number of games (e.g. 10, 20, 30)
	Label      string `json:"label"`       // e.g. "last_10", "last_30"

	// Ordered by window end game/date
	Points []RunDifferentialWindowPoint `json:"points"`
}

type RunDifferentialWindowPoint struct {
	EndGameID GameID `json:"end_game_id"`
	EndDate   string `json:"end_date"` // YYYY-MM-DD

	GamesInWindow   int `json:"games_in_window"`
	RunsScored      int `json:"runs_scored"`
	RunsAllowed     int `json:"runs_allowed"`
	RunDifferential int `json:"run_differential"` // RS - RA in window
}

// WinProbabilityCurve is the win-probability timeline for a single game, derived from play-by-play events.
type WinProbabilityCurve struct {
	GameID   GameID `json:"game_id"`
	Season   int    `json:"season"`
	HomeTeam TeamID `json:"home_team"`
	AwayTeam TeamID `json:"away_team"`

	// One point per play/event, ordered by EventIndex ascending.
	Points []WinProbabilityPoint `json:"points"`
}

// WinProbabilityPoint describes state & win probabilities AFTER a single event.
type WinProbabilityPoint struct {
	EventIndex  int  `json:"event_index"` // 1..N within the game
	Inning      int  `json:"inning"`      // 1..9, 10+
	TopOfInning bool `json:"top_of_inning"`

	// Score state after the event.
	HomeScore int `json:"home_score"`
	AwayScore int `json:"away_score"`
	Outs      int `json:"outs"` // 0..3

	// Simple base-state encoding
	// Example: "100" = runner on 1st only, "011" = runners on 2nd & 3rd.
	Bases string `json:"bases"`

	// Win probabilities for each team AFTER the event (0.0–1.0).
	HomeWinProb float64 `json:"home_win_prob"`
	AwayWinProb float64 `json:"away_win_prob"`

	Description string `json:"description,omitempty"` // e.g. "Trout homers to LF (fly ball)."
}
