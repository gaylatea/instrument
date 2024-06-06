package instrument

import "github.com/charmbracelet/lipgloss"

// Level represents a standard logging level.
type Level int

const (
	// Extremely fine-grained logs.
	TRACE Level = iota
	// Logs that would help fix potential issues, but are too verbose by default.
	DEBUG
	// Logs emitted by default.
	INFO
	// Logs that indicate a non-critical issue.
	WARN
	// Logs that indicate an issue that stops the current module but shouldn't stop all processes.
	ERROR
	// Logs that indicate a critical error that must stop all processes.
	FATAL
)

func (l Level) String() string {
	return levelToName[l]
}

func (l Level) Style() lipgloss.Style {
	return levelToColor[l]
}

func (l Level) SetStyle(s lipgloss.Style) {
	levelToColor[l] = s
}

var (
	// levelToName stores a mapping of level const to a console-friendly name.
	levelToName = map[Level]string{
		TRACE: "TRA",
		DEBUG: "DBG",
		INFO:  "INF",
		WARN:  "WRN",
		ERROR: "ERR",
		FATAL: "FTL",
	}

	// levelToColor stores a mapping of level const to lipgloss style.
	levelToColor = map[Level]lipgloss.Style{
		TRACE: lipgloss.NewStyle().Foreground(lipgloss.Color("#8ae234")),
		DEBUG: lipgloss.NewStyle().Foreground(lipgloss.Color("#ad7fa8")),
		INFO:  lipgloss.NewStyle().Foreground(lipgloss.Color("#34e2e2")),
		WARN:  lipgloss.NewStyle().Foreground(lipgloss.Color("#fce94f")).Bold(true),
		ERROR: lipgloss.NewStyle().Foreground(lipgloss.Color("#ef2929")).Bold(true),
		FATAL: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true),
	}
)
