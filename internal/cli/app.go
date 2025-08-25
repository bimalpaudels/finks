package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bimalpaudels/finks/internal/deployment"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	appPort    string
	appEnvVars []string
	appVolumes []string
	force      bool
)

var appManager *deployment.Manager

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Application management commands",
	Long:  `Commands for deploying and managing containerized applications.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		appManager, err = deployment.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize app manager: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := appManager.CheckDockerAvailable(ctx); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
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
	RunE: func(cmd *cobra.Command, args []string) error {
		image := args[0]
		appName, _ := cmd.Flags().GetString("name")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		envVars := parseEnvVars(appEnvVars)

		spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Deploying application '%s' from image '%s'...", appName, image))

		if err := appManager.DeployApp(ctx, appName, image, appPort, envVars, appVolumes); err != nil {
			spinner.Fail(fmt.Sprintf("Failed to deploy application: %v", err))
			return fmt.Errorf("failed to deploy application: %w", err)
		}

		spinner.Success(fmt.Sprintf("Application '%s' deployed successfully!", appName))
		if appPort != "" {
			pterm.Info.Println(fmt.Sprintf("Available at: http://localhost:%s", strings.Split(appPort, ":")[0]))
		}
		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start <app-name>",
	Short: "Start a stopped application",
	Long:  `Start a previously stopped application.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Starting application '%s'...", appName))

		if err := appManager.StartApp(ctx, appName); err != nil {
			spinner.Fail(fmt.Sprintf("Failed to start application: %v", err))
			return fmt.Errorf("failed to start application: %w", err)
		}

		spinner.Success(fmt.Sprintf("Application '%s' started successfully!", appName))
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop <app-name>",
	Short: "Stop a running application",
	Long:  `Stop a currently running application.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Stopping application '%s'...", appName))

		if err := appManager.StopApp(ctx, appName); err != nil {
			spinner.Fail(fmt.Sprintf("Failed to stop application: %v", err))
			return fmt.Errorf("failed to stop application: %w", err)
		}

		spinner.Success(fmt.Sprintf("Application '%s' stopped successfully!", appName))
		return nil
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove <app-name>",
	Short: "Remove an application",
	Long:  `Remove an application and its container. Use --force to remove running applications.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Removing application '%s'...", appName))

		if err := appManager.RemoveApp(ctx, appName, force); err != nil {
			spinner.Fail(fmt.Sprintf("Failed to remove application: %v", err))
			return fmt.Errorf("failed to remove application: %w", err)
		}

		spinner.Success(fmt.Sprintf("Application '%s' removed successfully!", appName))
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applications",
	Long:  `List all deployed applications with their current status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		apps, err := appManager.ListApps(ctx)
		if err != nil {
			return fmt.Errorf("failed to list applications: %w", err)
		}

		if len(apps) == 0 {
			pterm.Info.Println("No applications deployed.")
			return nil
		}

		tableData := pterm.TableData{{"NAME", "IMAGE", "STATUS", "PORT", "CREATED"}}
		for _, app := range apps {
			status := getStatusIcon(app.Status) + " " + app.Status
			port := valueOrDefault(app.Port, "-")
			tableData = append(tableData, []string{
				app.Name,
				app.Image,
				status,
				port,
				app.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		return nil
	},
}

func parseEnvVars(envVars []string) map[string]string {
	result := make(map[string]string)
	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func getStatusIcon(status string) string {
	switch status {
	case "running":
		return "🟢"
	case "stopped":
		return "🔴"
	case "failed":
		return "❌"
	default:
		return "❓"
	}
}

func init() {
	appCmd.AddCommand(deployCmd, startCmd, stopCmd, removeCmd, listCmd)

	deployCmd.Flags().String("name", "", "Name of the application (required)")
	deployCmd.Flags().StringVarP(&appPort, "port", "p", "", "Port mapping (e.g., 8080:80)")
	deployCmd.Flags().StringSliceVarP(&appEnvVars, "env", "e", []string{}, "Environment variables (e.g., KEY=VALUE)")
	deployCmd.Flags().StringSliceVarP(&appVolumes, "volume", "v", []string{}, "Volume mounts (e.g., /host:/container)")
	deployCmd.MarkFlagRequired("name")

	removeCmd.Flags().BoolVarP(&force, "force", "f", false, "Force remove running application")
}
