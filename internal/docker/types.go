package docker

type RunOptions struct {
	Name    string
	Image   string
	Port    string
	EnvVars map[string]string
	Volumes []string
}

type Container struct {
	Name   string
	Image  string
	Status string
	Ports  string
}
