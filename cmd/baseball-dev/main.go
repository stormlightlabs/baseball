package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var root *cobra.Command

func init() {
	root = &cobra.Command{
		Use:          "baseball-dev",
		Short:        "Big Fly dev tools",
		SilenceUsage: true,
	}

	root.AddCommand(newGenerateCmd())
	root.AddCommand(newRunCmd())
	root.AddCommand(newCleanupCmd())
	root.AddCommand(newSwaggerCmd())
}

func main() {
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
