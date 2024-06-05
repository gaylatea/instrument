package instrument

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mattn/go-isatty"
)

type EventSink interface {
	Event(ctx context.Context, name string, t Tags) error
}

func PostEvent(ctx context.Context, name string, t Tags) {
	eventID, err := uuid.NewV7()
	if err != nil {
		Errorf(ctx, "cannot generate a new event ID: %v", err)
		eventID = uuid.Nil
	}

	t["event.id"] = eventID
	traceID := TraceIDFromContext(ctx)
	if traceID != uuid.Nil {
		t["event.trace.id"] = traceID
	}

	t["event.name"] = name
	t["meta.instance"] = globalUUID

	for sinkName, sink := range globalEventSinks {
		if err := sink.Event(ctx, name, t); err != nil {
			Errorf(ctx, "Could not process event sink '%s': %v", sinkName, err)
		}
	}
}

func UseEventSink(name string, s EventSink) {
	globalEventSinks[name] = s
}

type ConsoleEventSink struct{}

func (cs *ConsoleEventSink) Event(ctx context.Context, name string, t Tags) error {
	if isatty.IsTerminal(os.Stderr.Fd()) {
		return cs.logForHumans(ctx, name, t)
	}

	return cs.logAsJSON(ctx, name, t)
}

func (cs *ConsoleEventSink) logAsJSON(_ context.Context, _ string, t Tags) error {
	if final, err := json.Marshal(t); err != nil {
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
func (cs *ConsoleEventSink) logForHumans(ctx context.Context, name string, t Tags) error {
	color := levelToColor[INFO]

	metadata := fmt.Sprintf("[%s] %-33s %s", color.Render("EVT"), time.Now().Format(time.RFC3339Nano), name)

	var b strings.Builder
	b.WriteString(metadata)

	for k, v := range t {
		switch k {
		// Remove some tags that aren't useful in console output.
		case "event.name":
		case "event.id":
		case "event.trace.id":
		case "meta.instance":
		default:
			b.WriteString(fmt.Sprintf(" %s=%v", color.Render(k), v))
		}

	}

	_, err := os.Stderr.WriteString(prefixed(ctx, false, b.String()))
	return err
}
