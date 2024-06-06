package instrument

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
)

// logf emits an event for a given message, with log-specific metadata.
func logf(ctx context.Context, l Level, msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	caller, filename, line := getCaller(3)

	t := tagsFromContext(ctx)
	traceID := traceIDFromContext(ctx)
	if traceID != uuid.Nil {
		t["trace.parent"] = traceID
	}

	t["meta.level"] = l
	t["meta.caller"] = caller
	t["meta.file"] = filename
	t["meta.line"] = line
	t["log.message"] = msg

	emit(ctx, t)
}

// Infof prints an informational string to the console.
func Infof(ctx context.Context, msg string, args ...interface{}) {
	logf(ctx, INFO, msg, args...)
}

// Debugf prints debug info when in debug mode.
func Debugf(ctx context.Context, msg string, args ...interface{}) {
	logf(ctx, DEBUG, msg, args...)
}

// Tracef prints tracing information when in trace mode.
func Tracef(ctx context.Context, msg string, args ...interface{}) {
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
