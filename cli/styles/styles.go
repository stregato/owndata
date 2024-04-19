package styles

import "github.com/charmbracelet/lipgloss"

var ErrorStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FF0000")).
	Padding(1, 1)

var UsageStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FF00FF")).
	Bold(true).
	PaddingLeft(2)

var UseStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#000000")).
	Bold(true).
	PaddingLeft(4).
	Width(20)

var ShortStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#000000"))
