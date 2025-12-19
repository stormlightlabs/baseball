package api

import (
	"net/http"

	"stormlightlabs.org/baseball/internal/core"
)

type ManagerRoutes struct {
	repo core.ManagerRepository
}

func NewManagerRoutes(repo core.ManagerRepository) *ManagerRoutes {
	return &ManagerRoutes{repo: repo}
}

func (mr *ManagerRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/managers", mr.handleListManagers)
	mux.HandleFunc("GET /v1/managers/{manager_id}", mr.handleGetManager)
	mux.HandleFunc("GET /v1/managers/{manager_id}/seasons", mr.handleManagerSeasons)
}

// handleListManagers godoc
// @Summary List managers
// @Description Get a paginated list of all managers
// @Tags managers
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param per_page query integer false "Results per page" default(50)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /managers [get]
func (mr *ManagerRoutes) handleListManagers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pagination := core.Pagination{
		Page:    getIntQuery(r, "page", 1),
		PerPage: getIntQuery(r, "per_page", 50),
	}

	managers, err := mr.repo.List(ctx, pagination)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	total := len(managers)

	writeJSON(w, http.StatusOK, PaginatedResponse{
		Data:    managers,
		Page:    pagination.Page,
		PerPage: pagination.PerPage,
		Total:   total,
	})
}

// handleGetManager godoc
// @Summary Get manager by ID
// @Description Get detailed information about a specific manager including extended biodata from Retrosheet (debut/last game, full name, use name)
// @Tags managers
// @Accept json
// @Produce json
// @Param manager_id path string true "Manager ID (playerID)" example(roberda07)
// @Success 200 {object} core.Manager
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /managers/{manager_id} [get]
func (mr *ManagerRoutes) handleGetManager(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	managerID := core.ManagerID(r.PathValue("manager_id"))

	manager, err := mr.repo.GetByID(ctx, managerID)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, manager)
}

// handleManagerSeasons godoc
// @Summary Get manager season records
// @Description Get all season records for a specific manager including wins, losses, and team rank
// @Tags managers
// @Accept json
// @Produce json
// @Param manager_id path string true "Manager ID (playerID)" example(roberda07)
// @Success 200 {object} ManagerSeasonsResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /managers/{manager_id}/seasons [get]
func (mr *ManagerRoutes) handleManagerSeasons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	managerID := core.ManagerID(r.PathValue("manager_id"))

	records, err := mr.repo.SeasonRecords(ctx, managerID)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ManagerSeasonsResponse{
		ManagerID: managerID,
		Seasons:   records,
	})
}

// ManagerSeasonsResponse wraps manager season records
type ManagerSeasonsResponse struct {
	ManagerID core.ManagerID             `json:"manager_id"`
	Seasons   []core.ManagerSeasonRecord `json:"seasons"`
}
