package main

import (
	"fmt"
	"os"

	"stormlightlabs.org/baseball/commands"
)

func main() {
	if err := commands.NewETLRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
