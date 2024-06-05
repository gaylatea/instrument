package instrument

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TraceFunc func(ctx context.Context, addToParent func(Tags)) error
type TraceSink interface {
	Trace(context.Context) error
}

func WithSpan(ctx context.Context, name string, fn TraceFunc) error {
	traceID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("cannot generate a new trace ID: %w", err)
	}

	newDepth := TraceDepthFromContext(ctx) + 1
	ctx = context.WithValue(ctx, keyTraceDepth, newDepth)

	parent := TraceIDFromContext(ctx)
	if parent != uuid.Nil {
		ctx = context.WithValue(ctx, keyTraceParent, parent)
	}

	ctx = context.WithValue(ctx, keyTraceID, traceID)
	ctx = context.WithValue(ctx, keyTraceName, name)
	start := time.Now()
	wrappedErr := fn(ctx, func(ts Tags) {
		ctx = WithAll(ctx, ts)
	})
	ctx = context.WithValue(ctx, keyTraceStart, start)
	ctx = context.WithValue(ctx, keyTraceDuration, time.Since(start))

	if wrappedErr != nil {
		ctx = context.WithValue(ctx, keyTraceError, wrappedErr)
	}

	for name, sink := range globalTraceSinks {
		if err := sink.Trace(ctx); err != nil {
			Errorf(ctx, "Could not process trace sink '%s': %v", name, err)
		}
	}

	ctxSinks := traceSinksFromContext(ctx)
	for name, sink := range ctxSinks {
		if err := sink.Trace(ctx); err != nil {
			Errorf(ctx, "Could not process trace sink '%s': %v", name, err)
		}
	}

	return wrappedErr
}

func UseTraceSink(name string, s TraceSink) {
	globalTraceSinks[name] = s
}

func WithTraceSink(ctx context.Context, name string, s TraceSink) context.Context {
	sinks := traceSinksFromContext(ctx)
	sinks[name] = s

	return context.WithValue(ctx, keyConfiguredTraceSinks, sinks)
}

func TraceIDFromContext(ctx context.Context) uuid.UUID {
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

func TraceParentFromContext(ctx context.Context) uuid.UUID {
	val := ctx.Value(keyTraceParent)
	if val == nil {
		return uuid.Nil
	}

	typed, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}

	return typed
}

func TraceDepthFromContext(ctx context.Context) int {
	val := ctx.Value(keyTraceDepth)
	if val == nil {
		return 0
	}

	typed, ok := val.(int)
	if !ok {
		return 0
	}

	return typed
}

func TraceErrorFromContext(ctx context.Context) error {
	val := ctx.Value(keyTraceError)
	if val == nil {
		return nil
	}

	typed, ok := val.(error)
	if !ok {
		return nil
	}

	return typed
}

func TraceDurationFromContext(ctx context.Context) time.Duration {
	val := ctx.Value(keyTraceDuration)
	if val == nil {
		return 0
	}

	typed, ok := val.(time.Duration)
	if !ok {
		return 0
	}

	return typed
}

func TraceStartFromContext(ctx context.Context) time.Time {
	val := ctx.Value(keyTraceStart)
	if val == nil {
		return time.Time{}
	}

	typed, ok := val.(time.Time)
	if !ok {
		return time.Time{}
	}

	return typed
}

func TraceNameFromContext(ctx context.Context) string {
	val := ctx.Value(keyTraceName)
	if val == nil {
		return "(unknown)"
	}

	typed, ok := val.(string)
	if !ok {
		return "(unknown)"
	}

	return typed
}

func traceSinksFromContext(ctx context.Context) traceSinks {
	val := ctx.Value(keyConfiguredTraceSinks)
	if val == nil {
		return traceSinks{}
	}

	typed, ok := val.(traceSinks)
	if !ok {
		return traceSinks{}
	}

	return typed
}
