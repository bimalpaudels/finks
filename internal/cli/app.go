package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bimalpaudels/finks/internal/app"
	"github.com/spf13/cobra"
)

var (
	appName    string
	appPort    string
	appEnvVars []string
	appVolumes []string
	force      bool
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Application management commands",
	Long:  `Commands for deploying and managing containerized applications.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available app commands:")
		fmt.Println("  deploy     Deploy an application from a Docker image")
		fmt.Println("  start      Start a stopped application")
		fmt.Println("  stop       Stop a running application")
		fmt.Println("  rm         Remove an application")
		fmt.Println("  ps         List all applications")
		fmt.Println("")
		fmt.Println("Use 'finks app [command] --help' for more information.")
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy <image> --name <app-name>",
	Short: "Deploy an application from a Docker image",
	Long: `Deploy a containerized application from an existing Docker image.

Examples:
  finks app deploy nginx --name my-web --port 8080:80
  finks app deploy postgres:13 --name my-db --env POSTGRES_PASSWORD=secret
  finks app deploy redis --name cache --volume /data:/data`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		image := args[0]
		appName := cmd.Flag("name").Value.String()

		if appName == "" {
			fmt.Println("Error: --name flag is required")
			os.Exit(1)
		}

		manager, err := app.NewManager()
		if err != nil {
			fmt.Printf("Error: Failed to initialize app manager: %v\n", err)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := manager.CheckDockerAvailable(ctx); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		envVars := make(map[string]string)
		for _, env := range appEnvVars {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}

		fmt.Printf("🚀 Deploying application '%s' from image '%s'...\n", appName, image)

		if err := manager.DeployApp(ctx, appName, image, appPort, envVars, appVolumes); err != nil {
			fmt.Printf("Error: Failed to deploy application: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Application '%s' deployed successfully!\n", appName)
		if appPort != "" {
			fmt.Printf("   Available at: http://localhost:%s\n", strings.Split(appPort, ":")[0])
		}
	},
}

var startCmd = &cobra.Command{
	Use:   "start <app-name>",
	Short: "Start a stopped application",
	Long:  `Start a previously stopped application.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]

		manager, err := app.NewManager()
		if err != nil {
			fmt.Printf("Error: Failed to initialize app manager: %v\n", err)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := manager.CheckDockerAvailable(ctx); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("▶️  Starting application '%s'...\n", appName)

		if err := manager.StartApp(ctx, appName); err != nil {
			fmt.Printf("Error: Failed to start application: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Application '%s' started successfully!\n", appName)
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop <app-name>",
	Short: "Stop a running application",
	Long:  `Stop a currently running application.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]

		manager, err := app.NewManager()
		if err != nil {
			fmt.Printf("Error: Failed to initialize app manager: %v\n", err)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := manager.CheckDockerAvailable(ctx); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("⏹️  Stopping application '%s'...\n", appName)

		if err := manager.StopApp(ctx, appName); err != nil {
			fmt.Printf("Error: Failed to stop application: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Application '%s' stopped successfully!\n", appName)
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm <app-name>",
	Short: "Remove an application",
	Long:  `Remove an application and its container. Use --force to remove running applications.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]

		manager, err := app.NewManager()
		if err != nil {
			fmt.Printf("Error: Failed to initialize app manager: %v\n", err)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := manager.CheckDockerAvailable(ctx); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🗑️  Removing application '%s'...\n", appName)

		if err := manager.RemoveApp(ctx, appName, force); err != nil {
			fmt.Printf("Error: Failed to remove application: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Application '%s' removed successfully!\n", appName)
	},
}

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List all applications",
	Long:  `List all deployed applications with their current status.`,
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := app.NewManager()
		if err != nil {
			fmt.Printf("Error: Failed to initialize app manager: %v\n", err)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := manager.CheckDockerAvailable(ctx); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		apps, err := manager.ListApps(ctx)
		if err != nil {
			fmt.Printf("Error: Failed to list applications: %v\n", err)
			os.Exit(1)
		}

		if len(apps) == 0 {
			fmt.Println("No applications deployed.")
			return
		}

		fmt.Printf("%-15s %-25s %-15s %-15s %s\n", "NAME", "IMAGE", "STATUS", "PORT", "CREATED")
		fmt.Println(strings.Repeat("-", 80))

		for _, app := range apps {
			status := app.Status
			switch status {
			case "running":
				status = "🟢 " + status
			case "stopped":
				status = "🔴 " + status
			case "failed":
				status = "❌ " + status
			default:
				status = "❓ " + status
			}

			port := app.Port
			if port == "" {
				port = "-"
			}

			fmt.Printf("%-15s %-25s %-15s %-15s %s\n",
				app.Name,
				app.Image,
				status,
				port,
				app.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
	},
}

func init() {
	appCmd.AddCommand(deployCmd, startCmd, stopCmd, rmCmd, psCmd)

	deployCmd.Flags().StringVar(&appName, "name", "", "Name of the application (required)")
	deployCmd.Flags().StringVarP(&appPort, "port", "p", "", "Port mapping (e.g., 8080:80)")
	deployCmd.Flags().StringSliceVarP(&appEnvVars, "env", "e", []string{}, "Environment variables (e.g., KEY=VALUE)")
	deployCmd.Flags().StringSliceVarP(&appVolumes, "volume", "v", []string{}, "Volume mounts (e.g., /host:/container)")
	deployCmd.MarkFlagRequired("name")

	rmCmd.Flags().BoolVarP(&force, "force", "f", false, "Force remove running application")
}