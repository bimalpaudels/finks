package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/bimalpaudels/finks/internal/docker"
	"github.com/bimalpaudels/finks/internal/proxy"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var proxyDockerClient *docker.Client

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Manage Traefik proxy",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		proxyDockerClient, err = docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to initialize Docker client: %w", err)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var installProxyCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Traefik proxy container",
	Long:  `Install and configure Traefik proxy container with proper networking setup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		spinner, _ := pterm.DefaultSpinner.Start("Installing Traefik proxy...")

		if err := proxy.InstallTraefik(ctx, proxyDockerClient); err != nil {
			spinner.Fail(fmt.Sprintf("Failed to install Traefik: %v", err))
			return fmt.Errorf("failed to install Traefik: %w", err)
		}

		spinner.Success("Traefik proxy installed successfully!")
		pterm.Success.Println("Traefik dashboard available at: http://localhost:8080/dashboard/")

		return nil
	},
}

var statusProxyCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Traefik proxy status",
	Long:  `Check the status of the Traefik proxy container and network configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		status, err := proxy.GetTraefikStatus(ctx, proxyDockerClient)
		if err != nil {
			return fmt.Errorf("failed to get Traefik status: %w", err)
		}

		if !status.ContainerExists {
			pterm.Warning.Println("Traefik container is not installed")
			pterm.Info.Println("Run 'finks proxy install' to install Traefik")
			return nil
		}

		pterm.Success.Println(fmt.Sprintf("Traefik container: %s", status.ContainerStatus))

		if status.NetworkExists {
			pterm.Success.Println("Traefik network: finks-traefik exists")
		} else {
			pterm.Warning.Println("Traefik network: finks-traefik does not exist")
		}

		if status.IsRunning {
			pterm.Info.Println(fmt.Sprintf("Traefik dashboard: %s", status.DashboardURL))
		}

		return nil
	},
}

var connectProxyCmd = &cobra.Command{
	Use:   "connect <network-name>",
	Short: "Connect Traefik to an application network",
	Long:  `Connect the Traefik proxy container to a specific application network for routing.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		networkName := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Connecting Traefik to network '%s'...", networkName))

		if err := proxyDockerClient.ConnectContainerToNetwork(ctx, networkName, "finks-traefik"); err != nil {
			spinner.Fail(fmt.Sprintf("Failed to connect Traefik to network: %v", err))
			return fmt.Errorf("failed to connect Traefik to network %s: %w", networkName, err)
		}

		spinner.Success(fmt.Sprintf("Traefik connected to network '%s' successfully!", networkName))
		return nil
	},
}

func init() {
	proxyCmd.AddCommand(installProxyCmd, statusProxyCmd, connectProxyCmd)
}