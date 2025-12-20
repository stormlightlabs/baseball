package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"stormlightlabs.org/baseball/internal/core"
)

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

// writeError writes an error response with the appropriate HTTP status code.
// Returns 404 for NotFoundError, 500 for all other errors.
func writeError(w http.ResponseWriter, err error) {
	if core.IsNotFound(err) {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	writeError(w, err)
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

func getFloatQuery(r *http.Request, key string, defaultVal float64) float64 {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}

	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultVal
	}
	return f
}
