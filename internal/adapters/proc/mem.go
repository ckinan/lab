package proc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ckinan/cktop/internal/domain"
)

type ProcMemoryReader struct{}

func (r ProcMemoryReader) ReadMemory() (domain.Memory, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return domain.Memory{}, fmt.Errorf("open /proc/meminfo: %w", err)
	}
	defer f.Close()
	return parseMeminfo(f)
}

func parseMeminfo(r io.Reader) (domain.Memory, error) {
	var totalKB, availKB int64
	var hasTotal, hasAvail bool
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		v, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		switch key {
		case "MemTotal":
			totalKB, hasTotal = v, true
		case "MemAvailable":
			availKB, hasAvail = v, true
		}
		if hasTotal && hasAvail {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return domain.Memory{}, fmt.Errorf("reading /proc/meminfo: %w", err)
	}

	if !hasTotal || !hasAvail {
		return domain.Memory{}, fmt.Errorf("/proc/meminfo: missing MemTotal or MemAvailable")
	}

	total := totalKB * 1024
	available := availKB * 1024
	return domain.Memory{
		Total:     total,
		Available: available,
		Used:      total - available,
	}, nil
}
