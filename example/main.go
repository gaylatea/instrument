package main

import (
	"context"
	"errors"
	"time"

	"github.com/gaylatea/instrument"
)

var ErrUnknown = errors.New("unknown error")

// CounterSink counts incoming events.
type CounterSink struct {
	count int
}

// Event counts events.
func (c *CounterSink) Event(_ context.Context, _ instrument.Tags) error {
	c.count++

	return nil
}

//nolint:funlen
func main() {
	counter := &CounterSink{count: 0}
	instrument.UseSink("counter", counter)

	bare := context.Background()
	instrument.PostEvent(bare, "Test event", nil)
	instrument.Infof(bare, "This shouldn't have any tags.")

	instrument.Tracef(bare, "This shouldn't show up.")
	instrument.Debugf(bare, "This shouldn't show up.")

	instrument.SetTrace(true)
	instrument.Tracef(bare, "This should show up.")
	instrument.Debugf(bare, "This should show up.")
	instrument.SetTrace(false)
	instrument.SetDebug(true)
	instrument.Tracef(bare, "This shouldn't show up.")
	instrument.Debugf(bare, "This should show up.")

	tagged := instrument.With(bare, "test", "hello")
	instrument.Infof(tagged, "This should have a new tag.")

	_ = instrument.WithSpan(
		tagged,
		"Span 1",
		func(ctx context.Context, _ func(instrument.Tags)) error {
			instrument.Infof(ctx, "This should appear inside of the trace.")

			newCounter := &CounterSink{count: 0}
			ctx = instrument.WithSink(ctx, "counter2", newCounter)

			_ = instrument.WithSpan(
				ctx,
				"Span 2",
				func(ctx context.Context, _ func(instrument.Tags)) error {
					instrument.Infof(ctx, "This should be parented to Span 2.")

					return ErrUnknown
				},
			)

			instrument.Infof(ctx, "Span 1 collected %d events.", newCounter.count)

			return nil
		},
	)

	instrument.Silence(true)
	instrument.Infof(tagged, "This shouldn't show up.")
	instrument.Silence(false)

	override := instrument.With(tagged, "test", "list")
	instrument.Infof(override, "This should have a different tag.")
	instrument.Infof(tagged, "This should have the original tag.")

	instrument.Colorize()
	instrument.Errorf(tagged, "This should be colorized no matter what.")
	instrument.ResetColor()

	timedOut, cancel := context.WithTimeout(tagged, 1*time.Second)
	defer cancel()

	instrument.Infof(timedOut, "This should still have tags.")

	instrument.Infof(tagged, "%d events counted.", counter.count)
	instrument.Fatalf(bare, "This should exit the program.")
}
