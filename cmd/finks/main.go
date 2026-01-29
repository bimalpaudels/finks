package main

import (
	"fmt"
	"os"

	"github.com/bimalpaudels/finks/internal/cli"
	"github.com/bimalpaudels/finks/internal/installer"
)

var version = "0.1.0"

func main() {
	// If no arguments provided, run installation wizard
	if len(os.Args) == 1 {
		if err := installer.Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Otherwise, run CLI commands
	if err := cli.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
