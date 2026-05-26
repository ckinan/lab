package proc

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ckinan/cktop/internal/domain"
)

// ProcServiceStatsReader scans all /proc/[pid] entries and aggregates RSS and CPU
// per systemd unit (extracted from /proc/[pid]/cgroup).
// Like ProcCPUReader, the first call seeds the baseline and returns all-zero CPU values.
type ProcServiceStatsReader struct {
	prev     map[int]procSnapshot
	prevTime time.Time
}

func NewProcServiceStatsReader() *ProcServiceStatsReader {
	return &ProcServiceStatsReader{prev: make(map[int]procSnapshot)}
}

func (r *ProcServiceStatsReader) ReadServiceStats() ([]domain.ServiceStat, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("read /proc: %w", err)
	}

	now := time.Now()
	elapsed := now.Sub(r.prevTime).Seconds()

	type agg struct {
		rss          int64
		cpuTicks     float64
		processCount int
	}
	units := make(map[string]*agg)
	livePIDs := make(map[int]bool, len(entries))

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		livePIDs[pid] = true

		base := fmt.Sprintf("/proc/%d", pid)

		unit, err := readCgroupUnit(base + "/cgroup")
		if err != nil {
			continue
		}

		rss, err := readRSSKB(base + "/status")
		if err != nil {
			continue
		}

		utime, stime, err := readStatTicksFromPath(base + "/stat")
		if err != nil {
			continue
		}
		ticks := float64(utime + stime)

		a := units[unit]
		if a == nil {
			a = &agg{}
			units[unit] = a
		}
		a.rss += rss * 1024
		a.processCount++

		if prev, ok := r.prev[pid]; ok && elapsed > 0 {
			a.cpuTicks += ticks - prev.ticks
		}
		r.prev[pid] = procSnapshot{ticks: ticks, wallTime: now}
	}

	// evict dead PIDs
	for pid := range r.prev {
		if !livePIDs[pid] {
			delete(r.prev, pid)
		}
	}
	r.prevTime = now

	stats := make([]domain.ServiceStat, 0, len(units))
	for unit, a := range units {
		cpuPct := 0.0
		if elapsed > 0 {
			cpuPct = a.cpuTicks / clkTck / elapsed * 100
		}
		stats = append(stats, domain.ServiceStat{
			Unit:         unit,
			RSSBytes:     a.rss,
			CPUPercent:   cpuPct,
			ProcessCount: a.processCount,
		})
	}
	return stats, nil
}

// readCgroupUnit extracts a short unit name from /proc/[pid]/cgroup.
// cgroup v2 format: "0::<path>" e.g. "0::/system.slice/cups.service"
// Returns the last path component, or "(untracked)" for bare "/" or kernel threads.
func readCgroupUnit(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "0::") {
			continue
		}
		cgPath := strings.TrimPrefix(line, "0::")
		cgPath = strings.TrimSpace(cgPath)
		if cgPath == "/" || cgPath == "" {
			return "(untracked)", nil
		}
		return filepath.Base(cgPath), nil
	}
	return "(untracked)", nil
}

// readRSSKB reads VmRSS from /proc/[pid]/status (in kB).
func readRSSKB(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "VmRSS:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0, nil
		}
		return strconv.ParseInt(fields[1], 10, 64)
	}
	return 0, nil
}

// readStatTicksFromPath reads utime+stime from /proc/[pid]/stat.
func readStatTicksFromPath(path string) (utime, stime uint64, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	return parseStatTicks(data)
}
