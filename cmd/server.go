package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/api"
	"stormlightlabs.org/baseball/internal/config"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
	"stormlightlabs.org/baseball/internal/middleware"
)

// TODO: configurable baseURL
const baseURL string = "http://localhost:8080/v1/"

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
	cmd.AddCommand(ServerAuthCmd())
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

	cmd.Flags().Bool("debug", false, "Enable debug mode (disables authentication)")
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
	cmd.Flags().StringP("token", "t", "", "Bearer token for authentication")
	cmd.Flags().StringP("api-key", "k", "", "API key for authentication")
	return cmd
}

// ServerHealthCmd creates the health command
func ServerHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check server health",
		Long:  "Perform health check on the running API server.",
		RunE:  checkHealth,
	}
}

// ServerAuthCmd creates the auth command
func ServerAuthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Get API authentication instructions",
		Long:  "Display instructions for authenticating with the Baseball API.",
		RunE:  authInstructions,
	}
}

func fetchEndpoint(cmd *cobra.Command, args []string) error {
	path := args[0]
	format, _ := cmd.Flags().GetString("format")
	raw, _ := cmd.Flags().GetBool("raw")
	token, _ := cmd.Flags().GetString("token")
	apiKey, _ := cmd.Flags().GetString("api-key")

	url := baseURL + path

	if !raw {
		echo.Header("API Test")
		echo.Infof("Fetching: %s", url)
		echo.Info("")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error: failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
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

	serverURL := "http://localhost:8080/v1/health"
	echo.Infof("Checking: %s", serverURL)
	echo.Info("")

	resp, err := http.Get(serverURL)
	if err != nil {
		return fmt.Errorf("error: server is not running or unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		echo.Successf("✓ Server is healthy (Status: %s)", resp.Status)

		body, err := io.ReadAll(resp.Body)
		if err == nil && len(body) > 0 {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
				echo.Info("")
				echo.Info(prettyJSON.String())
			}
		}
		return nil
	}

	return fmt.Errorf("error: server returned status: %s", resp.Status)
}

func authInstructions(cmd *cobra.Command, args []string) error {
	echo.Header("API Authentication")
	echo.Info("")
	echo.Info("To access the Baseball API, you need an API key or session token.")
	echo.Info("")
	echo.Info("Step 1: Login to the Dashboard")
	echo.Info("  Visit: http://localhost:8080/dashboard")
	echo.Info("  Or login directly:")
	echo.Info("    • GitHub: http://localhost:8080/v1/auth/github")
	echo.Info("    • Codeberg: http://localhost:8080/v1/auth/codeberg")
	echo.Info("")
	echo.Info("Step 2: Generate an API Key")
	echo.Info("  After logging in, you can generate API keys from the dashboard.")
	echo.Info("  API keys allow programmatic access without logging in each time.")
	echo.Info("")
	echo.Info("Step 3: Use Your API Key")
	echo.Info("  You can use your API key in two ways:")
	echo.Info("")
	echo.Info("  A. With the CLI fetch command:")
	echo.Infof("     baseball server fetch 'players?name=ruth' --api-key 'sk_...'")
	echo.Info("")
	echo.Info("  B. With HTTP requests:")
	echo.Info("     curl -H 'Authorization: Bearer sk_...' http://localhost:8080/v1/players")
	echo.Info("")
	echo.Success("✓ For local development, start the server with --debug to disable authentication")
	echo.Infof("  baseball server start --debug")
	echo.Info("")
	return nil
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
		echo.Info("⚠ Debug mode enabled - authentication disabled")
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

	if _, err := redisClient.Ping(cmd.Context()).Result(); err != nil {
		echo.Infof("⚠ Redis connection failed: %v", err)
		echo.Info("  Rate limiting will be disabled")
		redisClient = nil
	} else {
		echo.Success("✓ Connected to Redis")
	}

	server := api.NewServer(database.DB)

	timeFmt := time.DateTime
	if cfg.Server.DebugMode {
		timeFmt = time.Kitchen
	}

	logger := log.NewWithOptions(cmd.OutOrStdout(), log.Options{
		ReportTimestamp: true,
		TimeFormat:      timeFmt,
		Prefix:          "⚾️",
		ReportCaller:    cfg.Server.DebugMode,
	})

	rateLimiter := middleware.NewRateLimiter(redisClient, cfg.Server.DebugMode, 60, time.Minute)

	var handler http.Handler = server
	bind := middleware.Logger(logger)
	handler = bind(handler)

	if !cfg.Server.DebugMode && redisClient != nil {
		handler = rateLimiter.Middleware(handler)
		echo.Info("✓ Rate limiting enabled (60 req/min per IP)")
	} else if cfg.Server.DebugMode {
		echo.Info("⚠ Rate limiting disabled (debug mode)")
	} else if redisClient == nil {
		echo.Info("⚠ Rate limiting disabled (Redis unavailable)")
	} else {
		echo.Info("⚠ Rate limiting disabled (debug mode or Redis unavailable)")
	}

	echo.Info("✓ Request logging enabled")

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	echo.Success(fmt.Sprintf("✓ Server starting on %s", addr))
	if !cfg.Server.DebugMode {
		echo.Info("✓ Authentication enabled")
		echo.Info("  GitHub OAuth: /v1/auth/github")
		echo.Info("  Codeberg OAuth: /v1/auth/codeberg")
		echo.Info("  Dashboard: /dashboard")
	}
	echo.Info("Press Ctrl+C to stop")
	echo.Info("")
	return http.ListenAndServe(addr, handler)
}
