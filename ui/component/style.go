package component

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// ANSIBlack
// ANSIRed
// ANSIGreen
// ANSIYellow
// ANSIBlue
// ANSIMagenta
// ANSICyan
// ANSIWhite
// ANSIBrightBlack
// ANSIBrightRed
// ANSIBrightGreen
// ANSIBrightYellow
// ANSIBrightBlue
// ANSIBrightMagenta
// ANSIBrightCyan
// ANSIBrightWhite

var (
	StyleSelectedWorkConfig = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIBrightBlue))
	StyleCompletedItem      = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIBrightGreen))
	StyleActiveItem         = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIBrightYellow))
	StyleTodo               = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIYellow))
	StyleFinish             = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIGreen))
	StyleBreaking           = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIBrightGreen))
	StyleWorking            = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(termenv.ANSIMagenta))
)
