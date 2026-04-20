package main

import (
	"fmt"
	"os"

	"stormlightlabs.org/baseball/commands"
)

func main() {
	if err := commands.NewBaseballRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
