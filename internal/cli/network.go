package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bimalpaudels/finks/internal/docker"
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

		formatNetworkTable(networks)
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

func formatNetworkTable(networks []docker.NetworkInfo) {
	if len(networks) == 0 {
		fmt.Println("No networks found.")
		return
	}

	fmt.Printf("%-15s %-20s %-10s %-18s %-15s\n", "NAME", "NETWORK ID", "DRIVER", "SUBNET", "GATEWAY")
	fmt.Println(strings.Repeat("-", 78))

	for _, net := range networks {
		networkID := net.ID
		if len(networkID) > 12 {
			networkID = networkID[:12]
		}

		subnet := net.Subnet
		if subnet == "" {
			subnet = "-"
		}

		gateway := net.Gateway
		if gateway == "" {
			gateway = "-"
		}

		fmt.Printf("%-15s %-20s %-10s %-18s %-15s\n",
			net.Name,
			networkID,
			net.Driver,
			subnet,
			gateway,
		)
	}
}

func init() {
	networkCmd.AddCommand(listNetworksCmd, createNetworkCmd)

	// Add flags for create command
	createNetworkCmd.Flags().StringP("driver", "d", "bridge", "Network driver (bridge, overlay, etc.)")

	rootCmd.AddCommand(networkCmd)
}
