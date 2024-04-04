package styles

import "github.com/charmbracelet/lipgloss"

var ErrorStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FF0000")).
	Padding(0, 1)

var UsageStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FF00FF")).
	Bold(true).
	Padding(0, 1)
