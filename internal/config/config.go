package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func Load(configPath string) (*Config, error) {
	// Set defaults
	config := &Config{
		Deployment: DeploymentConfig{
			DataDir: "/var/lib/finks",
		},
		Monitoring: MonitoringConfig{
			MetricsInterval:     30 * time.Second,
			HealthCheckInterval: 10 * time.Second,
		},
		Docker: DockerConfig{
			Socket:   "/var/run/docker.sock",
			Network:  "finks-network",
			Registry: "",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	return config, nil
}
