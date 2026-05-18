package proc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type cpuSample struct {
	busy, total uint64
}

// ProcCPUReader needs two calls to return a non-zero value - first call seeds the baseline.
type ProcCPUReader struct {
	prev *cpuSample
}

func NewProcCPUReader() *ProcCPUReader {
	return &ProcCPUReader{}
}

func (r *ProcCPUReader) ReadCPU() (float64, error) {
	cur, err := readCPUSample()
	if err != nil {
		return 0, err
	}

	if r.prev == nil {
		r.prev = &cur
		return 0, nil
	}

	deltaBusy := cur.busy - r.prev.busy
	deltaTotal := cur.total - r.prev.total
	r.prev = &cur

	if deltaTotal == 0 {
		return 0, nil
	}
	return float64(deltaBusy) / float64(deltaTotal) * 100, nil
}

func readCPUSample() (cpuSample, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuSample{}, fmt.Errorf("read /proc/stat: %w", err)
	}
	return parseCPUSample(string(data))
}

func parseCPUSample(data string) (cpuSample, error) {
	line := strings.SplitN(data, "\n", 2)[0]
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuSample{}, fmt.Errorf("unexpected /proc/stat format")
	}

	ticks := make([]uint64, len(fields)-1)
	for i, f := range fields[1:] {
		v, err := strconv.ParseUint(f, 10, 64)
		if err != nil {
			return cpuSample{}, fmt.Errorf("parse /proc/stat field %d: %w", i, err)
		}
		ticks[i] = v
	}

	// idle includes iowait: CPU is parked but not doing useful work in both cases
	var total uint64
	for _, t := range ticks {
		total += t
	}
	idle := ticks[3]
	if len(ticks) > 4 {
		idle += ticks[4]
	}
	return cpuSample{busy: total - idle, total: total}, nil
}
