package instrument

import (
	"context"
	"fmt"
	"os"
)

// TerminalSink emits events to the terminal with optional colors.
type TerminalSink struct{}

// Event writes colorized JSON to the terminal for debugging.
func (cs *TerminalSink) Event(_ context.Context, t Tags) error {
	if *silent {
		return nil
	}

	keyColor := levelToColor[INFO]
	if rawLevel, ok := t["meta.level"]; ok {
		level, ok := rawLevel.(Level)
		if ok {
			keyColor = levelToColor[level]
		}
	}

	if final, err := marshal(t, keyColor); err != nil {
		return fmt.Errorf("could not create JSON output: %w", err)
	} else {
		if _, err := os.Stderr.Write(final); err != nil {
			return fmt.Errorf("could not emit log: %w", err)
		}

		if _, err := os.Stderr.WriteString("\n"); err != nil {
			return fmt.Errorf("could not add newline: %w", err)
		}
	}

	return nil
}
