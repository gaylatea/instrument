package instrument

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/mattn/go-isatty"
)

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

// ConsoleSink dumps out events to the console.
type ConsoleLogSink struct{}

func (cs *ConsoleLogSink) Log(ctx context.Context, l Level, filename string, line int, msg string, args ...interface{}) error {
	msg = fmt.Sprintf(msg, args...)

	if isatty.IsTerminal(os.Stderr.Fd()) {
		return cs.logForHumans(ctx, l, filename, line, msg)
	}

	return cs.logAsJSON(ctx, l, filename, line, msg)
}

func (cs *ConsoleLogSink) logAsJSON(ctx context.Context, l Level, filename string, line int, msg string) error {
	output := map[string]any{}

	// Add the tags first so that they can't override builtins.
	// SAFETY: skipping the error-handling here because this won't error out at runtime.
	_ = EachTag(ctx, func(k string, v any) error {
		output[k] = v
		return nil
	})

	traceID := TraceIDFromContext(ctx)
	if traceID != uuid.Nil {
		output["log.trace.id"] = traceID
	}

	output["log.meta.level"] = levelToName[l]
	output["log.meta.filename"] = filename
	output["log.meta.line"] = line
	output["log.message"] = msg

	if final, err := json.Marshal(output); err != nil {
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

// logForHumans emits a formatted text string with colorized tags.
func (cs *ConsoleLogSink) logForHumans(ctx context.Context, l Level, filename string, line int, msg string) error {
	color := levelToColor[l]
	levelName := levelToName[l]

	metadata := fmt.Sprintf("%s:%d %-33s", color.Render(filename), line, time.Now().Format(time.RFC3339Nano))

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[%s] %-50s ", color.Render(levelName), metadata))
	b.WriteString(msg)

	// SAFETY: skipping the error-handling here because this won't error out at runtime.
	_ = EachTag(ctx, func(k string, v interface{}) error {
		// Skip the instance ID since it just clutters up the output.
		if k == "meta.instance" {
			return nil
		}

		b.WriteString(fmt.Sprintf(" %s=%v", color.Render(k), v))
		return nil
	})

	_, err := os.Stderr.WriteString(prefixed(ctx, false, b.String()))
	return err
}

type ConsoleTraceSink struct{}

func (ct *ConsoleTraceSink) Trace(ctx context.Context) error {
	if isatty.IsTerminal(os.Stderr.Fd()) {
		return ct.logForHumans(ctx)
	}

	return ct.logAsJSON(ctx)
}

func (ct *ConsoleTraceSink) logAsJSON(ctx context.Context) error {
	output := map[string]any{}

	// Add the tags first so that they can't override builtins.
	// SAFETY: skipping the error-handling here because this won't error out at runtime.
	_ = EachTag(ctx, func(k string, v any) error {
		output[k] = v
		return nil
	})

	output["trace.id"] = TraceIDFromContext(ctx)
	parent := TraceParentFromContext(ctx)
	if parent != uuid.Nil {
		output["trace.parent"] = parent
	}
	traceErr := TraceErrorFromContext(ctx)
	if traceErr != nil {
		output["trace.error"] = traceErr.Error()
	}
	output["trace.name"] = TraceNameFromContext(ctx)
	output["trace.start"] = TraceStartFromContext(ctx)
	output["trace.duration"] = TraceDurationFromContext(ctx)

	if final, err := json.Marshal(output); err != nil {
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

// logForHumans emits a formatted text string with colorized tags.
func (ct *ConsoleTraceSink) logForHumans(ctx context.Context) error {
	traceColor := levelToColor[INFO]
	traceErr := TraceErrorFromContext(ctx)
	if traceErr != nil {
		traceColor = levelToColor[ERROR]
	}
	traceName := TraceNameFromContext(ctx)
	traceDuration := TraceDurationFromContext(ctx)

	msg := prefixed(ctx, true, fmt.Sprintf("%s (%s)", traceColor.Render(traceName), traceDuration))

	_, err := os.Stderr.WriteString(msg)
	return err
}

func prefixed(ctx context.Context, isTrace bool, msg string) string {
	var b bytes.Buffer

	depth := TraceDepthFromContext(ctx)
	if depth != 0 {
		if isTrace {
			depth--
		}
		for range depth {
			b.WriteString("  ┃")
		}
		if isTrace {
			b.WriteString("  ┗")
		}
	}

	b.WriteString(fmt.Sprintf(" %s\n", msg))

	return b.String()
}
