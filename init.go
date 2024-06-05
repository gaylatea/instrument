package instrument

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
)

type Tags map[string]any
type tagOrder []string
type logSinks map[string]LogSink
type traceSinks map[string]TraceSink
type eventSinks map[string]EventSink

type contextKey int

const (
	keyTags contextKey = iota
	keyOrder

	keyConfiguredLogSinks
	keyConfiguredMetricSinks
	keyConfiguredTraceSinks

	keyTraceID
	keyTraceParent
	keyTraceDepth
	keyTraceName
	keyTraceError
	keyTraceDuration
	keyTraceStart
)

var (
	debug = flag.Bool("debug", false, "Enable debug logging.")
	trace = flag.Bool("trace", false, "Enable trace logging. EXTREMELY VERBOSE.")

	// The logging context always includes a random UUID for this particular invocation of this program.
	globalUUID uuid.UUID

	globalLogSinks = logSinks{
		"console": &ConsoleLogSink{},
	}

	globalTraceSinks = traceSinks{
		"console": &ConsoleTraceSink{},
	}

	globalEventSinks = eventSinks{
		"console": &ConsoleEventSink{},
	}
)

func init() {
	if id, err := uuid.NewV7(); err != nil {
		Fatalf(context.Background(), "Could not create a unique ID for this instance: %v", err)
	} else {
		globalUUID = id
	}

	// Startup a signal handler to allow switching log levels at runtime.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGUSR1)
		for {
			switch <-c {
			case syscall.SIGHUP:
				SetDebug(!*debug)
			case syscall.SIGUSR1:
				SetTrace(!*trace)
			}
		}
	}()
}

func TagsFromContext(ctx context.Context) Tags {
	val := ctx.Value(keyTags)
	if val == nil {
		return Tags{}
	}

	typed, ok := val.(Tags)
	if !ok {
		return Tags{}
	}

	return typed
}

func OrderFromContext(ctx context.Context) tagOrder {
	val := ctx.Value(keyOrder)
	if val == nil {
		return tagOrder{}
	}

	typed, ok := val.(tagOrder)
	if !ok {
		return tagOrder{}
	}

	return typed
}

// With adds a tag to the context to carry into subsequent logging calls.
func With(ctx context.Context, k string, v any) context.Context {
	return WithAll(ctx, Tags{k: v})
}

// WithAll adds multiple tags at once to a context, which avoids a ton of GC churn.
func WithAll(ctx context.Context, tags Tags) context.Context {
	order := OrderFromContext(ctx)
	ts := TagsFromContext(ctx)

	// Add all the tags.
	for k, val := range tags {
		// Don't print multiple times.
		if _, exists := ts[k]; !exists {
			order = append(order, k)
		}

		ts[k] = val
	}

	ctx = context.WithValue(ctx, keyOrder, order)
	ctx = context.WithValue(ctx, keyTags, ts)

	return ctx
}

// EachTag executes a given func() over each of the tags in order of addition.
func EachTag(ctx context.Context, f func(k string, v any) error) error {
	order := OrderFromContext(ctx)
	tags := TagsFromContext(ctx)

	// Process the tags in order of addition, which makes a nice nesting effect.
	for _, k := range order {
		val := tags[k]

		if err := f(k, val); err != nil {
			return err
		}
	}

	// Always add the globalUUID last.
	if err := f("meta.instance", globalUUID); err != nil {
		return err
	}

	return nil
}

func SetDebug(to bool) {
	*debug = to
}

func SetTrace(to bool) {
	*trace = to
}
