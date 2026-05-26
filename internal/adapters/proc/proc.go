package proc

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ckinan/lab/internal/domain"
)

// clkTck is USER_HZ - virtually always 100 on Linux x86/arm64.
const clkTck = 100.0

type procSnapshot struct {
	ticks    float64
	wallTime time.Time
}

// ProcProcessReader maintains per-PID tick snapshots between calls to compute CPU %.
type ProcProcessReader struct {
	prev      map[int]procSnapshot
	userCache map[string]string // uid string -> username
}

func NewProcProcessReader() *ProcProcessReader {
	return &ProcProcessReader{
		prev:      make(map[int]procSnapshot),
		userCache: make(map[string]string),
	}
}

func (r *ProcProcessReader) ReadProcesses() ([]domain.Process, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("read /proc: %w", err)
	}

	now := time.Now()
	livePIDs := make(map[int]bool, len(entries))
	results := make([]domain.Process, 0, len(entries))

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue // not a PID directory
		}
		livePIDs[pid] = true

		proc, err := r.readOne(pid, now)
		if err != nil {
			continue // process may have exited mid-scan
		}
		results = append(results, proc)
	}

	// evict dead PIDs from the snapshot cache
	for pid := range r.prev {
		if !livePIDs[pid] {
			delete(r.prev, pid)
		}
	}

	return results, nil
}

type statusInfo struct {
	name  string
	ppid  int
	uid   string
	rssKB int64
}

func (r *ProcProcessReader) readOne(pid int, now time.Time) (domain.Process, error) {
	base := fmt.Sprintf("/proc/%d", pid)

	statusData, err := os.ReadFile(base + "/status")
	if err != nil {
		return domain.Process{}, err
	}
	st := parseStatus(statusData)

	cmdlineData, _ := os.ReadFile(base + "/cmdline")
	cmdline := parseCmdline(cmdlineData)
	isKthread := cmdline == ""
	if isKthread {
		cmdline = "[" + st.name + "]"
	} else {
		parts := strings.SplitN(cmdline, " ", 2)
		parts[0] = filepath.Base(parts[0])
		cmdline = strings.Join(parts, " ")
	}

	statData, err := os.ReadFile(base + "/stat")
	if err != nil {
		return domain.Process{}, err
	}
	utime, stime, err := parseStatTicks(statData)
	if err != nil {
		return domain.Process{}, err
	}
	ticks := float64(utime + stime)
	cpuPct := r.cpuPercent(pid, ticks, now)
	r.prev[pid] = procSnapshot{ticks: ticks, wallTime: now}

	return domain.Process{
		Pid:       pid,
		Ppid:      st.ppid,
		Rss:       int(st.rssKB * 1024),
		CPU:       cpuPct,
		Cmdline:   cmdline,
		Username:  r.lookupUID(st.uid),
		IsKthread: isKthread,
	}, nil
}

func (r *ProcProcessReader) cpuPercent(pid int, ticks float64, now time.Time) float64 {
	prev, ok := r.prev[pid]
	if !ok {
		return 0
	}
	elapsed := now.Sub(prev.wallTime).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return (ticks - prev.ticks) / clkTck / elapsed * 100
}

func (r *ProcProcessReader) lookupUID(uid string) string {
	if name, ok := r.userCache[uid]; ok {
		return name
	}
	u, err := user.LookupId(uid)
	if err != nil {
		r.userCache[uid] = uid
		return uid
	}
	r.userCache[uid] = u.Username
	return u.Username
}

func parseStatus(data []byte) statusInfo {
	var st statusInfo
	found := 0
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		val := strings.TrimSpace(parts[1])
		switch parts[0] {
		case "Name":
			st.name = val
			found++
		case "PPid":
			st.ppid, _ = strconv.Atoi(val)
			found++
		case "Uid":
			// format: real effective saved fs - take real uid
			st.uid = strings.Fields(val)[0]
			found++
		case "VmRSS":
			// format: "1234 kB"
			st.rssKB, _ = strconv.ParseInt(strings.Fields(val)[0], 10, 64)
			found++
		}
		if found == 4 {
			break
		}
	}
	return st
}

func parseCmdline(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	return strings.TrimRight(strings.ReplaceAll(string(data), "\x00", " "), " ")
}

func parseStatTicks(data []byte) (utime, stime uint64, err error) {
	s := string(data)
	// comm can contain spaces and parentheses; find the last ')' to be safe.
	// Fields after ") ": state ppid ... utime(idx 11) stime(idx 12)
	idx := strings.LastIndex(s, ")")
	if idx < 0 {
		return 0, 0, fmt.Errorf("invalid /proc stat format")
	}
	fields := strings.Fields(s[idx+2:])
	if len(fields) < 13 {
		return 0, 0, fmt.Errorf("not enough fields in stat")
	}
	utime, err = strconv.ParseUint(fields[11], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	stime, err = strconv.ParseUint(fields[12], 10, 64)
	return utime, stime, err
}
