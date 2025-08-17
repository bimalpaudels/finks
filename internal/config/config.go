package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Deployment DeploymentConfig `yaml:"deployment"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	Docker     DockerConfig     `yaml:"docker"`
	Logging    LoggingConfig    `yaml:"logging"`
}

type DeploymentConfig struct {
	DataDir string `yaml:"data_dir"`
}

type MonitoringConfig struct {
	MetricsInterval     time.Duration `yaml:"metrics_interval"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
}

type DockerConfig struct {
	Socket   string `yaml:"socket"`
	Network  string `yaml:"network"`
	Registry string `yaml:"registry"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

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
