package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bimalpaudels/finks/internal/docker"
)

func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".finks")
	configPath := filepath.Join(dataDir, "apps.json")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dockerClient, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	manager := &Manager{
		dockerClient: dockerClient,
		configPath:   configPath,
		config: &Config{
			Apps:    make(map[string]*App),
			DataDir: dataDir,
		},
	}

	if err := manager.loadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return manager, nil
}

func (m *Manager) Close() error {
	return m.dockerClient.Close()
}

func (m *Manager) CheckDockerAvailable(ctx context.Context) error {
	return m.dockerClient.IsAvailable(ctx)
}

func (m *Manager) DeployApp(ctx context.Context, name, image, port string, envVars map[string]string, volumes []string) error {
	if err := m.CheckDockerAvailable(ctx); err != nil {
		return err
	}

	containerName := fmt.Sprintf("finks-%s", name)

	if exists, err := m.dockerClient.ContainerExists(ctx, containerName); err != nil {
		return fmt.Errorf("failed to check if container exists: %w", err)
	} else if exists {
		return fmt.Errorf("application %s already exists", name)
	}

	if err := m.dockerClient.PullImage(ctx, image); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	runOpts := docker.RunOptions{
		Name:    containerName,
		Image:   image,
		Port:    port,
		EnvVars: envVars,
		Volumes: volumes,
	}

	if err := m.dockerClient.RunContainer(ctx, runOpts); err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}

	app := &App{
		Name:      name,
		Image:     image,
		Port:      port,
		EnvVars:   envVars,
		Volumes:   volumes,
		Status:    StatusRunning,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.config.Apps[name] = app
	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func (m *Manager) StopApp(ctx context.Context, name string) error {
	if err := m.CheckDockerAvailable(ctx); err != nil {
		return err
	}

	app, exists := m.config.Apps[name]
	if !exists {
		return fmt.Errorf("application %s not found", name)
	}

	containerName := fmt.Sprintf("finks-%s", name)
	if err := m.dockerClient.StopContainer(ctx, containerName); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	app.Status = StatusStopped
	app.UpdatedAt = time.Now()
	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func (m *Manager) StartApp(ctx context.Context, name string) error {
	if err := m.CheckDockerAvailable(ctx); err != nil {
		return err
	}

	app, exists := m.config.Apps[name]
	if !exists {
		return fmt.Errorf("application %s not found", name)
	}

	containerName := fmt.Sprintf("finks-%s", name)
	if err := m.dockerClient.StartContainer(ctx, containerName); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	app.Status = StatusRunning
	app.UpdatedAt = time.Now()
	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func (m *Manager) RemoveApp(ctx context.Context, name string, force bool) error {
	if err := m.CheckDockerAvailable(ctx); err != nil {
		return err
	}

	_, exists := m.config.Apps[name]
	if !exists {
		return fmt.Errorf("application %s not found", name)
	}

	containerName := fmt.Sprintf("finks-%s", name)
	if err := m.dockerClient.RemoveContainer(ctx, containerName, force); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	delete(m.config.Apps, name)
	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func (m *Manager) ListApps(ctx context.Context) ([]*App, error) {
	if err := m.CheckDockerAvailable(ctx); err != nil {
		return nil, err
	}

	containers, err := m.dockerClient.ListContainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	containerStatuses := make(map[string]string)
	for _, container := range containers {
		if appName, found := strings.CutPrefix(container.Name, "finks-"); found {
			status := StatusRunning
			if strings.Contains(strings.ToLower(container.Status), "exited") {
				status = StatusStopped
			}
			containerStatuses[appName] = status
		}
	}

	apps := make([]*App, 0, len(m.config.Apps))
	for name, app := range m.config.Apps {
		if status, exists := containerStatuses[name]; exists {
			app.Status = status
		} else {
			app.Status = StatusUnknown
		}
		app.UpdatedAt = time.Now()
		apps = append(apps, app)
	}

	if err := m.saveConfig(); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return apps, nil
}

func (m *Manager) GetApp(name string) (*App, error) {
	app, exists := m.config.Apps[name]
	if !exists {
		return nil, fmt.Errorf("application %s not found", name)
	}
	return app, nil
}

func (m *Manager) loadConfig() error {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if m.config.Apps == nil {
		m.config.Apps = make(map[string]*App)
	}

	return nil
}

func (m *Manager) saveConfig() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
