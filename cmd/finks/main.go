package main

import (
	"os"

	"github.com/bimalpaudels/finks/internal/cli"
	"github.com/bimalpaudels/finks/internal/installer"
)

func main() {
	// When run with no arguments (e.g. from install script), run the installation wizard.
	if len(os.Args) == 1 {
		if err := installer.Run(); err != nil {
			os.Exit(1)
		}
		return
	}
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
