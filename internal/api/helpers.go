package api

import (
	"encoding/json"
	"fmt"
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

// CareerBattingLeadersResponse wraps career batting leaders with metadata
type CareerBattingLeadersResponse struct {
	Stat    string                     `json:"stat"`
	Leaders []core.PlayerBattingSeason `json:"leaders"`
}

// CareerPitchingLeadersResponse wraps career pitching leaders with metadata
type CareerPitchingLeadersResponse struct {
	Stat    string                      `json:"stat"`
	Leaders []core.PlayerPitchingSeason `json:"leaders"`
}

// FranchisesResponse wraps franchise list with total count
type FranchisesResponse struct {
	Franchises []core.Franchise `json:"franchises"`
	Total      int              `json:"total"`
}

// AwardsListResponse wraps awards list
type AwardsListResponse struct {
	Awards []core.Award `json:"awards"`
}

// HealthResponse is the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// PlayerBattingStatsResponse wraps player career and season batting stats
type PlayerBattingStatsResponse struct {
	Career  core.PlayerBattingSeason   `json:"career"`
	Seasons []core.PlayerBattingSeason `json:"seasons"`
}

// PlayerPitchingStatsResponse wraps player career and season pitching stats
type PlayerPitchingStatsResponse struct {
	Career  core.PlayerPitchingSeason   `json:"career"`
	Seasons []core.PlayerPitchingSeason `json:"seasons"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("writeJSON marshal error: %v", err)
		return
	}

	if _, err := w.Write(data); err != nil {
		log.Printf("writeJSON write error: %v", err)
	}
}

func writeInternalServerError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
}

func writeBadRequest(w http.ResponseWriter, err string) {
	writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err})
}

func writeNotFound(w http.ResponseWriter, r string) {
	writeJSON(w, http.StatusNotFound, ErrorResponse{Error: fmt.Sprintf("%v not found", r)})
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
