package instrument

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/google/uuid"
)

type (
	Tags  map[string]any
	sinks map[string]Sink
)

type contextKey int

const (
	keyTags contextKey = iota
	keyOrder
	keyConfiguredSinks
	keyTraceID
)

var (
	debug  = flag.Bool("debug", false, "Enable debug logging.")
	trace  = flag.Bool("trace", false, "Enable trace logging. EXTREMELY VERBOSE.")
	silent = flag.Bool(
		"silent",
		false,
		"Silence terminal output from default sink. Will not affect other sinks.",
	)

	// The logging context always includes a random ID to differentiate program runs.
	instanceID uuid.UUID

	// Always include the terminal sink.
	terminal    = &TerminalSink{}
	globalSinks = sinks{
		"terminal": terminal,
	}
)

// Sink implementers receive events and pass them along to downstream systems.
type Sink interface {
	Event(ctx context.Context, t Tags) error
}

func init() {
	if id, err := uuid.NewV7(); err != nil {
		Fatalf(context.Background(), "Could not create a unique ID for this instance: %v", err)
	} else {
		instanceID = id
	}

	// Startup a signal handler to switch log levels at runtime.
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

// UseSink sets a global sink for all events.
func UseSink(name string, newSink Sink) {
	ctx := context.Background()

	for sinkName := range allSinks(ctx) {
		if name == sinkName {
			Fatalf(ctx, "Cannot override existing '%s' sink!")
		}
	}

	globalSinks[name] = newSink
}

// WithSink adds a sink for the given context.
func WithSink(ctx context.Context, name string, newSink Sink) context.Context {
	for sinkName := range allSinks(ctx) {
		if name == sinkName {
			Fatalf(ctx, "Cannot override existing '%s' sink!")
		}
	}

	sinks := sinksFromContext(ctx)
	sinks[name] = newSink

	return context.WithValue(ctx, keyConfiguredSinks, sinks)
}

// With adds a single tag to the given context.
func With(ctx context.Context, k string, v any) context.Context {
	return WithAll(ctx, Tags{k: v})
}

// WithAll adds multiple tags to the given context.
func WithAll(ctx context.Context, tags Tags) context.Context {
	parentTags := tagsFromContext(ctx)

	// Add all the tags.
	for k, val := range tags {
		parentTags[k] = val
	}

	return context.WithValue(ctx, keyTags, parentTags)
}

// SetDebug sets the visibility of debug events.
func SetDebug(to bool) {
	*debug = to
}

// SetTrace sets the visibility of trace and debug events.
func SetTrace(to bool) {
	*debug = to
	*trace = to
}

// Silence toggles the default terminal output.
func Silence(to bool) {
	*silent = to
}

// getCaller returns information up the stack, used for metadata.
func getCaller(depth int) (string, string, int) {
	pc, _, _, _ := runtime.Caller(depth)
	fn := runtime.FuncForPC(pc)
	file, line := fn.FileLine(pc)

	return fn.Name(), file, line
}
