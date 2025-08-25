package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/network"
)



func (c *Client) CreateNetwork(ctx context.Context, name, driver string, labels map[string]string) (string, error) {
	options := network.CreateOptions{
		Driver: driver,
		Labels: labels,
	}

	resp, err := c.cli.NetworkCreate(ctx, name, options)
	if err != nil {
		return "", fmt.Errorf("failed to create network %s: %w", name, err)
	}

	return resp.ID, nil
}

func (c *Client) NetworkExists(ctx context.Context, name string) (bool, error) {
	networks, err := c.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list networks: %w", err)
	}

	for _, net := range networks {
		if net.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (c *Client) GetNetworkInfo(ctx context.Context, name string) (*NetworkInfo, error) {
	resp, err := c.cli.NetworkInspect(ctx, name, network.InspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect network %s: %w", name, err)
	}

	info := &NetworkInfo{
		ID:     resp.ID,
		Name:   resp.Name,
		Driver: resp.Driver,
		Labels: resp.Labels,
	}

	// Extract subnet and gateway from IPAM config
	if len(resp.IPAM.Config) > 0 {
		config := resp.IPAM.Config[0]
		info.Subnet = config.Subnet
		info.Gateway = config.Gateway
	}

	return info, nil
}

func (c *Client) ConnectContainerToNetwork(ctx context.Context, networkName, containerName string) error {
	err := c.cli.NetworkConnect(ctx, networkName, containerName, nil)
	if err != nil {
		return fmt.Errorf("failed to connect container %s to network %s: %w", containerName, networkName, err)
	}

	return nil
}

func (c *Client) DisconnectContainerFromNetwork(ctx context.Context, networkName, containerName string) error {
	err := c.cli.NetworkDisconnect(ctx, networkName, containerName, false)
	if err != nil {
		return fmt.Errorf("failed to disconnect container %s from network %s: %w", containerName, networkName, err)
	}

	return nil
}

func (c *Client) ListNetworks(ctx context.Context) ([]NetworkInfo, error) {
	networks, err := c.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	result := make([]NetworkInfo, 0, len(networks))
	for _, net := range networks {
		info := NetworkInfo{
			ID:     net.ID,
			Name:   net.Name,
			Driver: net.Driver,
			Labels: net.Labels,
		}

		// Extract subnet and gateway from IPAM config
		if len(net.IPAM.Config) > 0 {
			config := net.IPAM.Config[0]
			info.Subnet = config.Subnet
			info.Gateway = config.Gateway
		}

		result = append(result, info)
	}

	return result, nil
}

func (c *Client) RemoveNetwork(ctx context.Context, name string) error {
	err := c.cli.NetworkRemove(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to remove network %s: %w", name, err)
	}

	return nil
}

// EnsureNetwork ensures a network exists, creating it if necessary
func (c *Client) EnsureNetwork(ctx context.Context, name, driver string, labels map[string]string) (string, error) {
	exists, err := c.NetworkExists(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to check if network exists: %w", err)
	}

	if exists {
		// Get existing network info to return ID
		info, err := c.GetNetworkInfo(ctx, name)
		if err != nil {
			return "", fmt.Errorf("failed to get existing network info: %w", err)
		}
		return info.ID, nil
	}

	// Create network
	networkID, err := c.CreateNetwork(ctx, name, driver, labels)
	if err != nil {
		return "", fmt.Errorf("failed to create network: %w", err)
	}

	return networkID, nil
}