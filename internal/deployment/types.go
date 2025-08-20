package deployment

import (
	"time"

	"github.com/bimalpaudels/finks/internal/docker"
)

type App struct {
	Name      string            `json:"name"`
	Image     string            `json:"image"`
	Port      string            `json:"port,omitempty"`
	EnvVars   map[string]string `json:"env_vars,omitempty"`
	Volumes   []string          `json:"volumes,omitempty"`
	Status    string            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type Config struct {
	Apps    map[string]*App `json:"apps"`
	DataDir string          `json:"data_dir"`
}

type Manager struct {
	dockerClient *docker.Client
	configPath   string
	config       *Config
}

const (
	StatusRunning = "running"
	StatusStopped = "stopped"
	StatusFailed  = "failed"
	StatusUnknown = "unknown"
)
