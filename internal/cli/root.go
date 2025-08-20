package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "finks",
	Short: "🦜 Finks - Lightweight Self-Hosting PaaS Tool",
	Long: `Finks is a CLI tool for managing your self-hosted server infrastructure.
It provides commands to monitor, deploy, and manage applications with ease.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🦜 Finks - Lightweight Self-Hosting PaaS Tool")
		fmt.Println("")
		fmt.Println("Available Commands:")
		fmt.Println("  app        Application deployment and management")
		fmt.Println("  server     Server management commands")
		fmt.Println("")
		fmt.Println("Use 'finks [command] --help' for more information about a command.")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(appCmd, serverCmd)
}
