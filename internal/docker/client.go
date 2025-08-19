package docker

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var (
	ErrDockerNotFound   = errors.New("docker command not found - please install Docker")
	ErrDockerNotRunning = errors.New("docker daemon is not running - please start Docker")
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &Client{cli: cli}, nil
}

func (c *Client) IsAvailable(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	if err != nil {
		return ErrDockerNotRunning
	}
	return nil
}

func (c *Client) PullImage(ctx context.Context, imageName string) error {
	_, err := c.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	return nil
}

func (c *Client) RunContainer(ctx context.Context, opts RunOptions) error {
	config := &container.Config{
		Image: opts.Image,
		Env:   make([]string, 0, len(opts.EnvVars)),
	}
	
	for key, value := range opts.EnvVars {
		config.Env = append(config.Env, fmt.Sprintf("%s=%s", key, value))
	}
	
	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
		Binds:         opts.Volumes,
	}
	
	if opts.Port != "" {
		parts := strings.Split(opts.Port, ":")
		if len(parts) == 2 {
			hostConfig.PortBindings = nat.PortMap{
				nat.Port(parts[1] + "/tcp"): []nat.PortBinding{{HostPort: parts[0]}},
			}
		}
	}
	
	_, err := c.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, opts.Name)
	if err != nil {
		return fmt.Errorf("failed to create container %s: %w", opts.Name, err)
	}
	
	if err := c.cli.ContainerStart(ctx, opts.Name, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container %s: %w", opts.Name, err)
	}
	
	return nil
}

func (c *Client) StopContainer(ctx context.Context, name string) error {
	if err := c.cli.ContainerStop(ctx, name, container.StopOptions{}); err != nil {
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
	options := container.RemoveOptions{Force: force}
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
	for _, container := range containers {
		name := ""
		if len(container.Names) > 0 {
			name = strings.TrimPrefix(container.Names[0], "/")
		}
		
		ports := ""
		if len(container.Ports) > 0 {
			var portStrings []string
			for _, port := range container.Ports {
				if port.PublicPort != 0 {
					portStrings = append(portStrings, fmt.Sprintf("%d:%d", port.PublicPort, port.PrivatePort))
				}
			}
			ports = strings.Join(portStrings, ", ")
		}
		
		result = append(result, Container{
			Name:   name,
			Image:  container.Image,
			Status: container.Status,
			Ports:  ports,
		})
	}
	
	return result, nil
}

func (c *Client) ContainerExists(ctx context.Context, name string) (bool, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("name", fmt.Sprintf("^%s$", name))
	
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check if container exists: %w", err)
	}
	
	return len(containers) > 0, nil
}

func (c *Client) GetContainerStatus(ctx context.Context, name string) (string, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("name", fmt.Sprintf("^%s$", name))
	
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get container status: %w", err)
	}
	
	if len(containers) == 0 {
		return "", fmt.Errorf("container %s not found", name)
	}
	
	return containers[0].Status, nil
}