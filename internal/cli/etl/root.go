package etl

import (
	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/internal/cli/shared"
)

func NewRootCmd() *cobra.Command {
	root := ETLCmd()
	root.Use = "baseball-etl"
	root.Short = "Baseball ETL toolkit"
	root.Long = "Extract, Transform, and Load operations for Lahman and Retrosheet data sources."
	root.PersistentFlags().String("config", "conf.toml", "Path to config file")
	root.AddCommand(shared.DbCmd())
	return root
}
