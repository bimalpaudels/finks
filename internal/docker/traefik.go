package docker

import (
	"fmt"
	"strings"
)


func GenerateTraefikLabels(config TraefikConfig) map[string]string {
	labels := make(map[string]string)
	
	// Sanitize app name for router/service names
	routerName := sanitizeName(config.AppName)
	serviceName := sanitizeName(config.AppName)
	networkName := config.NetworkName
	if networkName == "" {
		networkName = "finks-network"
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