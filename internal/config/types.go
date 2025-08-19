package config

import "time"

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
