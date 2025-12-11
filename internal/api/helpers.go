package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"stormlightlabs.org/baseball/internal/core"
)

type PaginatedResponse struct {
	Data    any `json:"data"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// PlayerSeasonsResponse wraps batting and pitching season stats
type PlayerSeasonsResponse struct {
	Batting  []core.PlayerBattingSeason  `json:"batting"`
	Pitching []core.PlayerPitchingSeason `json:"pitching"`
}

// HallOfFameResponse wraps Hall of Fame records
type HallOfFameResponse struct {
	Records []core.HallOfFameRecord `json:"records"`
}

// BattingLeadersResponse wraps season batting leaders with metadata
type BattingLeadersResponse struct {
	Year    core.SeasonYear            `json:"year" swaggertype:"integer"`
	Stat    string                     `json:"stat"`
	League  *core.LeagueID             `json:"league,omitempty" swaggertype:"string"`
	Leaders []core.PlayerBattingSeason `json:"leaders"`
}

// PitchingLeadersResponse wraps season pitching leaders with metadata
type PitchingLeadersResponse struct {
	Year    core.SeasonYear             `json:"year" swaggertype:"integer"`
	Stat    string                      `json:"stat"`
	League  *core.LeagueID              `json:"league,omitempty" swaggertype:"string"`
	Leaders []core.PlayerPitchingSeason `json:"leaders"`
}

// FranchisesResponse wraps franchise list with total count
type FranchisesResponse struct {
	Franchises []core.Franchise `json:"franchises"`
	Total      int              `json:"total"`
}

// HealthResponse is the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

// TODO: map error types to HTTP status codes
func writeError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
}

func getIntQuery(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

func getIntPathValue(r *http.Request, key string) int {
	val := r.PathValue(key)
	if val == "" {
		return 0
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return i
}
