package instrument

// This file is adapted from Coda Hale's metrics library with Instrument-specific modifications.
// The original can be found at: https://github.com/codahale/metrics
//
// Counters
//
// A counter is a monotonically-increasing, unsigned, 64-bit integer used to represent the number of times an event
// has occurred. By tracking the deltas between measurements of a counter over intervals of time, an aggregation
// layer can derive rates, acceleration, etc.
//
// Gauges
//
// A gauge returns instantaneous measurements of something using signed, 64-bit integers.
// This value does not need to be monotonic.
//
// Histograms
//
// A histogram tracks the distribution of a stream of values (e.g. the number of milliseconds it takes to handle
// requests), adding gauges for the values at meaningful quantiles: 50th, 75th, 90th, 95th, 99th, 99.9th.
import (
	"context"
	"sync"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/pkg/errors"
)

// A Counter is a monotonically increasing unsigned integer.
//
// Use a counter to derive rates (e.g., record total number of requests, derive requests per second).
type Counter string

// Add increments the counter by one.
func (c Counter) Add() {
	c.AddN(1)
}

// AddN increments the counter by N.
func (c Counter) AddN(delta uint64) {
	current, ok := counters.Load(c)
	if ok {
		current += delta
		counters.Store(c, current)
	} else {
		counters.Store(c, delta)
	}
}

// A Gauge is an instantaneous measurement of a value.
//
// Use a gauge to track metrics which increase and decrease (e.g., amount of free memory).
type Gauge string

// Set the gauge's value to the given value.
func (g Gauge) Set(value int64) {
	gauges.Store(g, func() int64 {
		return value
	})
}

// setBatchFunc sets the gauge's value to the lazily-called return value of the given function, with an additional
// initializer function for a related batch of gauges, all of which are keyed by an arbitrary value.
//
// At the moment this is unexported because it's only used by histograms, and I want to keep the interface simple.
func (g Gauge) setBatchFunc(key any, init func(), f func() int64) {
	gauges.Store(g, f)

	if _, ok := inits.Load(key); !ok {
		inits.Store(key, init)
	}
}

type hname string // unexported to prevent collisions

// NewHistogram returns a windowed HDR histogram which drops data older than five minutes.
// The returned histogram is safe to use from multiple goroutines.
//
// Use a histogram to track the distribution of a stream of values (e.g., the latency associated with HTTP requests).
func NewHistogram(name string, minValue, maxValue int64, sigfigs int) *Histogram {
	if _, ok := histograms.Load(name); ok {
		panic(name + " already exists")
	}

	hist := &Histogram{
		name: name,
		hist: hdrhistogram.NewWindowed(5, minValue, maxValue, sigfigs),
	}
	histograms.Store(name, hist)

	Gauge(name+".p50").setBatchFunc(hname(name), hist.merge, hist.valueAt(50))
	Gauge(name+".p75").setBatchFunc(hname(name), hist.merge, hist.valueAt(75))
	Gauge(name+".p90").setBatchFunc(hname(name), hist.merge, hist.valueAt(90))
	Gauge(name+".p95").setBatchFunc(hname(name), hist.merge, hist.valueAt(95))
	Gauge(name+".p99").setBatchFunc(hname(name), hist.merge, hist.valueAt(99))
	Gauge(name+".p999").setBatchFunc(hname(name), hist.merge, hist.valueAt(99.9))

	return hist
}

// A Histogram measures the distribution of a stream of values.
type Histogram struct {
	name string
	hist *hdrhistogram.WindowedHistogram
	m    *hdrhistogram.Histogram
	rw   sync.RWMutex
}

// Name returns the name of the histogram.
func (h *Histogram) Name() string {
	return h.name
}

// RecordValue records the given value, or returns an error if the value is out of range.
func (h *Histogram) RecordValue(v int64) error {
	h.rw.Lock()
	defer h.rw.Unlock()

	err := h.hist.Current.RecordValue(v)
	if err != nil {
		return errors.Wrap(err, h.name)
	}
	return nil
}

func (h *Histogram) rotate() {
	h.rw.Lock()
	defer h.rw.Unlock()

	h.hist.Rotate()
}

func (h *Histogram) merge() {
	h.rw.Lock()
	defer h.rw.Unlock()

	h.m = h.hist.Merge()
}

func (h *Histogram) valueAt(q float64) func() int64 {
	return func() int64 {
		h.rw.RLock()
		defer h.rw.RUnlock()

		if h.m == nil {
			return 0
		}

		return h.m.ValueAtQuantile(q)
	}
}

type SyncMap[K comparable, V any] struct {
	m sync.Map
}

func (sm *SyncMap[K, V]) Load(key K) (V, bool) {
	val, ok := sm.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return val.(V), true
}

func (sm *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	sm.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

func (sm *SyncMap[K, V]) Store(key K, value V) {
	sm.m.Store(key, value)
}

var (
	counters   = SyncMap[Counter, uint64]{}
	gauges     = SyncMap[Gauge, func() int64]{}
	inits      = SyncMap[any, func()]{}
	histograms = SyncMap[string, *Histogram]{}

	metricsTotal Gauge = "instrument.metrics.registered"
)

// Flush is called on a given interval to emit the metrics events to all configured sinks.
// It can also be called manually to immediately flush all known events.
func Flush() {
	total := 0
	inits.Range(func(key any, value func()) bool {
		value()
		return true
	})

	counters.Range(func(key Counter, value uint64) bool {
		total += 1
		emit(context.Background(), Tags{
			"metric.name":  string(key),
			"metric.value": value,
			"meta.level":   METRIC,
		})

		return true
	})

	gauges.Range(func(key Gauge, value func() int64) bool {
		total += 1
		emit(context.Background(), Tags{
			"metric.name":  string(key),
			"metric.value": value(),
			"meta.level":   METRIC,
		})

		return true
	})

	metricsTotal.Set(int64(total))
	return
}

func init() {
	go func() {
		for _ = range time.NewTicker(1 * time.Minute).C {
			histograms.Range(func(key string, value *Histogram) bool {
				value.rotate()
				return true
			})
		}
	}()

	// TODO: make this configurable at runtime
	go func() {
		for _ = range time.NewTicker(1 * time.Minute).C {
			Flush()
		}
	}()
}
