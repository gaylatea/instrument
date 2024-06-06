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

type Tags map[string]any
type sinks map[string]Sink

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
	silent = flag.Bool("silent", false, "Silence terminal output from default sink. Will not affect other sinks.")

	// The logging context always includes a random UUID for this particular invocation of this program.
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

// UseSink sets a global sink for all events.
func UseSink(name string, s Sink) {
	ctx := context.Background()

	for sinkName := range allSinks(ctx) {
		if name == sinkName {
			Fatalf(ctx, "Cannot override existing '%s' sink!")
		}
	}

	globalSinks[name] = s
}

// WithSink adds a sink for this context.
func WithSink(ctx context.Context, name string, s Sink) context.Context {
	for sinkName := range allSinks(ctx) {
		if name == sinkName {
			Fatalf(ctx, "Cannot override existing '%s' sink!")
		}
	}

	sinks := sinksFromContext(ctx)
	sinks[name] = s

	return context.WithValue(ctx, keyConfiguredSinks, sinks)
}

// With adds a tag to the context to carry into subsequent logging calls.
func With(ctx context.Context, k string, v any) context.Context {
	return WithAll(ctx, Tags{k: v})
}

// WithAll adds multiple tags at once to a context, which might avoid GC churn.
func WithAll(ctx context.Context, tags Tags) context.Context {
	ts := tagsFromContext(ctx)

	// Add all the tags.
	for k, val := range tags {
		ts[k] = val
	}

	return context.WithValue(ctx, keyTags, ts)
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
func getCaller(depth int) (caller, file string, line int) {
	pc, _, _, _ := runtime.Caller(depth)
	fn := runtime.FuncForPC(pc)

	file, line = fn.FileLine(pc)
	return fn.Name(), file, line
}
