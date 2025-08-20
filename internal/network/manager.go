package network

import (
	"context"
	"fmt"
	"time"

	"github.com/bimalpaudels/finks/internal/docker"
)

// NewManager creates a new network manager instance
func NewManager() (*Manager, error) {
	dockerClient, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	config := &Config{
		NetworkName: DefaultNetworkName,
		Driver:      DefaultDriver,
		Subnet:      DefaultSubnet,
		Gateway:     DefaultGateway,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return &Manager{
		dockerClient: dockerClient,
		config:       config,
	}, nil
}

// Close closes the network manager and cleans up resources
func (m *Manager) Close() error {
	return m.dockerClient.Close()
}

// EnsureNetwork ensures the Finks network exists, creating it if necessary
func (m *Manager) EnsureNetwork(ctx context.Context) error {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return fmt.Errorf("Docker is not available: %w", err)
	}

	exists, err := m.NetworkExists(ctx, m.config.NetworkName)
	if err != nil {
		return fmt.Errorf("failed to check if network exists: %w", err)
	}

	if exists {
		return nil // Network already exists
	}

	// Create the network
	if err := m.CreateNetwork(ctx); err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	return nil
}

// CreateNetwork creates the Finks Docker network
func (m *Manager) CreateNetwork(ctx context.Context) error {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return fmt.Errorf("Docker is not available: %w", err)
	}

	// For now, we'll use a basic approach since the docker client may not have network creation methods
	// This is a placeholder that would need to be implemented when extending docker.Client
	return fmt.Errorf("network creation not yet implemented - extend docker.Client with network methods")
}

// NetworkExists checks if a network with the given name exists
func (m *Manager) NetworkExists(ctx context.Context, networkName string) (bool, error) {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return false, fmt.Errorf("Docker is not available: %w", err)
	}

	// For now, assume network doesn't exist
	// This would need to be implemented when extending docker.Client
	return false, nil
}

// GetNetworkInfo retrieves information about the Finks network
func (m *Manager) GetNetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return nil, fmt.Errorf("Docker is not available: %w", err)
	}

	// Placeholder implementation
	return &NetworkInfo{
		Name:    m.config.NetworkName,
		Driver:  m.config.Driver,
		Subnet:  m.config.Subnet,
		Gateway: m.config.Gateway,
		Labels:  DefaultLabels,
		Created: m.config.CreatedAt,
	}, nil
}

// ConnectContainer connects a container to the Finks network
func (m *Manager) ConnectContainer(ctx context.Context, containerNameOrID string) error {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return fmt.Errorf("Docker is not available: %w", err)
	}

	// Ensure network exists first
	if err := m.EnsureNetwork(ctx); err != nil {
		return fmt.Errorf("failed to ensure network exists: %w", err)
	}

	// Connect container to network
	// This would need to be implemented when extending docker.Client
	return fmt.Errorf("container network connection not yet implemented - extend docker.Client")
}

// DisconnectContainer disconnects a container from the Finks network
func (m *Manager) DisconnectContainer(ctx context.Context, containerNameOrID string) error {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return fmt.Errorf("Docker is not available: %w", err)
	}

	// Disconnect container from network
	// This would need to be implemented when extending docker.Client
	return fmt.Errorf("container network disconnection not yet implemented - extend docker.Client")
}

// ListConnections lists all containers connected to the Finks network
func (m *Manager) ListConnections(ctx context.Context) ([]ConnectionInfo, error) {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return nil, fmt.Errorf("Docker is not available: %w", err)
	}

	// List network connections
	// This would need to be implemented when extending docker.Client
	return []ConnectionInfo{}, nil
}

// RemoveNetwork removes the Finks network (use with caution)
func (m *Manager) RemoveNetwork(ctx context.Context, force bool) error {
	if err := m.dockerClient.IsAvailable(ctx); err != nil {
		return fmt.Errorf("Docker is not available: %w", err)
	}

	// Remove network
	// This would need to be implemented when extending docker.Client
	return fmt.Errorf("network removal not yet implemented - extend docker.Client")
}

// ValidateNetworkConfig validates the network configuration
func (m *Manager) ValidateNetworkConfig() error {
	if m.config.NetworkName == "" {
		return fmt.Errorf("network name cannot be empty")
	}

	if m.config.Driver == "" {
		return fmt.Errorf("network driver cannot be empty")
	}

	// Additional validation could be added here for subnet/gateway format
	return nil
}

// GetConfig returns the current network configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// UpdateConfig updates the network configuration
func (m *Manager) UpdateConfig(config *Config) error {
	if err := m.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	config.UpdatedAt = time.Now()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}

	m.config = config
	return nil
}

// validateConfig validates a network configuration
func (m *Manager) validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.NetworkName == "" {
		return fmt.Errorf("network name cannot be empty")
	}

	if config.Driver == "" {
		return fmt.Errorf("driver cannot be empty")
	}

	return nil
}