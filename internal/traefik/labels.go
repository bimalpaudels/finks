package traefik

import (
	"fmt"
	"strings"
)

// GenerateLabels creates Docker labels for Traefik routing based on app configuration
func GenerateLabels(appName, domain, port string, localMode bool) map[string]string {
	labels := make(map[string]string)
	
	// Sanitize app name for use in router names
	routerName := sanitizeRouterName(appName)
	serviceName := sanitizeServiceName(appName)

	// Basic Traefik configuration
	labels["traefik.enable"] = "true"
	labels["traefik.docker.network"] = DefaultNetwork

	// Router configuration
	labels[fmt.Sprintf("traefik.http.routers.%s.rule", routerName)] = fmt.Sprintf("Host(`%s`)", domain)
	labels[fmt.Sprintf("traefik.http.routers.%s.service", routerName)] = serviceName

	// Service configuration
	if port != "" {
		labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", serviceName)] = port
	}

	// Configure entrypoints and TLS based on mode
	if localMode {
		// Local development mode - HTTP only
		labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName)] = EntrypointWeb
	} else {
		// Production mode - HTTPS with Let's Encrypt
		labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName)] = EntrypointWebSecure
		labels[fmt.Sprintf("traefik.http.routers.%s.tls", routerName)] = "true"
		labels[fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", routerName)] = "letsencrypt"

		// HTTP to HTTPS redirect router
		redirectRouterName := routerName + "-redirect"
		labels[fmt.Sprintf("traefik.http.routers.%s.rule", redirectRouterName)] = fmt.Sprintf("Host(`%s`)", domain)
		labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", redirectRouterName)] = EntrypointWeb
		labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", redirectRouterName)] = "https-redirect"
		
		// HTTPS redirect middleware
		labels["traefik.http.middlewares.https-redirect.redirectscheme.scheme"] = "https"
		labels["traefik.http.middlewares.https-redirect.redirectscheme.permanent"] = "true"
	}

	return labels
}

// GenerateLabelsFromConfig creates Docker labels from RouterConfig
func GenerateLabelsFromConfig(config *RouterConfig) map[string]string {
	labels := make(map[string]string)
	
	// Enable Traefik
	labels["traefik.enable"] = "true"
	labels["traefik.docker.network"] = DefaultNetwork

	// Router configuration
	routerPrefix := fmt.Sprintf("traefik.http.routers.%s", config.Name)
	labels[routerPrefix+".rule"] = config.Rule
	labels[routerPrefix+".service"] = config.Service
	labels[routerPrefix+".entrypoints"] = config.Entrypoint

	// TLS configuration
	if config.TLS != nil {
		labels[routerPrefix+".tls"] = "true"
		if config.TLS.CertResolver != "" {
			labels[routerPrefix+".tls.certresolver"] = config.TLS.CertResolver
		}
	}

	// Middlewares
	if len(config.Middlewares) > 0 {
		labels[routerPrefix+".middlewares"] = strings.Join(config.Middlewares, ",")
	}

	// Add any additional custom labels
	for key, value := range config.Labels {
		labels[key] = value
	}

	return labels
}

// CreateRouterConfig creates a RouterConfig for an application
func CreateRouterConfig(appName, domain, port string, localMode bool) *RouterConfig {
	routerName := sanitizeRouterName(appName)
	serviceName := sanitizeServiceName(appName)

	config := &RouterConfig{
		Name:    routerName,
		Rule:    fmt.Sprintf("Host(`%s`)", domain),
		Service: serviceName,
		Labels:  make(map[string]string),
	}

	if localMode {
		config.Entrypoint = EntrypointWeb
	} else {
		config.Entrypoint = EntrypointWebSecure
		config.TLS = &TLSConfig{
			CertResolver: "letsencrypt",
		}
		
		// Add HTTPS redirect middleware
		config.Middlewares = []string{"https-redirect"}
	}

	return config
}

// CreateServiceConfig creates a ServiceConfig for an application
func CreateServiceConfig(appName, port string) *ServiceConfig {
	return &ServiceConfig{
		Name: sanitizeServiceName(appName),
		Port: port,
	}
}

// AddHealthCheckLabels adds health check configuration to labels
func AddHealthCheckLabels(labels map[string]string, serviceName, healthPath string, interval, timeout string) {
	if healthPath != "" {
		labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.healthcheck.path", serviceName)] = healthPath
	}
	if interval != "" {
		labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.healthcheck.interval", serviceName)] = interval
	}
	if timeout != "" {
		labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.healthcheck.timeout", serviceName)] = timeout
	}
}

// AddCORSLabels adds CORS middleware configuration to labels
func AddCORSLabels(labels map[string]string, middlewareName string, origins []string, methods []string, headers []string) {
	middlewarePrefix := fmt.Sprintf("traefik.http.middlewares.%s.headers", middlewareName)
	
	if len(origins) > 0 {
		labels[middlewarePrefix+".accesscontrolalloworiginlist"] = strings.Join(origins, ",")
	}
	if len(methods) > 0 {
		labels[middlewarePrefix+".accesscontrolallowmethods"] = strings.Join(methods, ",")
	}
	if len(headers) > 0 {
		labels[middlewarePrefix+".accesscontrolallowheaders"] = strings.Join(headers, ",")
	}
	
	labels[middlewarePrefix+".accesscontrolallowcredentials"] = "true"
}

// AddRateLimitLabels adds rate limiting middleware configuration
func AddRateLimitLabels(labels map[string]string, middlewareName string, requests int, period string) {
	middlewarePrefix := fmt.Sprintf("traefik.http.middlewares.%s.ratelimit", middlewareName)
	
	labels[middlewarePrefix+".average"] = fmt.Sprintf("%d", requests)
	labels[middlewarePrefix+".period"] = period
	labels[middlewarePrefix+".burst"] = fmt.Sprintf("%d", requests*2) // Allow burst of 2x average
}

// AddCompressionLabels adds compression middleware configuration
func AddCompressionLabels(labels map[string]string, middlewareName string) {
	labels[fmt.Sprintf("traefik.http.middlewares.%s.compress", middlewareName)] = "true"
}

// ValidateDomain performs basic domain validation
func ValidateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	
	// Basic domain validation
	if strings.Contains(domain, " ") {
		return fmt.Errorf("domain cannot contain spaces")
	}
	
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return fmt.Errorf("domain cannot start or end with a dot")
	}
	
	if len(domain) > 253 {
		return fmt.Errorf("domain is too long (max 253 characters)")
	}
	
	return nil
}

// ValidatePort performs basic port validation
func ValidatePort(port string) error {
	if port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	
	// Additional port validation could be added here
	// For now, just check it's not empty
	return nil
}

// sanitizeRouterName sanitizes app name for use as Traefik router name
func sanitizeRouterName(appName string) string {
	// Replace invalid characters with hyphens
	sanitized := strings.ReplaceAll(appName, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ToLower(sanitized)
	
	// Remove any non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range sanitized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// sanitizeServiceName sanitizes app name for use as Traefik service name
func sanitizeServiceName(appName string) string {
	// Service names can be more flexible than router names
	return sanitizeRouterName(appName)
}