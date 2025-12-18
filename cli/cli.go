package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"stormlightlabs.org/baseball/cmd"
	"stormlightlabs.org/baseball/internal/echo"
)

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "baseball",
		Short: "Baseball API ETL and Server toolkit",
		Long: fmt.Sprintf(`%v

A comprehensive toolkit for baseball data ETL and API serving.

Supports Lahman and Retrosheet data sources.
`, echo.HeaderStyle().Render("Baseball API")),
	}

	root.PersistentFlags().String("config", "conf.toml", "Path to config file")
	root.AddCommand(cmd.ETLCmd())
	root.AddCommand(cmd.DbCmd())
	root.AddCommand(cmd.ServerCmd())
	root.AddCommand(cmd.CacheCmd())
	root.AddCommand(cmd.DeployCmd())
	return root
}

// RootCmd is the root command for the baseball CLI
var RootCmd *cobra.Command

func init() { RootCmd = rootCmd() }
