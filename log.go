package instrument

import (
	"context"
	"os"
	"runtime"
)

// Level represents a standard logging level.
// https://opentelemetry.io/docs/specs/otel/logs/data-model/#field-severitynumber
type Level int

const (
	// Extremely fine-grained log statements.
	TRACE Level = iota
	// Log statements not normally useful, but that would be helpful in fixing potential issues.
	DEBUG
	// Informational log statements, which are visible by default.
	INFO
	// Logs that indicate a non-critical issue encountered.
	WARN
	// Logs that indicate an encountered issue that stops the current module but shouldn't stop all processes.
	ERROR
	// Logs that indicate a critical error that must stop all processes.
	FATAL
)

// LogSink implementers accept event data and forward it to a user interface or analysis program.
type LogSink interface {
	// Log actually performs the sending of a log message.
	Log(ctx context.Context, l Level, filemame string, line int, msg string, args ...interface{}) error
}

// UseLogSink adds a sink which receives all logs output by the app.
func UseLogSink(name string, s LogSink) {
	globalLogSinks[name] = s
}

// logSinksFromContext fetches a copy of the currently configured log sinks, or an empty map if none exist.
func logSinksFromContext(ctx context.Context) logSinks {
	val := ctx.Value(keyConfiguredLogSinks)
	if val == nil {
		return logSinks{}
	}

	typed, ok := val.(logSinks)
	if !ok {
		return logSinks{}
	}

	return typed
}

// WithLogSink adds a log sink to the current context.
func WithLogSink(ctx context.Context, name string, s LogSink) context.Context {
	sinks := logSinksFromContext(ctx)
	sinks[name] = s

	return context.WithValue(ctx, keyConfiguredLogSinks, sinks)
}

// logf emits a log message to global sinks, as well as any defined in the context itself.
func logf(ctx context.Context, l Level, msg string, args ...interface{}) {
	_, filename, line, _ := runtime.Caller(2)

	for name, sink := range globalLogSinks {
		if err := sink.Log(ctx, l, filename, line, msg, args...); err != nil {
			_ = globalLogSinks["console"].Log(ctx, ERROR, filename, line, "Could not process log sink '%s': %v", name, err)
		}
	}

	ctxSinks := logSinksFromContext(ctx)
	for name, sink := range ctxSinks {
		if err := sink.Log(ctx, l, filename, line, msg, args...); err != nil {
			_ = globalLogSinks["console"].Log(ctx, ERROR, filename, line, "Could not process log sink '%s': %v", name, err)
		}
	}
}

// Infof prints an informational string to the console.
func Infof(ctx context.Context, msg string, args ...interface{}) {
	logf(ctx, INFO, msg, args...)
}

// Debugf prints debug info when in debug mode.
func Debugf(ctx context.Context, msg string, args ...interface{}) {
	// Tracing implies debug.
	if !*debug && !*trace {
		return
	}

	logf(ctx, DEBUG, msg, args...)
}

// Tracef prints tracing information when in trace mode.
func Tracef(ctx context.Context, msg string, args ...interface{}) {
	if !*trace {
		return
	}

	logf(ctx, TRACE, msg, args...)
}

// Errorf prints an error log to the console.
func Errorf(ctx context.Context, msg string, args ...interface{}) {
	logf(ctx, ERROR, msg, args...)
}

// Warnf prints a warning message.
func Warnf(ctx context.Context, msg string, args ...interface{}) {
	logf(ctx, WARN, msg, args...)
}

// Fatalf prints an error and immediately stops the app.
func Fatalf(ctx context.Context, msg string, args ...interface{}) {
	logf(ctx, FATAL, msg, args...)
	os.Exit(1)
}
