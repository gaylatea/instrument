package instrument

import (
	"context"
	"fmt"
	"maps"
	"time"
)

// PostEvent emits a user-created raw event. It only includes the given tags, and some important metadata.
func PostEvent(ctx context.Context, name string, t Tags) {
	caller, filename, line := getCaller(2)

	if t == nil {
		t = Tags{}
	}

	t["event.name"] = name
	t["meta.caller"] = caller
	t["meta.file"] = filename
	t["meta.line"] = line
	emit(ctx, t)
}

// emit fans out raw event data to the configured sinks.
func emit(ctx context.Context, t Tags) {
	if t == nil {
		t = Tags{}
	}

	t["meta.instance"] = instanceID
	t["meta.timestamp"] = time.Now()

	// Handle debug/trace messages here.
	if rawLevel, ok := t["meta.level"]; ok {
		switch rawLevel {
		case DEBUG:
			if !*debug {
				return
			}
		case TRACE:
			if !*trace {
				return
			}
		}
	}

	for sinkName, sink := range allSinks(ctx) {
		if err := sink.Event(ctx, t); err != nil {
			_ = terminal.Event(ctx, Tags{
				"meta.level": ERROR,
				"error":      fmt.Sprintf("could not process event sink '%s': %v", sinkName, err),
			})
		}
	}
}

// allSinks returns a merged view of global and context-specific sinks for an event.
func allSinks(ctx context.Context) sinks {
	s := maps.Clone(globalSinks)
	maps.Copy(s, sinksFromContext(ctx))

	return s
}
