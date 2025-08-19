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
	// Parse port mappings
	var portBindings nat.PortMap
	var exposedPorts nat.PortSet

	if opts.Port != "" {
		portBindings = make(nat.PortMap)
		exposedPorts = make(nat.PortSet)

		// Parse port format (e.g., "8080:80" or "8080:80/tcp")
		parts := strings.Split(opts.Port, ":")
		if len(parts) == 2 {
			hostPort := parts[0]
			containerPortStr := parts[1]

			// Handle protocol specification
			var proto string = "tcp"
			if strings.Contains(containerPortStr, "/") {
				protoParts := strings.Split(containerPortStr, "/")
				containerPortStr = protoParts[0]
				if len(protoParts) > 1 {
					proto = protoParts[1]
				}
			}

			containerPort, err := nat.NewPort(proto, containerPortStr)
			if err != nil {
				return fmt.Errorf("invalid container port %s: %w", containerPortStr, err)
			}

			exposedPorts[containerPort] = struct{}{}
			portBindings[containerPort] = []nat.PortBinding{
				{
					HostPort: hostPort,
				},
			}
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
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		Binds: opts.Volumes,
	}

	networkConfig := &network.NetworkingConfig{}

	// Create container
	resp, err := c.cli.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, opts.Name)
	if err != nil {
		return fmt.Errorf("failed to create container %s: %w", opts.Name, err)
	}

	// Start container
	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
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
