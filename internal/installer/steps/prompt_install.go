package steps

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	promptTitle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	promptCommand = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Background(lipgloss.Color("236")).Padding(0, 1)
	promptHint    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	promptAction  = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
)

// PromptInstallView renders the "Press Enter to install" prompt for a requirement.
// name: requirement name (e.g. "Docker")
// command: the install command or URL
// canAutoInstall: true if automatic install is supported on this OS
func PromptInstallView(name, command string, canAutoInstall bool) string {
	title := promptTitle.Render(name + " needs to be installed.")

	var body string
	if canAutoInstall {
		action := promptAction.Render("Press Enter to continue.")
		cmd := promptCommand.Render(command)
		hint := promptHint.Render("You may need to enter your sudo password if you don't have permission.")
		body = "\n\n  " + action + "\n\n  Command: " + cmd + "\n\n  " + hint
	} else {
		hint := promptHint.Render("Please install " + name + " manually:")
		link := promptCommand.Render(command)
		action := promptAction.Render("Press Enter after installing to continue.")
		body = "\n\n  " + hint + "\n\n  " + link + "\n\n  " + action
	}

	return "\n  " + title + body + "\n\n"
}
