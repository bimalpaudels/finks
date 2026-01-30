package installer

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Box is the border style for welcome and content boxes.
	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	// Title is the style for the welcome title.
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62"))

	// Success is used for checkmarks and success messages.
	Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color("78"))

	// Error is used for failure messages and status.
	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("203"))

	// Dim is used for secondary text.
	Dim = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// StatusOK is a short success indicator (e.g. "✓").
	StatusOK = Success.Render("✓")

	// StatusFail is a short failure indicator (e.g. "✗").
	StatusFail = Error.Render("✗")
)
