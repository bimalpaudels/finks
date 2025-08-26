package docker

type RunOptions struct {
	Name     string
	Image    string
	Ports    []string
	EnvVars  map[string]string
	Volumes  []string
	Labels   map[string]string // Added for Traefik labels
	Networks []string          // Added for network connections
}

type Container struct {
	Name   string
	Image  string
	Status string
	Ports  string
}

type NetworkInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Driver  string            `json:"driver"`
	Subnet  string            `json:"subnet"`
	Gateway string            `json:"gateway"`
	Labels  map[string]string `json:"labels"`
}

