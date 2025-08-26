package cli

import (
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server management commands",
	Long:  `Commands for managing and monitoring your server.`,

	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
