package docker

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Client{
		cli: cli,
	}, nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) IsAvailable(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	if err != nil {
		return fmt.Errorf("docker daemon is not available: %w", err)
	}
	return nil
}

func (c *Client) PullImage(ctx context.Context, imageName string) error {
	reader, err := c.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Read the response to ensure the pull completes
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("failed to complete image pull for %s: %w", imageName, err)
	}

	return nil
}

func (c *Client) RunContainer(ctx context.Context, opts RunOptions) error {
	// Parse port mappings using Docker SDK utility
	var portBindings nat.PortMap
	var exposedPorts nat.PortSet

	if len(opts.Ports) > 0 {
		// Use Docker SDK's ParsePortSpecs for robust port parsing
		var portSpecs nat.PortSet
		var err error
		portSpecs, portBindings, err = nat.ParsePortSpecs(opts.Ports)
		if err != nil {
			return fmt.Errorf("invalid port specification: %w", err)
		}

		// Convert port specs to exposed ports
		exposedPorts = make(nat.PortSet)
		for port := range portSpecs {
			exposedPorts[port] = struct{}{}
		}
	}

	// Convert environment variables
	var env []string
	for key, value := range opts.EnvVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// Create container configuration
	config := &container.Config{
		Image:        opts.Image,
		Env:          env,
		ExposedPorts: exposedPorts,
		Labels:       opts.Labels,
	}

	// Set restart policy with default fallback
	restartPolicy := opts.RestartPolicy
	if restartPolicy == "" {
		restartPolicy = "unless-stopped" // Default policy
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyMode(restartPolicy),
		},
		Binds: opts.Volumes,
	}

	// Configure networks
	networkConfig := &network.NetworkingConfig{}
	if len(opts.Networks) > 0 {
		endpointsConfig := make(map[string]*network.EndpointSettings)
		for _, networkName := range opts.Networks {
			endpointsConfig[networkName] = &network.EndpointSettings{}
		}
		networkConfig.EndpointsConfig = endpointsConfig
	}

	resp, err := c.cli.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, opts.Name)
	if err != nil {
		return fmt.Errorf("failed to create container %s: %w", opts.Name, err)
	}

	// Cleanup on failure
	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		// If start fails, attempt to remove the created container to avoid orphaned containers
		if removeErr := c.cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}); removeErr != nil {
			return fmt.Errorf("failed to start container %s and failed to cleanup: start error: %w, cleanup error: %v", opts.Name, err, removeErr)
		}
		return fmt.Errorf("failed to start container %s: %w", opts.Name, err)
	}

	return nil
}

func (c *Client) StopContainer(ctx context.Context, name string) error {
	timeout := 30 // 30 seconds timeout
	options := container.StopOptions{
		Timeout: &timeout,
	}

	if err := c.cli.ContainerStop(ctx, name, options); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", name, err)
	}
	return nil
}

func (c *Client) StartContainer(ctx context.Context, name string) error {
	if err := c.cli.ContainerStart(ctx, name, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container %s: %w", name, err)
	}
	return nil
}

func (c *Client) RemoveContainer(ctx context.Context, name string, force bool) error {
	options := container.RemoveOptions{
		Force: force,
	}

	if err := c.cli.ContainerRemove(ctx, name, options); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", name, err)
	}
	return nil
}

func (c *Client) ListContainers(ctx context.Context) ([]Container, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]Container, 0, len(containers))
	for _, cont := range containers {
		// Get container name (remove leading slash)
		name := ""
		if len(cont.Names) > 0 {
			name = strings.TrimPrefix(cont.Names[0], "/")
		}

		// Format ports
		var ports []string
		for _, port := range cont.Ports {
			if port.PublicPort != 0 {
				ports = append(ports, fmt.Sprintf("%d:%d", port.PublicPort, port.PrivatePort))
			}
		}

		result = append(result, Container{
			Name:   name,
			Image:  cont.Image,
			Status: cont.Status,
			Ports:  strings.Join(ports, ", "),
		})
	}

	return result, nil
}

func (c *Client) ContainerExists(ctx context.Context, name string) (bool, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false, fmt.Errorf("failed to list containers: %w", err)
	}

	for _, cont := range containers {
		for _, containerName := range cont.Names {
			if strings.TrimPrefix(containerName, "/") == name {
				return true, nil
			}
		}
	}

	return false, nil
}

func (c *Client) GetContainerStatus(ctx context.Context, name string) (string, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	for _, cont := range containers {
		for _, containerName := range cont.Names {
			if strings.TrimPrefix(containerName, "/") == name {
				return cont.Status, nil
			}
		}
	}

	return "", fmt.Errorf("container %s not found", name)
}
