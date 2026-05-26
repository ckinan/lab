package goruntime

import (
	"runtime/metrics"
	"strings"
)

// Desc is a scalar Go runtime metric descriptor.
type Desc struct {
	// Name is the runtime/metrics canonical name, e.g. /gc/heap/allocs:bytes.
	Name string
	// PromName is the Prometheus metric name, e.g. go_gc_heap_allocs_bytes.
	PromName string
	// Cumulative indicates the metric is a monotonically increasing counter.
	Cumulative bool
}

// Reader samples all scalar Go runtime metrics.
// Histograms and /godebug/* metrics are excluded.
type Reader struct {
	descs   []Desc
	samples []metrics.Sample
	values  map[string]float64
}

// NewReader constructs a Reader. Discovers available metrics at construction time.
func NewReader() *Reader {
	all := metrics.All()
	var descs []Desc
	var samples []metrics.Sample

	for _, d := range all {
		if d.Kind == metrics.KindFloat64Histogram {
			continue
		}
		if strings.HasPrefix(d.Name, "/godebug/") {
			continue
		}
		descs = append(descs, Desc{
			Name:       d.Name,
			PromName:   promName(d.Name),
			Cumulative: d.Cumulative,
		})
		samples = append(samples, metrics.Sample{Name: d.Name})
	}

	return &Reader{
		descs:   descs,
		samples: samples,
		values:  make(map[string]float64, len(descs)),
	}
}

// Descs returns the list of scalar metrics this reader exposes.
func (r *Reader) Descs() []Desc {
	return r.descs
}

// Refresh reads all metrics from the Go runtime. Must be called before Get.
func (r *Reader) Refresh() {
	metrics.Read(r.samples)
	for _, s := range r.samples {
		switch s.Value.Kind() {
		case metrics.KindUint64:
			r.values[s.Name] = float64(s.Value.Uint64())
		case metrics.KindFloat64:
			r.values[s.Name] = s.Value.Float64()
		}
	}
}

// Get returns the last sampled value for a metric by its runtime/metrics name.
// Returns 0 if the name is unknown.
func (r *Reader) Get(name string) float64 {
	return r.values[name]
}

// promName converts a runtime/metrics name to a Prometheus metric name.
// Example: /gc/heap/allocs:bytes → go_gc_heap_allocs_bytes
func promName(name string) string {
	s := strings.TrimPrefix(name, "/")
	s = strings.NewReplacer("/", "_", ":", "_", "-", "_").Replace(s)
	return "go_" + s
}
