package cli

import (
	"fmt"
	"time"

	"github.com/bimalpaudels/finks/pkg/monitor"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	live     bool
	interval string
	all      bool
)

var metricsService *monitor.MetricsService

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server management commands",
	Long:  `Commands for managing and monitoring your server.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		metricsService = monitor.NewMetricsService()
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// metricsCmd represents the metrics command
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "View server metrics",
	Long:  `Display server metrics with options for live streaming.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		duration, err := time.ParseDuration(interval)
		if err != nil {
			pterm.Warning.Printf("Invalid interval format, using default 2s: %v\n", err)
			duration = 2 * time.Second
		}

		if !live {
			return fetchAndDisplayMetrics()
		}

		pterm.Info.Println("📊 Live streaming server metrics (Press Ctrl+C to stop)...")
		pterm.Println()

		// Enable screen clearing for live mode
		monitor.SetClearScreenMode(true)
		defer monitor.SetClearScreenMode(false)

		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		for range ticker.C {
			if err := fetchAndDisplayMetrics(); err != nil {
				pterm.Error.Printf("Error retrieving metrics: %v\n", err)
			}
		}

		return nil
	},
}

func fetchAndDisplayMetrics() error {
	var metrics *monitor.ServerMetrics
	var err error

	if all {
		metrics, err = metricsService.GetMetrics()
	} else {
		metrics, err = metricsService.GetSimpleMetrics()
	}

	if err != nil {
		return fmt.Errorf("error retrieving metrics: %w", err)
	}

	if all {
		monitor.DisplayMetrics(metrics)
	} else {
		monitor.DisplaySimpleMetrics(metrics)
	}
	return nil
}

func init() {
	// Add commands to server command
	serverCmd.AddCommand(metricsCmd)

	// Add flags for metrics command
	metricsCmd.Flags().BoolVarP(&live, "live", "l", false, "Stream metrics in real-time")
	metricsCmd.Flags().StringVarP(&interval, "interval", "i", "2s", "Update interval for live metrics (e.g., 1s, 5s)")
	metricsCmd.Flags().BoolVarP(&all, "all", "a", false, "Show comprehensive metrics including processes and detailed stats")
}
