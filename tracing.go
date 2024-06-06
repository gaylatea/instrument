package instrument

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TraceFunc implementers run in the context of a trace.
type TraceFunc func(ctx context.Context, addToParent func(Tags)) error

// WithSpan runs a given function and emits trace-specific metadata.
func WithSpan(ctx context.Context, name string, fn TraceFunc) error {
	caller, filename, line := getCaller(2)

	traceID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("cannot generate a new trace ID: %w", err)
	}

	parent := traceIDFromContext(ctx)
	newCtx := context.WithValue(ctx, keyTraceID, traceID)
	start := time.Now()
	wrappedErr := fn(newCtx, func(ts Tags) {
		newCtx = WithAll(newCtx, ts)
	})

	t := tagsFromContext(newCtx)
	if wrappedErr != nil {
		t["meta.level"] = ERROR
		t["trace.error"] = wrappedErr
	} else {
		t["meta.level"] = INFO
	}

	t["trace.id"] = traceID
	if parent != uuid.Nil {
		t["trace.parent"] = parent
	}
	t["trace.name"] = name
	t["trace.start"] = start
	t["trace.duration"] = time.Since(start)
	t["meta.file"] = filename
	t["meta.line"] = line
	t["meta.caller"] = caller

	emit(newCtx, t)

	return wrappedErr
}
