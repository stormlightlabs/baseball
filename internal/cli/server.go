package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	charmlog "github.com/charmbracelet/log"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/api"
	"stormlightlabs.org/baseball/internal/cache"
	"stormlightlabs.org/baseball/internal/config"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
	"stormlightlabs.org/baseball/internal/middleware"
)

type Route struct {
	Method string
	Path   string
}

// ServerCmd creates the server command group
func ServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Server operations",
		Long:  "Start and manage the baseball API server.",
	}

	cmd.AddCommand(ServerStartCmd())
	cmd.AddCommand(ServerFetchCmd())
	cmd.AddCommand(ServerHealthCmd())
	cmd.AddCommand(ServerRoutesCmd())
	return cmd
}

// ServerStartCmd creates the start command
func ServerStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the API server",
		Long:  "Start the baseball API HTTP server.",
		RunE:  startServer,
	}

	cmd.Flags().Bool("debug", false, "Enable debug mode (disables rate limiting)")
	return cmd
}

// ServerFetchCmd creates the server fetch command
func ServerFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch [path]",
		Short: "Test API endpoints",
		Long: `cURL-like tool for testing API endpoints with formatted output.

Path should be relative to /v1/ (e.g., 'players?name=ruth' or 'teams/BOS?year=2023').`,
		Args: cobra.ExactArgs(1),
		RunE: fetchEndpoint,
	}

	cmd.Flags().StringP("format", "f", "json", "Output format (json|table)")
	cmd.Flags().BoolP("raw", "r", false, "Output raw JSON without colors or formatting (suitable for piping to jq)")
	return cmd
}

// ServerHealthCmd creates the health command
func ServerHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check server health",
		Long:  "Perform health check on the running API server.",
		RunE:  checkHealth,
	}
	cmd.Flags().Bool("ready", false, "Check readiness at /v1/ready instead of liveness at /v1/health")
	return cmd
}

// ServerRoutesCmd creates the routes command
func ServerRoutesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "routes",
		Short: "List all registered API routes",
		Long:  "Display all registered API routes in a Rails rake routes style format.",
		RunE:  showRoutes,
	}
}

func fetchEndpoint(cmd *cobra.Command, args []string) error {
	path := args[0]
	format, _ := cmd.Flags().GetString("format")
	raw, _ := cmd.Flags().GetBool("raw")

	configPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("error: failed to load config: %w", err)
	}

	url := cfg.Server.BaseURL + path

	if !raw {
		echo.Header("API Test")
		echo.Infof("Fetching: %s", url)
		echo.Info("")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error: failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer resp.Body.Close()

	if !raw {
		echo.Infof("Status: %s", resp.Status)
		echo.Info("")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error: failed to read response: %w", err)
	}

	if raw {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
			fmt.Println(string(body))
		} else {
			fmt.Println(prettyJSON.String())
		}
		return nil
	}

	if format == "table" {
		echo.Info("Table format not yet implemented, showing JSON:")
		echo.Info("")
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		echo.Info(string(body))
	} else {
		echo.Info(prettyJSON.String())
	}

	echo.Info("")
	echo.Successf("✓ Request completed (%d bytes)", len(body))
	return nil
}

func checkHealth(cmd *cobra.Command, args []string) error {
	echo.Header("Health Check")

	readyCheck, _ := cmd.Flags().GetBool("ready")
	path := "/v1/health"
	if readyCheck {
		path = "/v1/ready"
	}
	serverURL := "http://localhost:8080" + path
	echo.Infof("Checking: %s", serverURL)
	echo.Info("")

	resp, err := http.Get(serverURL)
	if err != nil {
		return fmt.Errorf("error: server is not running or unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error: failed to read response: %w", err)
	}

	if len(body) > 0 {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
			echo.Info(prettyJSON.String())
			echo.Info("")
		}
	}

	if resp.StatusCode == http.StatusOK {
		if readyCheck {
			echo.Successf("✓ Server is ready (Status: %s)", resp.Status)
		} else {
			echo.Successf("✓ Server is healthy (Status: %s)", resp.Status)
		}
		return nil
	}

	if readyCheck && resp.StatusCode == http.StatusServiceUnavailable {
		return fmt.Errorf("error: server is live but not ready: %s", resp.Status)
	}

	return fmt.Errorf("error: server returned status: %s", resp.Status)
}

func startServer(cmd *cobra.Command, args []string) error {
	echo.Header("Starting Server")
	echo.Info("Loading configuration...")

	configPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("error: failed to load config: %w", err)
	}

	debugMode, _ := cmd.Flags().GetBool("debug")
	if debugMode {
		cfg.Server.DebugMode = true
	}

	if cfg.Server.DebugMode {
		echo.Info("⚠ Debug mode enabled - rate limiting disabled")
	}

	echo.Info("Connecting to database...")
	database, err := db.Connect(cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	defer database.Close()

	echo.Success("✓ Connected to database")
	echo.Info("Connecting to Redis...")

	redisOpts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return fmt.Errorf("error: failed to parse Redis URL: %w", err)
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	var cacheClient *cache.Client
	if _, err := redisClient.Ping(cmd.Context()).Result(); err != nil {
		echo.Infof("⚠ Redis connection failed: %v", err)
		echo.Info("  Rate limiting and caching will be disabled")
		redisClient = nil
	} else {
		echo.Success("✓ Connected to Redis")

		env := "dev"
		if !cfg.Server.DebugMode {
			env = "prod"
		}

		cacheConfig := cache.Config{
			App:     "baseball",
			Env:     env,
			Version: cfg.Cache.Version,
			Enabled: cfg.Cache.Enabled,
			TTLs: cache.TTLConfig{
				Entity:   time.Duration(cfg.Cache.TTLs.Entity) * time.Second,
				List:     time.Duration(cfg.Cache.TTLs.List) * time.Second,
				Search:   time.Duration(cfg.Cache.TTLs.Search) * time.Second,
				Upstream: time.Duration(cfg.Cache.TTLs.Upstream) * time.Second,
				Negative: time.Duration(cfg.Cache.TTLs.Negative) * time.Second,
			},
		}

		cacheClient = cache.NewClient(redisClient, cacheConfig)
		if cfg.Cache.Enabled {
			echo.Success("✓ Cache enabled")
		} else {
			echo.Info("⚠ Cache disabled (set CACHE_ENABLED=true to enable)")
		}
	}

	server := api.NewServer(database.DB, cacheClient)

	timeFmt := time.DateTime
	if cfg.Server.DebugMode {
		timeFmt = time.Kitchen
	}

	var logger *slog.Logger
	if cfg.Server.DebugMode {
		devHandler := charmlog.NewWithOptions(cmd.OutOrStdout(), charmlog.Options{
			ReportTimestamp: true,
			TimeFormat:      timeFmt,
			Prefix:          "⚾",
			ReportCaller:    true,
		})
		logger = slog.New(devHandler).With("env", "dev")
	} else {
		logger = slog.New(slog.NewJSONHandler(cmd.OutOrStdout(), &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})).With("env", "prod")
	}
	slog.SetDefault(logger)

	rateLimiter := middleware.NewRateLimiter(
		redisClient,
		cfg.Server.DebugMode,
		cfg.Server.RateLimit.PublicPerMinute,
		time.Minute,
		cfg.Server.CORS.AllowedOrigins,
	)

	var handler http.Handler = server
	handler = middleware.Logger(logger)(handler)
	handler = middleware.MetricsMiddleware(nil)(handler)
	handler = middleware.TraceMiddleware(handler)

	if !cfg.Server.DebugMode && redisClient != nil {
		handler = rateLimiter.Middleware(handler)
		echo.Info("✓ Rate limiting enabled")
		echo.Infof("  Public: %d req/min per IP", cfg.Server.RateLimit.PublicPerMinute)
		echo.Info("  First-party bypass: X-BigFly-Client=web|mobile")
	} else if cfg.Server.DebugMode {
		echo.Info("⚠ Rate limiting disabled (debug mode)")
	} else if redisClient == nil {
		echo.Info("⚠ Rate limiting disabled (Redis unavailable)")
	} else {
		echo.Info("⚠ Rate limiting disabled (debug mode or Redis unavailable)")
	}

	handler = middleware.CORS(cfg.Server.CORS.AllowedOrigins)(handler)

	echo.Info("✓ Request logging enabled")
	echo.Info("✓ Metrics tracking enabled (/debug/vars)")
	echo.Info("✓ Request tracing enabled (X-Trace-ID)")
	if len(cfg.Server.CORS.AllowedOrigins) > 0 {
		echo.Infof("✓ CORS enabled (%d allowed origin(s))", len(cfg.Server.CORS.AllowedOrigins))
	} else {
		echo.Info("⚠ CORS disabled (no allowed origins configured)")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	echo.Info(fmt.Sprintf("ℹ Starting server on %s...", addr))
	echo.Info("")

	errChan := make(chan error, 1)
	go func() {
		errChan <- http.ListenAndServe(addr, handler)
	}()

	time.Sleep(100 * time.Millisecond)

	select {
	case err := <-errChan:
		return err
	default:
		// Server started successfully
	}

	echo.Success(fmt.Sprintf("✓ Server started on %s", addr))
	echo.Info("")
	echo.Info("Press Ctrl+C to stop")
	echo.Info("")

	return <-errChan
}

func showRoutes(cmd *cobra.Command, args []string) error {
	echo.Header("API Routes")
	echo.Info("Scanning internal/api directory...")

	routes, err := extractRoutesFromAST("internal/api")
	if err != nil {
		return fmt.Errorf("error: failed to extract routes: %w", err)
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Method != routes[j].Method {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	echo.Info("")
	echo.Success(fmt.Sprintf("✓ Found %d routes", len(routes)))
	echo.Info("")

	printRoutesTable(routes)

	echo.Info("")
	return nil
}

func extractRoutesFromAST(dir string) ([]Route, error) {
	var routes []Route

	matches, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	for _, match := range matches {
		file, err := parser.ParseFile(fset, match, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			if selExpr.Sel.Name != "HandleFunc" {
				return true
			}

			if len(callExpr.Args) < 2 {
				return true
			}

			patternLit, ok := callExpr.Args[0].(*ast.BasicLit)
			if !ok || patternLit.Kind != token.STRING {
				return true
			}

			pattern := strings.Trim(patternLit.Value, "\"")
			method, path := parsePattern(pattern)
			if path != "" {
				routes = append(routes, Route{Method: method, Path: path})
			}

			return true
		})
	}

	routes = append(routes, Route{Method: "GET", Path: "/v1/health"})
	routes = append(routes, Route{Method: "GET", Path: "/v1/docs/"})
	routes = append(routes, Route{Method: "GET", Path: "/debug/vars"})

	return routes, nil
}

func printRoutesTable(routes []Route) {
	if len(routes) == 0 {
		echo.Info("No routes found")
		return
	}

	maxMethodLen := 6
	maxPathLen := 4
	for _, r := range routes {
		if len(r.Method) > maxMethodLen {
			maxMethodLen = len(r.Method)
		}
		if len(r.Path) > maxPathLen {
			maxPathLen = len(r.Path)
		}
	}

	headerMethod := padRight("METHOD", maxMethodLen)
	headerPath := padRight("PATH", maxPathLen)
	echo.Info(fmt.Sprintf("%s  %s", headerMethod, headerPath))
	echo.Info(strings.Repeat("-", maxMethodLen+2+maxPathLen))

	for _, r := range routes {
		method := padRight(r.Method, maxMethodLen)
		echo.Info(fmt.Sprintf("%s  %s", method, r.Path))
	}
}
