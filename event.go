package instrument

import (
	"context"
	"fmt"
	"maps"
	"time"
)

const eventCallerSkip = 2

// PostEvent emits a user-created raw event without contextual metadata.
func PostEvent(ctx context.Context, name string, givenTags Tags) {
	caller, filename, line := getCaller(eventCallerSkip)

	if givenTags == nil {
		givenTags = Tags{}
	}

	givenTags["event.name"] = name
	givenTags["meta.caller"] = caller
	givenTags["meta.file"] = filename
	givenTags["meta.line"] = line
	emit(ctx, givenTags)
}

// emit fans out raw event data to the configured sinks.
func emit(ctx context.Context, givenTags Tags) {
	if givenTags == nil {
		givenTags = Tags{}
	}

	givenTags["meta.instance"] = instanceID
	givenTags["meta.timestamp"] = time.Now()

	// Handle debug/trace messages here.
	if rawLevel, ok := givenTags["meta.level"]; ok {
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
		if err := sink.Event(ctx, givenTags); err != nil {
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
