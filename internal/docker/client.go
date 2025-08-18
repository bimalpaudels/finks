package docker

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var (
	ErrDockerNotFound = errors.New("docker command not found - please install Docker")
	ErrDockerNotRunning = errors.New("docker daemon is not running - please start Docker")
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) IsAvailable(ctx context.Context) error {
	_, err := exec.LookPath("docker")
	if err != nil {
		return ErrDockerNotFound
	}

	cmd := exec.CommandContext(ctx, "docker", "version", "--format", "{{.Server.Version}}")
	if err := cmd.Run(); err != nil {
		return ErrDockerNotRunning
	}

	return nil
}

func (c *Client) PullImage(ctx context.Context, image string) error {
	cmd := exec.CommandContext(ctx, "docker", "pull", image)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w\nOutput: %s", image, err, string(output))
	}
	return nil
}

func (c *Client) RunContainer(ctx context.Context, opts RunOptions) error {
	args := []string{"run", "-d"}
	
	args = append(args, "--name", opts.Name)
	args = append(args, "--restart", "unless-stopped")
	
	if opts.Port != "" {
		args = append(args, "-p", opts.Port)
	}
	
	for key, value := range opts.EnvVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}
	
	for _, volume := range opts.Volumes {
		args = append(args, "-v", volume)
	}
	
	args = append(args, opts.Image)
	
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run container %s: %w\nOutput: %s", opts.Name, err, string(output))
	}
	
	return nil
}

func (c *Client) StopContainer(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "docker", "stop", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop container %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

func (c *Client) StartContainer(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "docker", "start", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start container %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

func (c *Client) RemoveContainer(ctx context.Context, name string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)
	
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove container %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

func (c *Client) ListContainers(ctx context.Context) ([]Container, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--format", "{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	containers := make([]Container, 0, len(lines))
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		
		container := Container{
			Name:   parts[0],
			Image:  parts[1],
			Status: parts[2],
		}
		if len(parts) >= 4 {
			container.Ports = parts[3]
		}
		
		containers = append(containers, container)
	}
	
	return containers, nil
}

func (c *Client) ContainerExists(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", fmt.Sprintf("name=^%s$", name), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check if container exists: %w", err)
	}
	
	return strings.TrimSpace(string(output)) == name, nil
}

func (c *Client) GetContainerStatus(ctx context.Context, name string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", fmt.Sprintf("name=^%s$", name), "--format", "{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get container status: %w", err)
	}
	
	status := strings.TrimSpace(string(output))
	if status == "" {
		return "", fmt.Errorf("container %s not found", name)
	}
	
	return status, nil
}