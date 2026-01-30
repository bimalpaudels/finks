package checker

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bimalpaudels/finks/internal/docker"
)

// Result holds the outcome of a requirement check.
type Result struct {
	Name    string
	OK      bool
	Message string
	Err     error
}

// CheckResultMsg is the aggregate result of checking all requirements (used by installer).
type CheckResultMsg struct {
	DockerOK bool
	Docker   Result
	Err      error
}

// VerifyResultMsg is the result of verifying dependencies (e.g. Docker daemon ping).
type VerifyResultMsg struct {
	DockerOK bool
	Err      error
}

// InstallDoneMsg signals that the install step has finished.
type InstallDoneMsg struct {
	Installed bool // true if we ran an installer, false if already present or skipped
	Err       error
}

// DockerRequirement checks for Docker CLI and optionally daemon availability.
type DockerRequirement struct {
	client *docker.Client
}

// NewDockerRequirement creates a Docker requirement checker.
func NewDockerRequirement() (*DockerRequirement, error) {
	client, err := docker.NewClient()
	if err != nil {
		return nil, err
	}
	return &DockerRequirement{client: client}, nil
}

// Close releases resources used by the checker.
func (d *DockerRequirement) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

// Name returns the requirement name.
func (d *DockerRequirement) Name() string {
	return "Docker"
}

// Check verifies Docker CLI is present and optionally that the daemon is reachable.
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
		return Result{Name: "Docker", OK: true, Message: "Docker CLI found"}
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

// CheckCLIOnly verifies only that the Docker CLI binary is present (no daemon).
func (d *DockerRequirement) CheckCLIOnly() Result {
	if _, err := exec.LookPath("docker"); err != nil {
		return Result{
			Name:    "Docker",
			OK:      false,
			Message: "Docker CLI not found in PATH",
			Err:     err,
		}
	}
	return Result{Name: "Docker", OK: true, Message: "Docker CLI found"}
}

// Verify pings the Docker daemon to confirm it is ready.
func (d *DockerRequirement) Verify(ctx context.Context) error {
	if d.client == nil {
		return fmt.Errorf("docker client not initialized")
	}
	return d.client.IsAvailable(ctx)
}

// Install runs the Docker convenience script on Linux; on other OSes it returns instructions.
func (d *DockerRequirement) Install(ctx context.Context) (installed bool, err error) {
	if runtime.GOOS != "linux" {
		return false, fmt.Errorf("automatic Docker install is only supported on Linux; see https://docs.docker.com/get-docker/")
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", "curl -fsSL https://get.docker.com | sh")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("Docker install script failed: %w", err)
	}
	return true, nil
}

// CheckDocker returns a Bubble Tea Cmd that runs the full Docker check (CLI + daemon) and sends CheckResultMsg.
func CheckDocker() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req, err := NewDockerRequirement()
		if err != nil {
			return CheckResultMsg{DockerOK: false, Err: err}
		}
		defer req.Close()
		res := req.Check(ctx)
		return CheckResultMsg{
			DockerOK: res.OK,
			Docker:   res,
			Err:      res.Err,
		}
	}
}

// VerifyDocker returns a Bubble Tea Cmd that pings the Docker daemon and sends VerifyResultMsg.
func VerifyDocker() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req, err := NewDockerRequirement()
		if err != nil {
			return VerifyResultMsg{DockerOK: false, Err: err}
		}
		defer req.Close()
		if err := req.Verify(ctx); err != nil {
			return VerifyResultMsg{DockerOK: false, Err: err}
		}
		return VerifyResultMsg{DockerOK: true}
	}
}

// RunInstallStep returns a Bubble Tea Cmd for the install step. If dockerOK is true
// it immediately returns InstallDoneMsg{Installed: false}. Otherwise on Linux it
// runs the Docker convenience script; on other OSes it returns an error with instructions.
func RunInstallStep(dockerOK bool) tea.Cmd {
	return func() tea.Msg {
		if dockerOK {
			return InstallDoneMsg{Installed: false}
		}
		if runtime.GOOS != "linux" {
			return InstallDoneMsg{
				Installed: false,
				Err:       fmt.Errorf("automatic Docker install is only supported on Linux; see https://docs.docker.com/get-docker/"),
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "sh", "-c", "curl -fsSL https://get.docker.com | sh")
		if err := cmd.Run(); err != nil {
			return InstallDoneMsg{Installed: false, Err: fmt.Errorf("Docker install script failed: %w", err)}
		}
		return InstallDoneMsg{Installed: true}
	}
}
