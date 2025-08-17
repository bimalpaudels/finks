package main

import (
	"fmt"
	"os"

	"github.com/bimalpaudels/finks/internal/cli"
)

const version = "0.1.0"

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}