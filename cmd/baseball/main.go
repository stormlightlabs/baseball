package main

import (
	"fmt"
	"os"

	cli "stormlightlabs.org/baseball/internal/cli"
)

func main() {
	if err := cli.NewBaseballRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
