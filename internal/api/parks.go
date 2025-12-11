package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type ParkRoutes struct {
	repo core.ParkRepository
}

func NewParkRoutes(repo core.ParkRepository) *ParkRoutes {
	return &ParkRoutes{repo: repo}
}

func (pr *ParkRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/parks", pr.handleListParks)
	mux.HandleFunc("GET /v1/parks/{park_id}", pr.handleGetPark)
}

// handleListParks godoc
// @Summary List ballparks
// @Description Get a paginated list of all ballparks
// @Tags parks
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /parks [get]
func (pr *ParkRoutes) handleListParks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 50),
	}

	parks, err := pr.repo.List(ctx, pagination)
	if err != nil {
		writeError(w, err)
		return
	}

	total := len(parks)

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    parks,
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		Total:   total,
	})
}

// handleGetPark godoc
// @Summary Get ballpark by ID
// @Description Get detailed information about a specific ballpark
// @Tags parks
// @Accept json
// @Produce json
// @Param park_id path string true "Park ID (park key)"
// @Success 200 {object} core.Park
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /parks/{park_id} [get]
func (pr *ParkRoutes) handleGetPark(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parkID := core.ParkID(r.PathValue("park_id"))

	park, err := pr.repo.GetByID(ctx, parkID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, park)
}
