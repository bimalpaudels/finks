package requirements

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/bimalpaudels/finks/internal/docker"
)

// Result holds the outcome of a requirement check.
type Result struct {
	Name    string
	OK      bool
	Message string
	Err     error
}

// Requirement defines the interface for a dependency that can be checked, installed, and verified.
type Requirement interface {
	// Name returns the human-readable name of the requirement.
	Name() string

	// Check verifies if the requirement is satisfied.
	Check(ctx context.Context) Result

	// InstallCommand returns the command string that will be run to install the requirement.
	// On non-Linux systems, this may return a URL or instructions instead.
	InstallCommand() string

	// CanAutoInstall returns true if automatic installation is supported on this OS.
	CanAutoInstall() bool

	// Install runs the installation process for this requirement.
	Install(ctx context.Context) (installed bool, err error)

	// Verify confirms the requirement is working after installation.
	Verify(ctx context.Context) error

	// Close releases any resources held by the requirement.
	Close() error
}

// DockerRequirement implements Requirement for Docker.
type DockerRequirement struct {
	client *docker.Client
}

// NewDockerRequirement creates a new Docker requirement checker.
func NewDockerRequirement() (*DockerRequirement, error) {
	client, err := docker.NewClient()
	if err != nil {
		// If we can't create a client, we can still check for CLI
		return &DockerRequirement{client: nil}, nil
	}
	return &DockerRequirement{client: client}, nil
}

// Name returns the requirement name.
func (d *DockerRequirement) Name() string {
	return "Docker"
}

// Check verifies Docker CLI is present and that the daemon is reachable.
func (d *DockerRequirement) Check(ctx context.Context) Result {
	if _, err := exec.LookPath("docker"); err != nil {
		return Result{
			Name:    "Docker",
			OK:      false,
			Message: "Docker CLI not found in PATH",
			Err:     err,
		}
	}
	if d.client == nil {
		// Try to create a client to check daemon
		client, err := docker.NewClient()
		if err != nil {
			return Result{
				Name:    "Docker",
				OK:      false,
				Message: "Docker daemon not available",
				Err:     err,
			}
		}
		d.client = client
	}
	if err := d.client.IsAvailable(ctx); err != nil {
		return Result{
			Name:    "Docker",
			OK:      false,
			Message: "Docker daemon not available",
			Err:     err,
		}
	}
	return Result{Name: "Docker", OK: true, Message: "Docker CLI and daemon ready"}
}

// InstallCommand returns the command used to install Docker.
func (d *DockerRequirement) InstallCommand() string {
	if runtime.GOOS == "linux" {
		return "curl -fsSL https://get.docker.com | sh"
	}
	return "https://docs.docker.com/get-docker/"
}

// CanAutoInstall returns true if automatic installation is supported.
func (d *DockerRequirement) CanAutoInstall() bool {
	return runtime.GOOS == "linux"
}

// Install runs the Docker convenience script on Linux.
func (d *DockerRequirement) Install(ctx context.Context) (installed bool, err error) {
	if runtime.GOOS != "linux" {
		return false, fmt.Errorf("automatic Docker install is only supported on Linux; see https://docs.docker.com/get-docker/")
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", "curl -fsSL https://get.docker.com | sh")
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("Docker install script failed: %w", err)
	}
	return true, nil
}

// Verify pings the Docker daemon to confirm it is ready.
func (d *DockerRequirement) Verify(ctx context.Context) error {
	if d.client == nil {
		client, err := docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create docker client: %w", err)
		}
		d.client = client
	}
	return d.client.IsAvailable(ctx)
}

// Close releases resources used by the requirement.
func (d *DockerRequirement) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

// AllRequirements returns the ordered list of all requirements to check and install.
// Add new requirements here to include them in the installation wizard.
func AllRequirements() []Requirement {
	docker, _ := NewDockerRequirement()
	return []Requirement{
		docker,
	}
}

// CheckRequirement is a Bubble Tea message for the result of checking a requirement.
type CheckRequirementMsg struct {
	Index  int
	Result Result
}

// InstallRequirementMsg is a Bubble Tea message for the result of installing a requirement.
type InstallRequirementMsg struct {
	Index     int
	Installed bool
	Err       error
}

// VerifyRequirementMsg is a Bubble Tea message for the result of verifying a requirement.
type VerifyRequirementMsg struct {
	Index int
	OK    bool
	Err   error
}

// CheckRequirementCmd returns a Bubble Tea Cmd that checks the requirement at the given index.
func CheckRequirementCmd(req Requirement, index int) func() interface{} {
	return func() interface{} {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result := req.Check(ctx)
		return CheckRequirementMsg{
			Index:  index,
			Result: result,
		}
	}
}

// InstallRequirementCmd returns a Bubble Tea Cmd that installs the requirement at the given index.
func InstallRequirementCmd(req Requirement, index int) func() interface{} {
	return func() interface{} {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		installed, err := req.Install(ctx)
		return InstallRequirementMsg{
			Index:     index,
			Installed: installed,
			Err:       err,
		}
	}
}

// VerifyRequirementCmd returns a Bubble Tea Cmd that verifies the requirement at the given index.
func VerifyRequirementCmd(req Requirement, index int) func() interface{} {
	return func() interface{} {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := req.Verify(ctx)
		return VerifyRequirementMsg{
			Index: index,
			OK:    err == nil,
			Err:   err,
		}
	}
}
