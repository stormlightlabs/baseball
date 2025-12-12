package api

import (
	"net/http"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

const apiVersion = "1.0.0"

type MetaRoutes struct {
	repo core.MetaRepository
}

func NewMetaRoutes(repo core.MetaRepository) *MetaRoutes {
	return &MetaRoutes{repo: repo}
}

func (mr *MetaRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/meta", mr.handleMeta)
	mux.HandleFunc("GET /v1/meta/datasets", mr.handleDatasetStatus)
}

type metaResponse struct {
	Version      string                     `json:"version"`
	GeneratedAt  time.Time                  `json:"generated_at"`
	Coverage     map[string]datasetCoverage `json:"coverage"`
	SchemaHashes map[string]string          `json:"schema_hashes"`
	Datasets     []core.DatasetStatus       `json:"datasets"`
}

type datasetCoverage struct {
	From *core.SeasonYear `json:"from,omitempty"`
	To   *core.SeasonYear `json:"to,omitempty"`
}

// handleMeta godoc
// @Summary API metadata
// @Description Returns API version, dataset freshness, coverage, and schema fingerprints
// @Tags meta
// @Accept json
// @Produce json
// @Success 200 {object} metaResponse
// @Failure 500 {object} ErrorResponse
// @Router /meta [get]
func (mr *MetaRoutes) handleMeta(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	datasets, err := mr.repo.DatasetStatuses(ctx)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	minLahman, maxLahman, minRetro, maxRetro, err := mr.repo.SeasonCoverage(ctx)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	schemaHashes, err := mr.repo.SchemaHashes(ctx)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	resp := metaResponse{
		Version:     apiVersion,
		GeneratedAt: time.Now().UTC(),
		Coverage: map[string]datasetCoverage{
			"lahman":     makeCoverage(minLahman, maxLahman),
			"retrosheet": makeCoverage(minRetro, maxRetro),
		},
		SchemaHashes: schemaHashes,
		Datasets:     datasets,
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleDatasetStatus godoc
// @Summary Dataset status
// @Description Returns dataset ETL metadata and coverage
// @Tags meta
// @Accept json
// @Produce json
// @Success 200 {array} core.DatasetStatus
// @Failure 500 {object} ErrorResponse
// @Router /meta/datasets [get]
func (mr *MetaRoutes) handleDatasetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	datasets, err := mr.repo.DatasetStatuses(ctx)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, datasets)
}

func makeCoverage(from, to core.SeasonYear) datasetCoverage {
	c := datasetCoverage{}
	if from != 0 {
		f := from
		c.From = &f
	}
	if to != 0 {
		t := to
		c.To = &t
	}
	return c
}
