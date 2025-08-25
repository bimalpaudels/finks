package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/bimalpaudels/finks/pkg/monitor"
	"github.com/spf13/cobra"
)

var (
	live     bool
	interval string
	all      bool
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

// metricsCmd represents the metrics command
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "View server metrics",
	Long:  `Display server metrics with options for live streaming.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create metrics service
		metricsService := monitor.NewMetricsService()

		if live {
			fmt.Println("📊 Live streaming server metrics (Press Ctrl+C to stop)...")
			fmt.Println("")

			// Enable screen clearing for live mode
			monitor.SetClearScreenMode(true)
			defer monitor.SetClearScreenMode(false)

			// Parse interval
			duration, err := time.ParseDuration(interval)
			if err != nil {
				duration = 2 * time.Second
			}

			ticker := time.NewTicker(duration)
			defer ticker.Stop()

			for range ticker.C {
				var metrics *monitor.ServerMetrics
				var err error

				if all {
					metrics, err = metricsService.GetMetrics()
				} else {
					metrics, err = metricsService.GetSimpleMetrics()
				}

				if err != nil {
					fmt.Printf("Error retrieving metrics: %v\n", err)
					continue
				}

				if all {
					monitor.DisplayMetrics(metrics)
				} else {
					monitor.DisplaySimpleMetrics(metrics)
				}
			}
		} else {
			var metrics *monitor.ServerMetrics
			var err error

			if all {
				metrics, err = metricsService.GetMetrics()
			} else {
				metrics, err = metricsService.GetSimpleMetrics()
			}

			if err != nil {
				fmt.Printf("Error retrieving metrics: %v\n", err)
				os.Exit(1)
			}

			if all {
				monitor.DisplayMetrics(metrics)
			} else {
				monitor.DisplaySimpleMetrics(metrics)
			}
		}
	},
}

func init() {
	// Add commands to server command
	serverCmd.AddCommand(metricsCmd)

	// Add flags for metrics command
	metricsCmd.Flags().BoolVarP(&live, "live", "l", false, "Stream metrics in real-time")
	metricsCmd.Flags().StringVarP(&interval, "interval", "i", "2s", "Update interval for live metrics (e.g., 1s, 5s)")
	metricsCmd.Flags().BoolVarP(&all, "all", "a", false, "Show comprehensive metrics including processes and detailed stats")
}
