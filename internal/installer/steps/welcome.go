package steps

import (
	"github.com/charmbracelet/lipgloss"
)

var welcomeBox = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	Padding(0, 1)

var welcomeTitle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("62"))

// WelcomeView renders the welcome stage (no async work; done after delay from main model).
func WelcomeView() string {
	title := welcomeTitle.Render("Welcome to Finks")
	content := welcomeBox.Render(title)
	return "\n" + content + "\n\n"
}
