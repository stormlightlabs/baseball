package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/echo"
)

func NewBaseballRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "baseball",
		Short: "Baseball API ETL and Server toolkit",
		Long: fmt.Sprintf(`%v

A comprehensive toolkit for baseball data ETL and API serving.

Supports Lahman and Retrosheet data sources.
`, echo.HeaderStyle().Render("Baseball API")),
	}

	root.PersistentFlags().String("config", "conf.toml", "Path to config file")
	root.AddCommand(ETLCmd())
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
