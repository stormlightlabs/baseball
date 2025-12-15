// Package api provides HTTP handlers for the Baseball API
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
//
// @tag.name meta
// @tag.description Metadata about the API
//
// @tag.name computed
// @tag.description Computed statistics
//
// @tag.name derived
// @tag.description Derived statistics
//
// @tag.name leverage
// @tag.description Leverage index data
//
// @tag.name win-expectancy
// @tag.description Win expectancy data
//
// @tag.name leaders
// @tag.description Leaderboard data
//
// @tag.name war
// @tag.description Computed WAR
//
// @tag.name mlb
// @tag.description MLB Stats API endpoint proxies
//
// @tag.name pitching
// @tag.description MLB pitching statistics
//
// @tag.name search
// @tag.description Searchable data
//
// @tag.name awards
// @tag.description MLB awards data
//
// @tag.name players
// @tag.description Player career data
//
// @tag.name teams
// @tag.description Team/franchise data
//
// @tag.name stats
// @tag.description MLB game statistics
//
// @tag.name pitching
// @tag.description MLB pitching statistics
//
// @tag.name batting
// @tag.description MLB batting statistics
//
// @tag.name parks
// @tag.description MLB park data
//
// @tag.name managers
// @tag.description MLB manager data
//
// @tag.name umpires
// @tag.description MLB umpire data
//
// @tag.name seasons
// @tag.description MLB season data
//
// @tag.name postseason
// @tag.description MLB postseason data
//
// @tag.name ejections
// @tag.description MLB ejection data
//
// @tag.name win-expectancy
// @tag.description Computed win expectancy data
//
// @tag.name federalleague
// @tag.description Federal League (1914-1915) data
package api

import (
	"database/sql"
	_ "expvar"
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
	weRepo := repository.NewWinExpectancyRepository(db)

	advancedStatsRepo := repository.NewAdvancedStatsRepository(db)
	leverageRepo := repository.NewLeverageRepository(db)
	parkFactorRepo := repository.NewParkFactorRepository(db)

	negroLeaguesRepo := repository.NewNegroLeaguesRepository(db, cacheClient)

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
		NewDerivedRoutes(derivedRepo, weRepo),
		NewComputedRoutes(advancedStatsRepo, leverageRepo, parkFactorRepo),
		NewAuthRoutes(userRepo, tokenRepo, apiKeyRepo),
		NewUIRoutes(apiKeyRepo),
		// TODO: rename to mlbstatsapiroutes
		NewMLBRoutes(cacheClient),
		NewFederalLeagueRoutes(gameRepo, playRepo, teamRepo),
		NewNegroLeaguesRoutes(negroLeaguesRepo),
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

	mux.Handle("GET /debug/vars", http.DefaultServeMux)
	return &Server{mux: mux}
}

// Implement http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
