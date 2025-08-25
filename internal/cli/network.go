package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage Docker networks",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available network commands:")
		fmt.Println("  create     Create a new Docker network")
		fmt.Println("  list       List all Docker networks")
		fmt.Println("  remove     Remove a Docker network")
		fmt.Println("  connect    Connect a container to a network")
		fmt.Println("  disconnect Disconnect a container from a network")
		fmt.Println("")
		fmt.Println("Use 'finks network [command] --help' for more information.")
	},
}

func init() {
	rootCmd.AddCommand(networkCmd)
}
