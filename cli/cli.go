// TODO: refactor [RootCmd] to be a func
package main

import (
	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/cmd"
	"stormlightlabs.org/baseball/internal/echo"
)

// RootCmd is the root command for the baseball CLI
var RootCmd = &cobra.Command{
	Use:   "baseball",
	Short: "Baseball API ETL and Server toolkit",
	Long: echo.HeaderStyle().Render("Baseball API") + "\n\n" +
		"A comprehensive toolkit for baseball data ETL and API serving.\n" +
		"Supports Lahman and Retrosheet data sources.",
}

func init() {
	RootCmd.AddCommand(cmd.ETLCmd())
	RootCmd.AddCommand(cmd.DbCmd())
	RootCmd.AddCommand(cmd.ServerCmd())
}
