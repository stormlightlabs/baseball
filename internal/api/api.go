package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

// Anything that can add its endpoints to a mux.
type Registrar interface {
	RegisterRoutes(mux *http.ServeMux)
}

type PaginatedResponse struct {
	Data    any `json:"data"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

func NewPaginatedResponse(data any, p, pp, t int) PaginatedResponse {
	return PaginatedResponse{Data: data, Page: p, PerPage: pp, Total: t}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// PlayerSeasonsResponse wraps batting and pitching season stats
type PlayerSeasonsResponse struct {
	Batting  []core.PlayerBattingSeason  `json:"batting"`
	Pitching []core.PlayerPitchingSeason `json:"pitching"`
}

func NewPlayerSeasonsResponse(b []core.PlayerBattingSeason, p []core.PlayerPitchingSeason) PlayerSeasonsResponse {
	return PlayerSeasonsResponse{Batting: b, Pitching: p}
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
	Page    int                        `json:"page"`
	PerPage int                        `json:"per_page"`
	Total   int                        `json:"total"`
}

// PitchingLeadersResponse wraps season pitching leaders with metadata
type PitchingLeadersResponse struct {
	Year    core.SeasonYear             `json:"year" swaggertype:"integer"`
	Stat    string                      `json:"stat"`
	League  *core.LeagueID              `json:"league,omitempty" swaggertype:"string"`
	Leaders []core.PlayerPitchingSeason `json:"leaders"`
	Page    int                         `json:"page"`
	PerPage int                         `json:"per_page"`
	Total   int                         `json:"total"`
}

// CareerBattingLeadersResponse wraps career batting leaders with metadata
type CareerBattingLeadersResponse struct {
	Stat    string                     `json:"stat"`
	Leaders []core.PlayerBattingSeason `json:"leaders"`
	Page    int                        `json:"page"`
	PerPage int                        `json:"per_page"`
	Total   int                        `json:"total"`
}

// CareerPitchingLeadersResponse wraps career pitching leaders with metadata
type CareerPitchingLeadersResponse struct {
	Stat    string                      `json:"stat"`
	Leaders []core.PlayerPitchingSeason `json:"leaders"`
	Page    int                         `json:"page"`
	PerPage int                         `json:"per_page"`
	Total   int                         `json:"total"`
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

func NewPlayerBattingStatsResponse(c core.PlayerBattingSeason, s []core.PlayerBattingSeason) PlayerBattingStatsResponse {
	return PlayerBattingStatsResponse{Career: c, Seasons: s}
}

// PlayerPitchingStatsResponse wraps player career and season pitching stats
type PlayerPitchingStatsResponse struct {
	Career  core.PlayerPitchingSeason   `json:"career"`
	Seasons []core.PlayerPitchingSeason `json:"seasons"`
}

func NewPlayerPitchingStatsResponse(c core.PlayerPitchingSeason, s []core.PlayerPitchingSeason) PlayerPitchingStatsResponse {
	return PlayerPitchingStatsResponse{Career: c, Seasons: s}
}

// SalarySummaryResponse wraps salary summary data
type SalarySummaryResponse struct {
	Data []core.SalarySummary `json:"data"`
}

func NewSalarySummaryResponse(data []core.SalarySummary) SalarySummaryResponse {
	return SalarySummaryResponse{Data: data}
}
