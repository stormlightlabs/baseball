package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/echo"
)

func NewBaseballRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "baseball",
		Short: "Baseball API server, database, and cache toolkit",
		Long: fmt.Sprintf(`%v

A comprehensive toolkit for baseball API serving, database operations, and cache tooling.

Supports Lahman and Retrosheet data sources.
`, echo.HeaderStyle().Render("Baseball API")),
	}

	root.PersistentFlags().String("config", "conf.toml", "Path to config file")
	root.AddCommand(DbCmd())
	root.AddCommand(ServerCmd())
	root.AddCommand(CacheCmd())
	return root
}

func NewETLRootCmd() *cobra.Command {
	root := ETLCmd()
	root.Use = "baseball-etl"
	root.Short = "Baseball ETL toolkit"
	root.Long = "Extract, Transform, and Load operations for Lahman and Retrosheet data sources."
	root.PersistentFlags().String("config", "conf.toml", "Path to config file")
	return root
}
