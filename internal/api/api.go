package api

import "net/http"

// Anything that can add its endpoints to a mux.
type Registrar interface {
	RegisterRoutes(mux *http.ServeMux)
}
