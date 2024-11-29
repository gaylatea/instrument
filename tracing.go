package instrument

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const traceCallerSkip = 2

var (
	tracesTotal  Counter = "instrument.traces.total"
	tracesErrors Counter = "instrument.traces.errors"
)

// TraceFunc implementers run in the context of a trace.
type TraceFunc func(ctx context.Context, addToParent func(Tags)) error

// WithSpan runs a given function and emits trace-specific metadata.
func WithSpan(ctx context.Context, name string, traced TraceFunc) error {
	caller, filename, line := getCaller(traceCallerSkip)

	traceID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("cannot generate a new trace ID: %w", err)
	}

	parent := traceIDFromContext(ctx)
	newCtx := context.WithValue(ctx, keyTraceID, traceID)
	start := time.Now()
	wrappedErr := traced(newCtx, func(ts Tags) {
		newCtx = WithAll(newCtx, ts)
	})

	tracesTotal.Add()
	newTags := tagsFromContext(newCtx)
	if wrappedErr != nil {
		tracesErrors.Add()

		newTags["meta.level"] = ERROR
		newTags["trace.error"] = wrappedErr
	} else {
		newTags["meta.level"] = INFO
	}

	newTags["trace.name"] = name
	newTags["trace.start"] = start
	newTags["trace.duration.ms"] = time.Since(start).Milliseconds()
	newTags["meta.file"] = filename
	newTags["meta.line"] = line
	newTags["meta.caller"] = caller
	newTags["trace.id"] = traceID

	if parent != uuid.Nil {
		newTags["trace.parent"] = parent
	}

	emit(newCtx, newTags)

	return wrappedErr
}
