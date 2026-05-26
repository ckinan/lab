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

type ProcNetDevReader struct{}

func (r ProcNetDevReader) ReadNetDev() ([]domain.NetDev, error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, fmt.Errorf("open /proc/net/dev: %w", err)
	}
	defer f.Close()
	return parseNetDev(f)
}

// parseNetDev parses /proc/net/dev.
// After stripping the "iface:" prefix the columns are:
//   rx: bytes(0) packets(1) errs(2) drop(3) fifo(4) frame(5) compressed(6) multicast(7)
//   tx: bytes(8) packets(9) errs(10) drop(11) fifo(12) colls(13) carrier(14) compressed(15)
func parseNetDev(r io.Reader) ([]domain.NetDev, error) {
	var devs []domain.NetDev
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		// skip the two header lines
		if !strings.Contains(line, ":") {
			continue
		}
		colon := strings.Index(line, ":")
		iface := strings.TrimSpace(line[:colon])
		fields := strings.Fields(line[colon+1:])
		if len(fields) < 16 {
			continue
		}
		parse := func(i int) uint64 {
			v, _ := strconv.ParseUint(fields[i], 10, 64)
			return v
		}
		devs = append(devs, domain.NetDev{
			Interface: iface,
			RxBytes:   parse(0),
			RxPackets: parse(1),
			RxErrors:  parse(2),
			RxDrops:   parse(3),
			TxBytes:   parse(8),
			TxPackets: parse(9),
			TxErrors:  parse(10),
			TxDrops:   parse(11),
		})
	}
	return devs, scanner.Err()
}
