package main

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/gaylatea/instrument"
)

var ErrUnknown = errors.New("unknown error")

// Create some sample metrics.
var (
	exampleCounter   instrument.Counter = "test.events"
	exampleGauge     instrument.Gauge   = "test.huh"
	exampleHistogram                    = instrument.NewHistogram("test.what", 0, 100, 2)
)

//nolint:funlen
func main() {
	bare := context.Background()
	instrument.PostEvent(bare, "Test event", nil)
	instrument.Infof(bare, "This shouldn't have any tags.")
	exampleGauge.Set(24)
	instrument.Flush()

	exampleGauge.Set(69)
	instrument.Tracef(bare, "This shouldn't show up.")
	instrument.Debugf(bare, "This shouldn't show up.")
	// Record some nice random values to test how histograms work.
	for _ = range [20]struct{}{} {
		val := rand.Int64() % 100
		exampleHistogram.RecordValue(val)
	}

	instrument.SetTrace(true)
	instrument.Tracef(bare, "This should show up.")
	instrument.Debugf(bare, "This should show up.")
	instrument.SetTrace(false)
	instrument.SetDebug(true)
	instrument.Tracef(bare, "This shouldn't show up.")
	instrument.Debugf(bare, "But this should.")

	tagged := instrument.With(bare, "test", "hello")
	instrument.Infof(tagged, "This should have a new tag.")

	_ = instrument.WithSpan(
		tagged,
		"Span 1",
		func(ctx context.Context, _ func(instrument.Tags)) error {
			instrument.Infof(ctx, "This should appear inside of the trace.")

			_ = instrument.WithSpan(
				ctx,
				"Span 2",
				func(ctx context.Context, _ func(instrument.Tags)) error {
					instrument.Infof(ctx, "This should be parented to Span 2.")

					return ErrUnknown
				},
			)
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
	instrument.Fatalf(bare, "This should exit the program while flushing telemetry.")
}
