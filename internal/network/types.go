package network

import (
	"time"

	"github.com/bimalpaudels/finks/internal/docker"
)

// Manager handles Docker network management for Finks
type Manager struct {
	dockerClient *docker.Client
	config       *Config
}

// Config represents network configuration
type Config struct {
	NetworkName string    `json:"network_name"`
	Driver      string    `json:"driver"`
	Subnet      string    `json:"subnet,omitempty"`
	Gateway     string    `json:"gateway,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NetworkInfo represents information about a Docker network
type NetworkInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Driver  string            `json:"driver"`
	Subnet  string            `json:"subnet"`
	Gateway string            `json:"gateway"`
	Labels  map[string]string `json:"labels"`
	Created time.Time         `json:"created"`
}

// ConnectionInfo represents a container's connection to a network
type ConnectionInfo struct {
	ContainerName string `json:"container_name"`
	ContainerID   string `json:"container_id"`
	IPAddress     string `json:"ip_address"`
	MacAddress    string `json:"mac_address"`
}

// Default configuration values
const (
	DefaultNetworkName = "finks-network"
	DefaultDriver      = "bridge"
	DefaultSubnet      = "172.20.0.0/16"
	DefaultGateway     = "172.20.0.1"
)

// Network management constants
const (
	LabelManagedBy = "finks.managed-by"
	LabelCreatedBy = "finks.created-by"
	LabelVersion   = "finks.version"
)

// Default labels for Finks-managed networks
var DefaultLabels = map[string]string{
	LabelManagedBy: "finks",
	LabelCreatedBy: "finks-network-manager",
	LabelVersion:   "1.0",
}