package main

import (
	"fmt"
	"os"

	servercli "stormlightlabs.org/baseball/internal/cli/server"
)

func main() {
	if err := servercli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
