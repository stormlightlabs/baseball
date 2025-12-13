// Package api provides HTTP handlers for the Baseball API
//
// TODO: finish tag docs
//
// @title Baseball API
// @description.markdown
// @version 1.0
// @BasePath /v1
//
// @contact.name API Support
// @contact.url https://github.com/stormlightlabs/baseball
// @contact.email info@stormlightlabs.org
//
// @license.name MPL-2.0
// @license.url https://opensource.org/license/mpl-2-0
//
// @tag.name allstar
// @tag.description MLB All-Star game data
//
// @tag.name games
// @tag.description Game data
package api

import (
	"database/sql"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	"stormlightlabs.org/baseball/internal/cache"
	docs "stormlightlabs.org/baseball/internal/docs"
	"stormlightlabs.org/baseball/internal/echo"
	"stormlightlabs.org/baseball/internal/repository"
)

type Server struct {
	mux *http.ServeMux
}

func NewServer(db *sql.DB, cacheClient *cache.Client) *Server {
	echo.Info("Initializing repositories...")

	playerRepo := repository.NewPlayerRepository(db, cacheClient)
	teamRepo := repository.NewTeamRepository(db, cacheClient)
	statsRepo := repository.NewStatsRepository(db)
	awardRepo := repository.NewAwardRepository(db)
	gameRepo := repository.NewGameRepository(db, cacheClient)
	playRepo := repository.NewPlayRepository(db)
	pitchRepo := repository.NewPitchRepository(db)
	metaRepo := repository.NewMetaRepository(db)
	managerRepo := repository.NewManagerRepository(db)
	parkRepo := repository.NewParkRepository(db)
	umpireRepo := repository.NewUmpireRepository(db)
	postseasonRepo := repository.NewPostseasonRepository(db)
	ejectionRepo := repository.NewEjectionRepository(db)
	derivedRepo := repository.NewDerivedStatsRepository(db)

	advancedStatsRepo := repository.NewAdvancedStatsRepository(db)
	leverageRepo := repository.NewLeverageRepository(db)
	parkFactorRepo := repository.NewParkFactorRepository(db)

	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	tokenRepo := repository.NewOAuthTokenRepository(db)

	echo.Info("Registering routes...")

	return newServer(
		NewPlayerRoutes(playerRepo, awardRepo),
		NewTeamRoutes(teamRepo, gameRepo),
		NewStatsRoutes(statsRepo),
		NewAwardRoutes(awardRepo),
		NewGameRoutes(gameRepo, playRepo),
		NewPlayRoutes(playRepo, playerRepo),
		NewPitchRoutes(pitchRepo),
		NewMetaRoutes(metaRepo),
		NewManagerRoutes(managerRepo),
		NewParkRoutes(parkRepo),
		NewUmpireRoutes(umpireRepo),
		NewPostseasonRoutes(postseasonRepo, gameRepo),
		NewAllStarRoutes(awardRepo),
		NewSearchRoutes(playerRepo, teamRepo, parkRepo, gameRepo),
		NewEjectionRoutes(ejectionRepo),
		NewDerivedRoutes(derivedRepo),
		NewComputedRoutes(advancedStatsRepo, leverageRepo, parkFactorRepo),
		NewAuthRoutes(userRepo, tokenRepo, apiKeyRepo),
		NewUIRoutes(apiKeyRepo),
		NewMLBRoutes(cacheClient),
	)
}

// NewServer wires all registrars into one mux.
func newServer(registrars ...Registrar) *Server {
	docs.SwaggerInfo.BasePath = "/v1"

	mux := http.NewServeMux()

	for _, r := range registrars {
		r.RegisterRoutes(mux)
	}

	// Health check endpoint
	// @Summary Health check
	// @Description Check if the API server is running
	// @Tags health
	// @Accept json
	// @Produce json
	// @Success 200 {object} HealthResponse
	// @Router /health [get]
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
	})

	mux.HandleFunc("/docs/", httpSwagger.WrapHandler)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})

	return &Server{mux: mux}
}

// Implement http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
