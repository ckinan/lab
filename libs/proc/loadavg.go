package proc

import (
	"fmt"
	"os"
	"strconv"
	"strings"

)

type ProcLoadAvgReader struct{}

func (r ProcLoadAvgReader) ReadLoadAvg() (LoadAvg, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return LoadAvg{}, fmt.Errorf("read /proc/loadavg: %w", err)
	}
	return parseLoadAvg(strings.TrimSpace(string(data)))
}

func parseLoadAvg(line string) (LoadAvg, error) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return LoadAvg{}, fmt.Errorf("unexpected /proc/loadavg format")
	}

	avg1m, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return LoadAvg{}, fmt.Errorf("parse load avg 1m: %w", err)
	}
	avg5m, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return LoadAvg{}, fmt.Errorf("parse load avg 5m: %w", err)
	}
	avg15m, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return LoadAvg{}, fmt.Errorf("parse load avg 15m: %w", err)
	}

	// field 3 is "running/total"
	parts := strings.SplitN(fields[3], "/", 2)
	if len(parts) != 2 {
		return LoadAvg{}, fmt.Errorf("unexpected tasks field in /proc/loadavg")
	}
	running, err := strconv.Atoi(parts[0])
	if err != nil {
		return LoadAvg{}, fmt.Errorf("parse running tasks: %w", err)
	}
	total, err := strconv.Atoi(parts[1])
	if err != nil {
		return LoadAvg{}, fmt.Errorf("parse total tasks: %w", err)
	}

	return LoadAvg{
		Avg1m:   avg1m,
		Avg5m:   avg5m,
		Avg15m:  avg15m,
		Running: running,
		Total:   total,
	}, nil
}
