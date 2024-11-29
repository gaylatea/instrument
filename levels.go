package instrument

import "github.com/charmbracelet/lipgloss"

// Level represents a standard logging level.
type Level int

const (
	TRACE Level = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	METRIC
)

// String returns a short name for the level.
func (l Level) String() string {
	return levelToName[l]
}

// Style returns the configured lipgloss style for the level.
func (l Level) Style() *lipgloss.Style {
	return levelToColor[l]
}

// SetStyle changes the configured lipgloss style for the level.
func (l Level) SetStyle(s *lipgloss.Style) {
	levelToColor[l] = s
}

// newStyle returns a lipgloss style for the given hexadecimal color.
func newStyle(hex string) *lipgloss.Style {
	thisStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hex))

	return &thisStyle
}

var (
	levelToName = map[Level]string{
		TRACE:  "TRA",
		DEBUG:  "DBG",
		INFO:   "INF",
		WARN:   "WRN",
		ERROR:  "ERR",
		FATAL:  "FTL",
		METRIC: "MET",
	}

	levelToColor = map[Level]*lipgloss.Style{
		TRACE:  newStyle("#ff87e9"),
		DEBUG:  newStyle("#ad7fa8"),
		INFO:   newStyle("#34e2e2"),
		WARN:   newStyle("#fce94f"),
		ERROR:  newStyle("#ef2929"),
		FATAL:  newStyle("#ffa500"),
		METRIC: newStyle("#daf0ee"),
	}
)
