package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/gaylatea/instrument"
)

func init() {
	flag.Parse()
}

type PrintLogger struct{}

func (pl *PrintLogger) Log(ctx context.Context, l instrument.Level, filename string, line int, msg string, args ...interface{}) error {
	fmt.Printf("[🌎] %s\n", fmt.Sprintf(msg, args...))
	return nil
}

func main() {
	bare := context.Background()
	instrument.Infof(bare, "This is a bare context with no attached tags.")

	ctx := instrument.With(bare, "test", "hello")
	instrument.Tracef(ctx, "This is a trace level message.")
	instrument.Debugf(ctx, "This is a debug level message.")

	instrument.Infof(ctx, "This is an info level message.")

	_ = instrument.WithSpan(ctx, "Hella cool span", func(ctx context.Context, _ func(instrument.Tags)) error {
		instrument.Warnf(ctx, "This should appear inside of the trace.")
		_ = instrument.WithSpan(ctx, "Nested span", func(ctx context.Context, _ func(instrument.Tags)) error {
			instrument.Infof(ctx, "This is a doubly-nested log.")
			instrument.Infof(ctx, "This is a doubly-nested log.")
			instrument.Infof(ctx, "This is a doubly-nested log.")
			instrument.Infof(ctx, "This is a doubly-nested log.")
			instrument.Infof(ctx, "This is a doubly-nested log.")
			instrument.Infof(ctx, "This is a doubly-nested log.")
			instrument.Infof(ctx, "This is a doubly-nested log.")
			instrument.Infof(ctx, "This is a doubly-nested log.")
			return nil
		})
		return nil
	})

	ctx = instrument.With(ctx, "test", "list")
	ctx = instrument.WithLogSink(ctx, "print", &PrintLogger{})
	instrument.Warnf(ctx, "This is a warning level message.")
	instrument.Errorf(ctx, "This is an error level message.")

	timedOut, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	instrument.Infof(timedOut, "This message should still appear with tags.")

	instrument.Fatalf(bare, "This is a fatal message- the program should exit.")
}
