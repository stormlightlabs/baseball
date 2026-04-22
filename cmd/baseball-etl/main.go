package main

import (
	"fmt"
	"os"

	etlcli "stormlightlabs.org/baseball/internal/cli/etl"
)

func main() {
	if err := etlcli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
