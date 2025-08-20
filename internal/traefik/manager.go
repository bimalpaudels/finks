package traefik

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bimalpaudels/finks/internal/docker"
)

// NewManager creates a new Traefik manager instance
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".finks")
	configPath := filepath.Join(dataDir, "traefik.json")

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
			ContainerName: DefaultContainerName,
			Image:         DefaultImage,
			Network:       DefaultNetwork,
			LocalMode:     true,
			Entrypoints:   DefaultEntrypoints,
			Status:        StatusStopped,
		},
	}

	if err := manager.loadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return manager, nil
}

// Close closes the manager and cleans up resources
func (m *Manager) Close() error {
	return m.dockerClient.Close()
}

// Setup initializes Traefik container with the specified configuration
func (m *Manager) Setup(ctx context.Context, email string, localMode bool) error {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return fmt.Errorf("Docker is not available: %w", err)
	}

	// Check if Traefik is already running
	if exists, err := m.dockerClient.ContainerExists(ctx, m.config.ContainerName); err != nil {
		return fmt.Errorf("failed to check if Traefik container exists: %w", err)
	} else if exists {
		return fmt.Errorf("Traefik container already exists: %s", m.config.ContainerName)
	}

	// Update configuration
	m.config.Email = email
	m.config.LocalMode = localMode
	m.config.Status = StatusRunning
	m.config.UpdatedAt = time.Now()
	if m.config.CreatedAt.IsZero() {
		m.config.CreatedAt = time.Now()
	}

	// Pull Traefik image
	if err := m.dockerClient.PullImage(ctx, m.config.Image); err != nil {
		return fmt.Errorf("failed to pull Traefik image: %w", err)
	}

	// Prepare container run options
	runOpts := m.buildRunOptions()

	// Run Traefik container
	if err := m.dockerClient.RunContainer(ctx, runOpts); err != nil {
		m.config.Status = StatusFailed
		if saveErr := m.saveConfig(); saveErr != nil {
			return fmt.Errorf("failed to run Traefik container: %w, failed to save config: %v", err, saveErr)
		}
		return fmt.Errorf("failed to run Traefik container: %w", err)
	}

	// Save configuration
	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("failed to save Traefik configuration: %w", err)
	}

	return nil
}

// Stop stops the Traefik container
func (m *Manager) Stop(ctx context.Context) error {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return fmt.Errorf("Docker is not available: %w", err)
	}

	if err := m.dockerClient.StopContainer(ctx, m.config.ContainerName); err != nil {
		return fmt.Errorf("failed to stop Traefik container: %w", err)
	}

	m.config.Status = StatusStopped
	m.config.UpdatedAt = time.Now()

	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// Start starts the Traefik container
func (m *Manager) Start(ctx context.Context) error {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return fmt.Errorf("Docker is not available: %w", err)
	}

	if err := m.dockerClient.StartContainer(ctx, m.config.ContainerName); err != nil {
		return fmt.Errorf("failed to start Traefik container: %w", err)
	}

	m.config.Status = StatusRunning
	m.config.UpdatedAt = time.Now()

	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// Remove removes the Traefik container and cleans up configuration
func (m *Manager) Remove(ctx context.Context, force bool) error {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return fmt.Errorf("Docker is not available: %w", err)
	}

	if err := m.dockerClient.RemoveContainer(ctx, m.config.ContainerName, force); err != nil {
		return fmt.Errorf("failed to remove Traefik container: %w", err)
	}

	m.config.Status = StatusStopped
	m.config.UpdatedAt = time.Now()

	if err := m.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetStatus returns the current status of Traefik
func (m *Manager) GetStatus(ctx context.Context) (*Config, error) {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return m.config, nil // Return cached status if Docker is not available
	}

	// Check actual container status
	containers, err := m.dockerClient.ListContainers(ctx)
	if err != nil {
		return m.config, nil // Return cached status on error
	}

	// Update status based on actual container state
	found := false
	for _, container := range containers {
		if container.Name == m.config.ContainerName {
			found = true
			if container.Status == "running" {
				m.config.Status = StatusRunning
			} else {
				m.config.Status = StatusStopped
			}
			break
		}
	}

	if !found {
		m.config.Status = StatusStopped
	}

	return m.config, nil
}

// IsRunning checks if Traefik container is currently running
func (m *Manager) IsRunning(ctx context.Context) (bool, error) {
	status, err := m.GetStatus(ctx)
	if err != nil {
		return false, err
	}
	return status.Status == StatusRunning, nil
}

// buildRunOptions constructs Docker run options for Traefik container
func (m *Manager) buildRunOptions() docker.RunOptions {
	command := []string{
		"--api.dashboard=true",
		"--providers.docker=true",
		"--providers.docker.exposedbydefault=false",
		"--providers.docker.network=" + m.config.Network,
		"--entrypoints.web.address=:80",
		"--entrypoints.traefik.address=:8080",
	}

	// Add insecure API for local mode
	if m.config.LocalMode {
		command = append(command, "--api.insecure=true")
	} else {
		// Production mode with Let's Encrypt
		if m.config.Email != "" {
			command = append(command,
				"--entrypoints.websecure.address=:443",
				"--certificatesresolvers.letsencrypt.acme.email="+m.config.Email,
				"--certificatesresolvers.letsencrypt.acme.storage=/acme.json",
				"--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web",
			)
		}
	}

	ports := []string{"80:80", "8080:8080"}
	if !m.config.LocalMode {
		ports = append(ports, "443:443")
	}

	volumes := []string{"/var/run/docker.sock:/var/run/docker.sock:ro"}
	if !m.config.LocalMode && m.config.Email != "" {
		volumes = append(volumes, "traefik-acme:/acme.json")
	}

	// For now, use basic RunOptions structure
	// TODO: Extend docker.RunOptions to support advanced Traefik configuration
	return docker.RunOptions{
		Name:    m.config.ContainerName,
		Image:   m.config.Image,
		Port:    "80:80,8080:8080", // Combined ports for basic functionality
		Volumes: volumes,
	}
}

// loadConfig loads Traefik configuration from file
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

	// Ensure default values are set
	if m.config.ContainerName == "" {
		m.config.ContainerName = DefaultContainerName
	}
	if m.config.Image == "" {
		m.config.Image = DefaultImage
	}
	if m.config.Network == "" {
		m.config.Network = DefaultNetwork
	}
	if m.config.Entrypoints == nil {
		m.config.Entrypoints = DefaultEntrypoints
	}

	return nil
}

// saveConfig saves Traefik configuration to file
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