package proxy

type TraefikConfig struct {
	AppName     string
	Domain      string
	Port        string
	NetworkName string
	LocalMode   bool
}
