package proxy

type TraefikConfig struct {
	AppName     string
	Domain      string
	Port        string
	NetworkName string
	LocalMode   bool
}

type TraefikStatus struct {
	ContainerExists bool
	ContainerStatus string
	NetworkExists   bool
	DashboardURL    string
	IsRunning       bool
}
