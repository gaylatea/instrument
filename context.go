package instrument

import (
	"context"
	"maps"

	"github.com/google/uuid"
)

// sinksFromContext returns any configured event sinks for the given context.
func sinksFromContext(ctx context.Context) sinks {
	val := ctx.Value(keyConfiguredSinks)
	if val == nil {
		return sinks{}
	}

	typed, ok := val.(sinks)
	if !ok {
		return sinks{}
	}

	return maps.Clone(typed)
}

// tagsFromContext returns any configured tags for the given context.
func tagsFromContext(ctx context.Context) Tags {
	val := ctx.Value(keyTags)
	if val == nil {
		return Tags{}
	}

	typed, ok := val.(Tags)
	if !ok {
		return Tags{}
	}

	return maps.Clone(typed)
}

// traceIDFromContext returns the current trace for the given context.
func traceIDFromContext(ctx context.Context) uuid.UUID {
	val := ctx.Value(keyTraceID)
	if val == nil {
		return uuid.Nil
	}

	typed, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}

	return typed
}
