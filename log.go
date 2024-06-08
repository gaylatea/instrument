package instrument

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
)

const logCallerSkip = 3

// logf emits an event for a given message, with log-specific metadata.
func logf(ctx context.Context, thisLevel Level, msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	caller, filename, line := getCaller(logCallerSkip)

	theseTags := tagsFromContext(ctx)
	traceID := traceIDFromContext(ctx)

	if traceID != uuid.Nil {
		theseTags["trace.parent"] = traceID
	}

	theseTags["meta.level"] = thisLevel
	theseTags["meta.caller"] = caller
	theseTags["meta.file"] = filename
	theseTags["meta.line"] = line
	theseTags["log.message"] = msg

	emit(ctx, theseTags)
}

// Infof prints an informational string to the console.
func Infof(ctx context.Context, msg string, args ...interface{}) {
	logf(ctx, INFO, msg, args...)
}

// Debugf prints debug information when in debug mode.
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

// Fatalf prints an error and quits the app.
func Fatalf(ctx context.Context, msg string, args ...interface{}) {
	logf(ctx, FATAL, msg, args...)
	os.Exit(1)
}
