package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
	"stormlightlabs.org/baseball/internal/seed"
)

type deployOptions struct {
	registry   string
	tag        string
	push       bool
	skipBuild  bool
	skipETL    bool
	dryRun     bool
	lahmanDir  string
	retroYears string
	retroDir   string
}

// DeployCmd creates the deploy command group
func DeployCmd() *cobra.Command {
	var (
		registry   string
		tag        string
		push       bool
		skipBuild  bool
		skipETL    bool
		dryRun     bool
		lahmanDir  string
		retroYears string
		retroDir   string
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy the Baseball API",
		Long: `Build Docker image, push to registry, and prepare for deployment.

This command helps with the deployment workflow:
  1. Build Docker image (unless --skip-build)
  2. Tag with version and latest
  3. Push to Docker registry (if --push)
  4. Run migrations (if database URL provided)
  5. Run ETL pipeline with idempotency checks (unless --skip-etl)

Examples:
  # Dry run - see what would happen
  baseball deploy --tag v1.0.0 --registry username --push --dry-run

  # Build and tag image
  baseball deploy --tag v1.0.0

  # Build, tag, and push to DockerHub
  baseball deploy --tag v1.0.0 --registry username --push

  # Build and prepare data for deployment
  baseball deploy --tag v1.0.0 --years 2020-2025

  # Just push existing image (skip build)
  baseball deploy --tag v1.0.0 --registry username --push --skip-build`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(cmd, deployOptions{
				registry:   registry,
				tag:        tag,
				push:       push,
				skipBuild:  skipBuild,
				skipETL:    skipETL,
				dryRun:     dryRun,
				lahmanDir:  lahmanDir,
				retroYears: retroYears,
				retroDir:   retroDir,
			})
		},
	}

	cmd.Flags().StringVar(&registry, "registry", "", "Docker registry/username (e.g., 'username' for DockerHub)")
	cmd.Flags().StringVar(&tag, "tag", "latest", "Image tag/version (e.g., 'v1.0.0' or 'latest')")
	cmd.Flags().BoolVar(&push, "push", false, "Push image to Docker registry")
	cmd.Flags().BoolVar(&skipBuild, "skip-build", false, "Skip building the image")
	cmd.Flags().BoolVar(&skipETL, "skip-etl", false, "Skip running ETL pipeline")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without running commands")
	cmd.Flags().StringVar(&lahmanDir, "lahman-dir", "", "Path to Lahman CSV directory")
	cmd.Flags().StringVar(&retroYears, "years", "", "Comma-separated years or ranges for Retrosheet (e.g., '2020,2023-2025')")
	cmd.Flags().StringVar(&retroDir, "retro-dir", "", "Path to Retrosheet data directory")
	return cmd
}

func runDeploy(cmd *cobra.Command, opts deployOptions) error {
	if opts.dryRun {
		echo.Header("Baseball API Deployment (DRY RUN)")
		echo.Info("Dry run mode - no commands will be executed")
		echo.Info("")
	} else {
		echo.Header("Baseball API Deployment")
	}

	imageName := "baseball-app"
	fullImageName := imageName

	if opts.registry != "" {
		fullImageName = opts.registry + "/" + imageName
	}

	taggedImage := fullImageName + ":" + opts.tag
	latestImage := fullImageName + ":latest"

	if !opts.skipBuild {
		echo.Info("Building Docker image...")
		echo.Infof("  Image: %s", taggedImage)

		buildCmd := fmt.Sprintf("docker build -t %s -t %s .", taggedImage, latestImage)
		echo.Infof("  Would run: %s", buildCmd)

		if !opts.dryRun {
			if err := runShellCommand(buildCmd); err != nil {
				return fmt.Errorf("error: failed to build Docker image: %w", err)
			}
			echo.Success("✓ Docker image built successfully")
		}
	} else {
		echo.Info("Skipping Docker build (--skip-build)")
	}

	if opts.push {
		if opts.registry == "" {
			return fmt.Errorf("error: --registry is required when using --push")
		}

		echo.Info("")
		echo.Info("Pushing images to Docker registry...")
		echo.Infof("  Registry: %s", opts.registry)

		if !opts.dryRun {
			echo.Info("  Checking Docker authentication...")
			checkCmd := "docker info --format '{{.RegistryConfig.IndexConfigs}}'"
			if err := runShellCommand(checkCmd); err != nil {
				echo.Info("")
				echo.Info("⚠ Docker authentication check failed")
				echo.Info("")
				echo.Info("To authenticate with DockerHub, run:")
				echo.Infof("  docker login -u %s", opts.registry)
				echo.Info("")
				echo.Info("Or set up credentials:")
				echo.Info("  export DOCKER_USERNAME=your_username")
				echo.Info("  export DOCKER_PASSWORD=your_password")
				echo.Info("  echo $DOCKER_PASSWORD | docker login -u $DOCKER_USERNAME --password-stdin")
				echo.Info("")
				return fmt.Errorf("error: Docker authentication required")
			}
			echo.Info("  ✓ Docker authenticated")
		}

		for _, img := range []string{taggedImage, latestImage} {
			pushCmd := fmt.Sprintf("docker push %s", img)
			echo.Infof("  Would run: %s", pushCmd)

			if !opts.dryRun {
				if err := runShellCommand(pushCmd); err != nil {
					echo.Info("")
					echo.Info("Push failed. Make sure you're logged in:")
					echo.Infof("  docker login -u %s", opts.registry)
					return fmt.Errorf("error: failed to push %s: %w", img, err)
				}
			}
		}

		if !opts.dryRun {
			echo.Success("✓ Images pushed successfully")
		}
	}

	if !opts.skipETL {
		echo.Info("")
		echo.Info("Running ETL pipeline with idempotency checks...")

		if opts.dryRun {
			var refreshes map[string]db.DatasetRefresh
			ctx := cmd.Context()
			database, err := db.Connect("")
			if err != nil {
				echo.Info("  Could not connect to database to check current state")
				echo.Infof("    Error: %v", err)
				echo.Info("  Would attempt to load data if database was available")
				refreshes = make(map[string]db.DatasetRefresh)
			} else {
				defer database.Close()

				refreshes, err = database.DatasetRefreshes(ctx)
				if err != nil {
					echo.Infof("  Could not query dataset state: %v", err)
					refreshes = make(map[string]db.DatasetRefresh)
				} else {
					echo.Info("")
					echo.Info("Current database state:")

					// Check Lahman
					if lahmanRefresh, exists := refreshes["lahman"]; exists {
						echo.Infof("  ✓ Lahman: Already loaded (%s rows, last updated %s)",
							formatLargeNumber(lahmanRefresh.RowCount),
							lahmanRefresh.LastLoadedAt.Format("2006-01-02"))
					} else {
						echo.Info("  ✗ Lahman: Not loaded")
					}

					gamesExists := false
					playsExists := false
					if _, exists := refreshes["retrosheet_games"]; exists {
						gamesExists = true
					}
					if _, exists := refreshes["retrosheet_plays"]; exists {
						playsExists = true
					}

					if gamesExists || playsExists {
						echo.Info("  ✓ Retrosheet:")

						if gamesExists {
							rows, err := database.QueryContext(ctx, `
								SELECT DISTINCT SUBSTRING(game_id FROM 4 FOR 4)::int as year
								FROM games
								ORDER BY year
							`)
							if err == nil {
								defer rows.Close()
								var gameYears []int
								for rows.Next() {
									var year int
									if err := rows.Scan(&year); err == nil {
										gameYears = append(gameYears, year)
									}
								}
								if len(gameYears) > 0 {
									echo.Infof("      Games: %s", formatYearRangeWithGaps(gameYears))
								}
							}
						}

						if playsExists {
							rows, err := database.QueryContext(ctx, `
								SELECT tablename
								FROM pg_tables
								WHERE schemaname = 'public'
									AND tablename LIKE 'plays_%'
								ORDER BY tablename
							`)
							if err == nil {
								defer rows.Close()
								var playYears []int
								for rows.Next() {
									var tableName string
									if err := rows.Scan(&tableName); err == nil {
										var year int
										if _, err := fmt.Sscanf(tableName, "plays_%d", &year); err == nil {
											playYears = append(playYears, year)
										}
									}
								}
								if len(playYears) > 0 {
									sort.Ints(playYears)
									echo.Infof("      Plays:  %s", formatYearRangeWithGaps(playYears))
								}
							}
						}
					} else {
						echo.Info("  ✗ Retrosheet: No data loaded")
					}
					echo.Info("")
				}
			}

			lahmanDir := opts.lahmanDir
			if lahmanDir == "" {
				lahmanDir = filepath.Join("data", "lahman", "csv")
			}

			if _, err := os.Stat(lahmanDir); err == nil {
				echo.Infof("Would load Lahman data from: %s", lahmanDir)
				if _, exists := refreshes["lahman"]; exists {
					echo.Info("  → Would skip (already loaded)")
				} else {
					echo.Info("  → Would load all Lahman tables")
				}
			} else {
				echo.Info("Would skip Lahman (data directory not found)")
			}

			echo.Info("")
			retroDir := opts.retroDir
			if retroDir == "" {
				retroDir = filepath.Join("data", "retrosheet")
			}

			if _, err := os.Stat(retroDir); err == nil {
				var requestedYears []int
				if opts.retroYears != "" {
					requestedYears, err = parseYearFlag(opts.retroYears)
					if err != nil {
						echo.Infof("  Invalid years format: %v", err)
					}
				}

				echo.Infof("Would load Retrosheet data from: %s", retroDir)
				if opts.retroYears != "" {
					echo.Infof("  Requested years: %s", opts.retroYears)
				} else {
					echo.Info("  Requested years: (default range)")
				}

				if len(requestedYears) > 0 {
					if _, exists := refreshes["retrosheet_games"]; exists {
						rows, err := database.QueryContext(ctx, `
							SELECT DISTINCT SUBSTRING(game_id FROM 4 FOR 4)::int as year
							FROM games
							ORDER BY year
						`)
						if err == nil {
							defer rows.Close()
							loadedYears := make(map[int]bool)
							for rows.Next() {
								var year int
								if err := rows.Scan(&year); err == nil {
									loadedYears[year] = true
								}
							}

							newYears := make([]int, 0)
							skipYears := make([]int, 0)
							for _, year := range requestedYears {
								if loadedYears[year] {
									skipYears = append(skipYears, year)
								} else {
									newYears = append(newYears, year)
								}
							}

							if len(newYears) > 0 {
								echo.Infof("  → Would load %d new years: %s", len(newYears), formatYearRange(newYears))
							}
							if len(skipYears) > 0 {
								echo.Infof("  → Would skip %d existing years: %s", len(skipYears), formatYearRange(skipYears))
							}
							if len(newYears) == 0 && len(skipYears) == 0 {
								echo.Info("  → No matching years found")
							}
						} else {
							echo.Info("  → only loads missing years")
						}
					} else {
						echo.Infof("  → Would load %d new years: %s", len(requestedYears), formatYearRange(requestedYears))
					}
				} else {
					echo.Info("  → only loads missing years")
				}
			} else {
				echo.Info("Would skip Retrosheet (data directory not found)")
			}
		} else {
			database, err := db.Connect("")
			if err != nil {
				return fmt.Errorf("error: failed to connect to database: %w", err)
			}
			defer database.Close()

			ctx := cmd.Context()

			lahmanDir := opts.lahmanDir
			if lahmanDir == "" {
				lahmanDir = filepath.Join("data", "lahman", "csv")
			}

			if _, err := os.Stat(lahmanDir); err == nil {
				echo.Info("Loading Lahman data...")
				_, err := seed.LoadLahman(ctx, database, seed.LahmanOptions{
					CSVDir: lahmanDir,
					Skip:   true,
				})
				if err != nil {
					return fmt.Errorf("error: failed to load Lahman data: %w", err)
				}
			} else {
				echo.Info("Skipping Lahman (data directory not found)")
			}

			retroDir := opts.retroDir
			if retroDir == "" {
				retroDir = filepath.Join("data", "retrosheet")
			}

			if _, err := os.Stat(retroDir); err == nil {
				echo.Info("Loading Retrosheet data...")

				var years []int
				if opts.retroYears != "" {
					years, err = parseYearFlag(opts.retroYears)
					if err != nil {
						return fmt.Errorf("error: invalid years flag: %w", err)
					}
				}

				_, err := seed.LoadRetrosheet(ctx, database, seed.RetrosheetOptions{
					DataDir: retroDir,
					Years:   years,
					Force:   false,
				})
				if err != nil {
					return fmt.Errorf("error: failed to load Retrosheet data: %w", err)
				}
			} else {
				echo.Info("Skipping Retrosheet (data directory not found)")
			}

			echo.Success("✓ ETL pipeline completed")
		}
	}

	echo.Info("")
	if opts.dryRun {
		echo.Success("✓ Dry run complete - no changes made")
		echo.Info("")
		echo.Info("To execute for real, remove the --dry-run flag")
	} else {
		echo.Success("✓ Deployment preparation complete")
		echo.Info("")
		echo.Info("Next steps for production deployment:")
		echo.Info("  1. Transfer data files to production server (if needed)")
		echo.Info("  2. Update docker-compose.yml on server")
		echo.Info("  3. Set environment variables (DATABASE_URL, REDIS_URL)")
		echo.Info("  4. Run: docker-compose pull && docker-compose up -d")
		echo.Info("  5. Run migrations: docker-compose exec app baseball db migrate")
		echo.Info("  6. Monitor logs: docker-compose logs -f app")
	}

	return nil
}

// runShellCommand executes a shell command and streams output to stdout/stderr.
func runShellCommand(cmdStr string) error {
	shell := "/bin/sh"
	if runtime := os.Getenv("SHELL"); runtime != "" {
		shell = runtime
	}

	cmd := exec.Command(shell, "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
