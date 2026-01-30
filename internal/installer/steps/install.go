package steps

import (
	"github.com/charmbracelet/lipgloss"
)

var installDim = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

// InstallView renders the "Installing dependencies..." stage.
// When alreadyPresent is true, shows "Docker already present" instead.
func InstallView(alreadyPresent bool) string {
	var msg string
	if alreadyPresent {
		msg = installDim.Render("Docker already present.")
	} else {
		msg = installDim.Render("Installing dependencies...")
	}
	return "\n  " + msg + "\n\n"
}
