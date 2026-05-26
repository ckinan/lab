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

type ProcSockStatReader struct{}

func (r ProcSockStatReader) ReadSockStat() (domain.SockStat, error) {
	f, err := os.Open("/proc/net/sockstat")
	if err != nil {
		return domain.SockStat{}, fmt.Errorf("open /proc/net/sockstat: %w", err)
	}
	defer f.Close()
	return parseSockStat(f)
}

func parseSockStat(r io.Reader) (domain.SockStat, error) {
	var s domain.SockStat
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		proto := strings.TrimSuffix(fields[0], ":")
		kv := fieldsToMap(fields[1:])
		switch proto {
		case "TCP":
			s.TCPUsed = kv["inuse"]
			s.TCPOrphan = kv["orphan"]
			s.TCPTimeWait = kv["tw"]
		case "UDP":
			s.UDPUsed = kv["inuse"]
		case "RAW":
			s.RAWUsed = kv["inuse"]
		}
	}
	if err := scanner.Err(); err != nil {
		return domain.SockStat{}, fmt.Errorf("reading /proc/net/sockstat: %w", err)
	}
	return s, nil
}

// fieldsToMap parses alternating key-value pairs like "inuse 31 orphan 0 tw 5"
func fieldsToMap(fields []string) map[string]int {
	m := make(map[string]int, len(fields)/2)
	for i := 0; i+1 < len(fields); i += 2 {
		v, err := strconv.Atoi(fields[i+1])
		if err == nil {
			m[fields[i]] = v
		}
	}
	return m
}
