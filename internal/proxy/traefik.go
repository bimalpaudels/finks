package proxy

import (
	"context"
	"fmt"
	"strings"

	"github.com/bimalpaudels/finks/internal/docker"
)

const (
	defaultNetworkName   = "finks-default"
	traefikNetworkName   = "finks-traefik"
	traefikContainerName = "finks-traefik"
	traefikImage         = "traefik:v3.0"
)

func GenerateTraefikLabels(config TraefikConfig) map[string]string {
	labels := make(map[string]string)

	// Sanitize app name for router/service names
	routerName := sanitizeName(config.AppName)
	serviceName := sanitizeName(config.AppName)
	networkName := config.NetworkName
	if networkName == "" {
		networkName = defaultNetworkName
	}

	// Basic Traefik configuration
	labels["traefik.enable"] = "true"
	labels["traefik.docker.network"] = networkName

	// Router configuration
	labels[fmt.Sprintf("traefik.http.routers.%s.rule", routerName)] = fmt.Sprintf("Host(`%s`)", config.Domain)
	labels[fmt.Sprintf("traefik.http.routers.%s.service", routerName)] = serviceName

	// Service configuration
	if config.Port != "" {
		labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", serviceName)] = config.Port
	}

	// Configure entrypoints based on mode
	if config.LocalMode {
		// Local development - HTTP only
		labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName)] = "web"
	} else {
		// Production - HTTPS with Let's Encrypt
		labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName)] = "websecure"
		labels[fmt.Sprintf("traefik.http.routers.%s.tls", routerName)] = "true"
		labels[fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", routerName)] = "letsencrypt"

		// HTTP to HTTPS redirect
		redirectRouter := routerName + "-redirect"
		labels[fmt.Sprintf("traefik.http.routers.%s.rule", redirectRouter)] = fmt.Sprintf("Host(`%s`)", config.Domain)
		labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", redirectRouter)] = "web"
		labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", redirectRouter)] = "https-redirect"

		// HTTPS redirect middleware
		labels["traefik.http.middlewares.https-redirect.redirectscheme.scheme"] = "https"
		labels["traefik.http.middlewares.https-redirect.redirectscheme.permanent"] = "true"
	}

	return labels
}

func AddTraefikHealthCheck(labels map[string]string, serviceName, healthPath string) {
	if healthPath != "" {
		labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.healthcheck.path", serviceName)] = healthPath
	}
}

// sanitizeName cleans app name for use in Traefik router/service names
func sanitizeName(name string) string {
	// Replace invalid characters with hyphens
	sanitized := strings.ReplaceAll(name, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ToLower(sanitized)

	// Keep only alphanumeric characters and hyphens
	var result strings.Builder
	for _, r := range sanitized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func InstallTraefik(ctx context.Context, dockerClient *docker.Client, localMode bool) error {
	if err := ensureTraefikNetwork(ctx, dockerClient); err != nil {
		return fmt.Errorf("failed to ensure Traefik network: %w", err)
	}

	exists, err := dockerClient.ContainerExists(ctx, traefikContainerName)
	if err != nil {
		return fmt.Errorf("failed to check if Traefik container exists: %w", err)
	}

	if exists {
		status, err := dockerClient.GetContainerStatus(ctx, traefikContainerName)
		if err != nil {
			return fmt.Errorf("failed to get Traefik container status: %w", err)
		}

		if strings.Contains(strings.ToLower(status), "running") {
			return nil
		}

		if err := dockerClient.StartContainer(ctx, traefikContainerName); err != nil {
			return fmt.Errorf("failed to start existing Traefik container: %w", err)
		}
		return nil
	}

	if err := dockerClient.PullImage(ctx, traefikImage); err != nil {
		return fmt.Errorf("failed to pull Traefik image: %w", err)
	}

	config := buildTraefikConfig(localMode)
	runOptions := docker.RunOptions{
		Name:     traefikContainerName,
		Image:    traefikImage,
		Port:     buildPortMapping(localMode),
		EnvVars:  config,
		Networks: []string{traefikNetworkName},
		Volumes:  buildTraefikVolumes(),
	}

	if err := dockerClient.RunContainer(ctx, runOptions); err != nil {
		return fmt.Errorf("failed to run Traefik container: %w", err)
	}

	return nil
}

func ensureTraefikNetwork(ctx context.Context, dockerClient *docker.Client) error {
	_, err := dockerClient.EnsureNetwork(ctx, traefikNetworkName, "bridge", nil)
	if err != nil {
		return fmt.Errorf("failed to ensure network %s: %w", traefikNetworkName, err)
	}
	return nil
}

func buildTraefikConfig(localMode bool) map[string]string {
	config := map[string]string{
		"TRAEFIK_API_DASHBOARD":                     "true",
		"TRAEFIK_PROVIDERS_DOCKER":                  "true",
		"TRAEFIK_PROVIDERS_DOCKER_EXPOSEDBYDEFAULT": "false",
		"TRAEFIK_ENTRYPOINTS_WEB_ADDRESS":           ":80",
	}

	if localMode {
		config["TRAEFIK_API_INSECURE"] = "true"
	} else {
		config["TRAEFIK_ENTRYPOINTS_WEBSECURE_ADDRESS"] = ":443"
		config["TRAEFIK_CERTIFICATESRESOLVERS_LETSENCRYPT_ACME_TLSCHALLENGE"] = "true"
		config["TRAEFIK_CERTIFICATESRESOLVERS_LETSENCRYPT_ACME_EMAIL"] = "admin@example.com"
		config["TRAEFIK_CERTIFICATESRESOLVERS_LETSENCRYPT_ACME_STORAGE"] = "/letsencrypt/acme.json"
	}

	return config
}

func buildPortMapping(localMode bool) string {
	if localMode {
		return "80:80"
	}
	return "80:80,443:443"
}

func buildTraefikVolumes() []string {
	return []string{
		"/var/run/docker.sock:/var/run/docker.sock:ro",
		"/letsencrypt:/letsencrypt",
	}
}
