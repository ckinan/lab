package proc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ckinan/lab/internal/domain"
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
	fields := map[string]*int64{}
	var m domain.Memory
	fields["MemTotal"] = &m.Total
	fields["MemFree"] = &m.Free
	fields["MemAvailable"] = &m.Available
	fields["Buffers"] = &m.Buffers
	fields["Cached"] = &m.Cached
	fields["Shmem"] = &m.Shmem
	fields["SwapTotal"] = &m.SwapTotal
	fields["SwapFree"] = &m.SwapFree

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		dest, ok := fields[key]
		if !ok {
			continue
		}
		v, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		*dest = v * 1024 // /proc/meminfo values are in kB
	}
	if err := scanner.Err(); err != nil {
		return domain.Memory{}, fmt.Errorf("reading /proc/meminfo: %w", err)
	}
	if m.Total == 0 || m.Available == 0 {
		return domain.Memory{}, fmt.Errorf("/proc/meminfo: missing MemTotal or MemAvailable")
	}

	m.Used = m.Total - m.Available
	m.SwapUsed = m.SwapTotal - m.SwapFree
	return m, nil
}
