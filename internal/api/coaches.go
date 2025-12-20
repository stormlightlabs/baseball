package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type CoachRoutes struct {
	repo core.CoachRepository
}

func NewCoachRoutes(repo core.CoachRepository) *CoachRoutes {
	return &CoachRoutes{repo: repo}
}

func (cr *CoachRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/coaches", cr.handleListCoaches)
	mux.HandleFunc("GET /v1/coaches/{id}", cr.handleGetCoach)
	mux.HandleFunc("GET /v1/coaches/{id}/seasons", cr.handleCoachSeasons)
}

// handleListCoaches godoc
// @Summary List coaches
// @Description Get a paginated list of all coaches
// @Tags coaches
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /coaches [get]
func (cr *CoachRoutes) handleListCoaches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 50),
	}

	coaches, err := cr.repo.List(ctx, pagination)
	if err != nil {
		writeError(w, err)
		return
	}

	total := len(coaches)

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    coaches,
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		Total:   total,
	})
}

// handleGetCoach godoc
// @Summary Get coach by ID
// @Description Get detailed information about a specific coach. Note: A person who later became a manager will appear in both the coaches and managers endpoints.
// @Tags coaches
// @Accept json
// @Produce json
// @Param id path string true "Coach ID (playerID)" example(roberda07)
// @Success 200 {object} core.Coach
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /coaches/{id} [get]
func (cr *CoachRoutes) handleGetCoach(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	coachID := core.PlayerID(r.PathValue("id"))

	coach, err := cr.repo.GetByID(ctx, coachID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, coach)
}

// handleCoachSeasons godoc
// @Summary Get coach season records
// @Description Get all season coaching records for a specific coach including team, role, and dates. Note: This endpoint returns only coaching seasons, not managerial seasons. For managerial records, use the /managers/{id}/seasons endpoint.
// @Tags coaches
// @Accept json
// @Produce json
// @Param id path string true "Coach ID (playerID)" example(roberda07)
// @Success 200 {object} CoachSeasonsResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /coaches/{id}/seasons [get]
func (cr *CoachRoutes) handleCoachSeasons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	coachID := core.PlayerID(r.PathValue("id"))

	records, err := cr.repo.SeasonRecords(ctx, coachID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, CoachSeasonsResponse{
		CoachID: coachID,
		Seasons: records,
	})
}

// CoachSeasonsResponse wraps coach season records
type CoachSeasonsResponse struct {
	CoachID core.PlayerID            `json:"coach_id"`
	Seasons []core.CoachSeasonRecord `json:"seasons"`
}
