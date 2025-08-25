package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bimalpaudels/finks/internal/docker"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

const finksNetworkPrefix = "finks-"

var dockerClient *docker.Client

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage Docker networks",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		dockerClient, err = docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to initialize Docker client: %w", err)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listNetworksCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Docker networks",
	Long:  `List all Docker networks with their details including name, driver, and subnet information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		networks, err := dockerClient.ListNetworks(ctx)
		if err != nil {
			return fmt.Errorf("failed to list networks: %w", err)
		}

		filteredNetworks := filterFinksNetworks(networks)
		formatNetworkTable(filteredNetworks)
		return nil
	},
}

var createNetworkCmd = &cobra.Command{
	Use:   "create <network-name>",
	Short: "Create a new Docker network",
	Long:  `Create a new Docker network with the specified name and optional driver.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		networkName := finksNetworkPrefix + args[0]
		driver, _ := cmd.Flags().GetString("driver")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Creating network '%s' with driver '%s'...", networkName, driver))

		networkID, err := dockerClient.CreateNetwork(ctx, networkName, driver, nil)
		if err != nil {
			spinner.Fail(fmt.Sprintf("Failed to create network: %v", err))
			return fmt.Errorf("failed to create network: %w", err)
		}

		spinner.Success(fmt.Sprintf("Network '%s' created successfully!", networkName))
		pterm.Success.Println(fmt.Sprintf("Network ID: %s", networkID[:12]))
		return nil
	},
}

func valueOrDefault(value, placeholder string) string {
	if value == "" {
		return placeholder
	}
	return value
}

func formatNetworkTable(networks []docker.NetworkInfo) {
	if len(networks) == 0 {
		pterm.Warning.Println("No finks networks found.")
		return
	}

	tableData := make(pterm.TableData, 1, len(networks)+1)
	tableData[0] = []string{"NAME", "NETWORK ID", "DRIVER", "SUBNET", "GATEWAY"}

	for _, net := range networks {
		networkID := net.ID
		if len(networkID) > 12 {
			networkID = networkID[:12]
		}

		tableData = append(tableData, []string{
			net.Name,
			networkID,
			net.Driver,
			valueOrDefault(net.Subnet, "-"),
			valueOrDefault(net.Gateway, "-"),
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
}

func filterFinksNetworks(networks []docker.NetworkInfo) []docker.NetworkInfo {
	filteredNetworks := []docker.NetworkInfo{}
	for _, net := range networks {
		if strings.HasPrefix(net.Name, finksNetworkPrefix) {
			filteredNetworks = append(filteredNetworks, net)
		}
	}
	return filteredNetworks
}

func init() {
	networkCmd.AddCommand(listNetworksCmd, createNetworkCmd)

	// Add flags for create command
	createNetworkCmd.Flags().StringP("driver", "d", "bridge", "Network driver (bridge, overlay, etc.)")

	rootCmd.AddCommand(networkCmd)
}
