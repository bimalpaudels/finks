package cli

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "finks",
	Short: "🦜 Finks - Lightweight Self-Hosting PaaS Tool",
	Long: `Finks is a CLI tool for managing your self-hosted server infrastructure.
It provides commands to monitor, deploy, and manage applications with ease.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
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
