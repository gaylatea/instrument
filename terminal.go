package instrument

import (
	"context"
	"fmt"
	"os"
)

// TerminalSink emits events to the terminal with optional colors.
type TerminalSink struct{}

// Event writes colorized JSON to the terminal for debugging.
func (cs *TerminalSink) Event(_ context.Context, givenTags Tags) error {
	if *silent {
		return nil
	}

	keyColor := levelToColor[INFO]

	if rawLevel, ok := givenTags["meta.level"]; ok {
		level, ok := rawLevel.(Level)
		if ok {
			keyColor = levelToColor[level]
		}
	}

	final := marshal(givenTags, keyColor)
	if _, err := os.Stderr.Write(final); err != nil {
		return fmt.Errorf("could not emit log: %w", err)
	}

	if _, err := os.Stderr.WriteString("\n"); err != nil {
		return fmt.Errorf("could not add newline: %w", err)
	}

	return nil
}
