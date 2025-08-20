package traefik

import (
	"time"

	"github.com/bimalpaudels/finks/internal/docker"
)

// Config represents the Traefik configuration
type Config struct {
	ContainerName string            `json:"container_name"`
	Image         string            `json:"image"`
	Network       string            `json:"network"`
	Email         string            `json:"email,omitempty"`
	LocalMode     bool              `json:"local_mode"`
	Entrypoints   map[string]string `json:"entrypoints"`
	Status        string            `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// Manager handles Traefik container lifecycle and configuration
type Manager struct {
	dockerClient *docker.Client
	config       *Config
	configPath   string
}

// RouterConfig represents a Traefik router configuration for an application
type RouterConfig struct {
	Name        string            `json:"name"`
	Rule        string            `json:"rule"`
	Service     string            `json:"service"`
	Entrypoint  string            `json:"entrypoint"`
	TLS         *TLSConfig        `json:"tls,omitempty"`
	Middlewares []string          `json:"middlewares,omitempty"`
	Labels      map[string]string `json:"labels"`
}

// TLSConfig represents TLS/SSL configuration for a router
type TLSConfig struct {
	CertResolver string   `json:"cert_resolver"`
	Domains      []string `json:"domains,omitempty"`
}

// ServiceConfig represents a Traefik service configuration
type ServiceConfig struct {
	Name string `json:"name"`
	Port string `json:"port"`
	URL  string `json:"url,omitempty"`
}

// LabelSet represents a complete set of Docker labels for Traefik routing
type LabelSet struct {
	Enable      string            `json:"enable"`
	Router      map[string]string `json:"router"`
	Service     map[string]string `json:"service"`
	Middlewares map[string]string `json:"middlewares,omitempty"`
}

// Status constants for Traefik manager
const (
	StatusRunning = "running"
	StatusStopped = "stopped"
	StatusFailed  = "failed"
	StatusUnknown = "unknown"
)

// Default configuration values
const (
	DefaultImage         = "traefik:v3.5"
	DefaultContainerName = "finks-traefik"
	DefaultNetwork       = "finks-network"
	DefaultWebPort       = "80"
	DefaultWebSecurePort = "443"
	DefaultAPIPort       = "8080"
)

// Entrypoint names
const (
	EntrypointWeb       = "web"
	EntrypointWebSecure = "websecure"
	EntrypointTraefik   = "traefik"
)

// Default entrypoints configuration
var DefaultEntrypoints = map[string]string{
	EntrypointWeb:       ":" + DefaultWebPort,
	EntrypointWebSecure: ":" + DefaultWebSecurePort,
	EntrypointTraefik:   ":" + DefaultAPIPort,
}