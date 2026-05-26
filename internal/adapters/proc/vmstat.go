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

type ProcVMStatReader struct{}

func (r ProcVMStatReader) ReadVMStat() (domain.VMStat, error) {
	f, err := os.Open("/proc/vmstat")
	if err != nil {
		return domain.VMStat{}, fmt.Errorf("open /proc/vmstat: %w", err)
	}
	defer f.Close()
	return parseVMStat(f)
}

func parseVMStat(r io.Reader) (domain.VMStat, error) {
	want := map[string]*uint64{}
	var s domain.VMStat
	want["pgfault"] = &s.PageFaults
	want["pgmajfault"] = &s.MajorPageFaults
	want["pswpin"] = &s.SwapIn
	want["pswpout"] = &s.SwapOut
	want["pgpgin"] = &s.PageIn
	want["pgpgout"] = &s.PageOut

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 2 {
			continue
		}
		dest, ok := want[parts[0]]
		if !ok {
			continue
		}
		v, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}
		*dest = v
	}
	return s, scanner.Err()
}
