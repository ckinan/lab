package proc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ProcUptimeReader struct{}

func (r ProcUptimeReader) ReadUptime() (float64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, fmt.Errorf("read /proc/uptime: %w", err)
	}
	return parseUptime(strings.TrimSpace(string(data)))
}

func parseUptime(line string) (float64, error) {
	fields := strings.Fields(line)
	if len(fields) < 1 {
		return 0, fmt.Errorf("unexpected /proc/uptime format")
	}
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse uptime seconds: %w", err)
	}
	return seconds, nil
}
