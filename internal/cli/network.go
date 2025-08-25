package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bimalpaudels/finks/internal/docker"
	"github.com/pterm/pterm"
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

var listNetworksCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Docker networks",
	Long:  `List all Docker networks with their details including name, driver, and subnet information.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := docker.NewClient()
		if err != nil {
			fmt.Printf("Error: Failed to initialize Docker client: %v\n", err)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		networks, err := client.ListNetworks(ctx)
		if err != nil {
			fmt.Printf("Error: Failed to list networks: %v\n", err)
			os.Exit(1)
		}

		filteredNetworks := filterFinksNetworks(networks)
		formatNetworkTable(filteredNetworks)
	},
}

var createNetworkCmd = &cobra.Command{
	Use:   "create <network-name>",
	Short: "Create a new Docker network",
	Long:  `Create a new Docker network with the specified name and optional driver.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		networkName := "finks-" + args[0]
		driver := cmd.Flag("driver").Value.String()
		if driver == "" {
			driver = "bridge"
		}

		client, err := docker.NewClient()
		if err != nil {
			fmt.Printf("Error: Failed to initialize Docker client: %v\n", err)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		fmt.Printf("🔧 Creating network '%s' with driver '%s'...\n", networkName, driver)

		networkID, err := client.CreateNetwork(ctx, networkName, driver, nil)
		if err != nil {
			fmt.Printf("Error: Failed to create network: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Network '%s' created successfully!\n", networkName)
		fmt.Printf("   Network ID: %s\n", networkID[:12])
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
		if strings.HasPrefix(net.Name, "finks-") {
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
