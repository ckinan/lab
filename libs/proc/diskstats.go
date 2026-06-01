package proc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

)

type ProcDiskStatsReader struct{}

func (r ProcDiskStatsReader) ReadDiskStats() ([]DiskStat, error) {
	f, err := os.Open("/proc/diskstats")
	if err != nil {
		return nil, fmt.Errorf("open /proc/diskstats: %w", err)
	}
	defer f.Close()
	return parseDiskStats(f)
}

// parseDiskStats parses /proc/diskstats.
// Fields (1-indexed after major/minor/name):
//   4: reads completed, 6: sectors read, 8: writes completed, 10: sectors written, 13: ms doing I/O
func parseDiskStats(r io.Reader) ([]DiskStat, error) {
	var stats []DiskStat
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 14 {
			continue
		}
		device := fields[2]
		if skipDevice(device) {
			continue
		}
		reads, err := strconv.ParseUint(fields[3], 10, 64)
		if err != nil {
			continue
		}
		sectorsRead, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			continue
		}
		writes, err := strconv.ParseUint(fields[7], 10, 64)
		if err != nil {
			continue
		}
		sectorsWritten, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			continue
		}
		ioTimeMs, err := strconv.ParseUint(fields[12], 10, 64)
		if err != nil {
			continue
		}
		stats = append(stats, DiskStat{
			Device:      device,
			ReadsTotal:  reads,
			WritesTotal: writes,
			ReadBytes:   sectorsRead * 512,
			WriteBytes:  sectorsWritten * 512,
			IOTimeMs:    ioTimeMs,
		})
	}
	return stats, scanner.Err()
}

// skipDevice filters out virtual/loop/ram devices that are not real storage.
func skipDevice(name string) bool {
	return strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram")
}
