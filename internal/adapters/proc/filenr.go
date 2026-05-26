package proc

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ckinan/cktop/internal/domain"
)

type ProcFileNRReader struct{}

func (r ProcFileNRReader) ReadFileNR() (domain.FileNR, error) {
	data, err := os.ReadFile("/proc/sys/fs/file-nr")
	if err != nil {
		return domain.FileNR{}, fmt.Errorf("read /proc/sys/fs/file-nr: %w", err)
	}
	return parseFileNR(strings.TrimSpace(string(data)))
}

// /proc/sys/fs/file-nr: "open  0  max"
// field 1 (index 0): allocated fds
// field 2 (index 1): allocated but unused (legacy, always 0 on modern kernels)
// field 3 (index 2): maximum fds
func parseFileNR(line string) (domain.FileNR, error) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return domain.FileNR{}, fmt.Errorf("unexpected /proc/sys/fs/file-nr format")
	}
	open, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return domain.FileNR{}, fmt.Errorf("parse open fds: %w", err)
	}
	max, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return domain.FileNR{}, fmt.Errorf("parse max fds: %w", err)
	}
	return domain.FileNR{Open: open, Max: max}, nil
}
