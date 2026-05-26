package proc

import (
	"os"
	"strings"
	"testing"

	"github.com/ckinan/cktop/internal/domain"
)

func TestParseMeminfo(t *testing.T) {
	input := `MemTotal:       16384000 kB
MemFree:         2048000 kB
MemAvailable:    8192000 kB
Buffers:          512000 kB
Cached:          1024000 kB
Shmem:            256000 kB
SwapTotal:       8192000 kB
SwapFree:        8192000 kB
`
	mem, err := parseMeminfo(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := domain.Memory{
		Total:     16384000 * 1024,
		Free:      2048000 * 1024,
		Available: 8192000 * 1024,
		Used:      (16384000 - 8192000) * 1024,
		Buffers:   512000 * 1024,
		Cached:    1024000 * 1024,
		Shmem:     256000 * 1024,
		SwapTotal: 8192000 * 1024,
		SwapFree:  8192000 * 1024,
		SwapUsed:  0,
	}
	if mem != want {
		t.Errorf("got %+v, want %+v", mem, want)
	}
}

func TestParseMeminfo_MissingField(t *testing.T) {
	_, err := parseMeminfo(strings.NewReader("MemTotal: 1024 kB\n"))
	if err == nil {
		t.Error("expected error when MemAvailable is missing")
	}
}

func TestParseCPUSample(t *testing.T) {
	// cpu user nice system idle iowait irq softirq steal guest guest_nice
	input := "cpu  100 0 50 800 50 0 0 0 0 0\n"
	s, err := parseCPUSample(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// total = 1000, idle = 800+50 = 850, busy = 150
	if s.total != 1000 {
		t.Errorf("total: got %d, want 1000", s.total)
	}
	if s.busy != 150 {
		t.Errorf("busy: got %d, want 150", s.busy)
	}
}

func TestParseCPUSample_InvalidFormat(t *testing.T) {
	_, err := parseCPUSample("notcpu 1 2 3 4\n")
	if err == nil {
		t.Error("expected error for non-cpu line")
	}
}

func TestParseStatus(t *testing.T) {
	input := []byte(`Name:	bash
Umask:	0022
State:	S (sleeping)
Tgid:	1234
Pid:	1234
PPid:	1000
Uid:	1000	1000	1000	1000
VmRSS:	 5432 kB
`)
	st := parseStatus(input)
	if st.name != "bash" {
		t.Errorf("name: got %q, want %q", st.name, "bash")
	}
	if st.ppid != 1000 {
		t.Errorf("ppid: got %d, want 1000", st.ppid)
	}
	if st.uid != "1000" {
		t.Errorf("uid: got %q, want %q", st.uid, "1000")
	}
	if st.rssKB != 5432 {
		t.Errorf("rssKB: got %d, want 5432", st.rssKB)
	}
}

func TestParseStatus_MissingVmRSS(t *testing.T) {
	// kernel threads often lack VmRSS; should return zero, not crash
	input := []byte(`Name:	kworker
PPid:	2
Uid:	0	0	0	0
`)
	st := parseStatus(input)
	if st.rssKB != 0 {
		t.Errorf("rssKB: got %d, want 0", st.rssKB)
	}
	if st.name != "kworker" {
		t.Errorf("name: got %q, want %q", st.name, "kworker")
	}
}

func TestParseCmdline(t *testing.T) {
	cases := []struct {
		input []byte
		want  string
	}{
		{[]byte("/usr/bin/bash\x00-l\x00"), "/usr/bin/bash -l"},
		{[]byte{}, ""},
		{[]byte("/usr/bin/go\x00run\x00.\x00"), "/usr/bin/go run ."},
		{[]byte("bash"), "bash"}, // single arg, no trailing NUL
	}
	for _, c := range cases {
		got := parseCmdline(c.input)
		if got != c.want {
			t.Errorf("parseCmdline(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestParseStatTicks(t *testing.T) {
	// real format: pid (comm with spaces) state ppid pgroup session tty tpgid flags minflt cminflt majflt cmajflt utime stime ...
	input := []byte("1234 (my (proc)) S 1000 1234 1234 0 -1 4194560 100 0 0 0 42 17 0 0 20 0 1 0 12345 1234567 300 18446744073709551615 1 1 0 0 0 0 0 0 0 0 0 17 0 0 0 0 0 0")
	utime, stime, err := parseStatTicks(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if utime != 42 {
		t.Errorf("utime: got %d, want 42", utime)
	}
	if stime != 17 {
		t.Errorf("stime: got %d, want 17", stime)
	}
}

func TestParseLoadAvg(t *testing.T) {
got, err := parseLoadAvg("0.52 0.73 0.88 3/412 12345")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
want := domain.LoadAvg{Avg1m: 0.52, Avg5m: 0.73, Avg15m: 0.88, Running: 3, Total: 412}
if got != want {
t.Errorf("got %+v, want %+v", got, want)
}
}

func TestParseLoadAvg_TooFewFields(t *testing.T) {
_, err := parseLoadAvg("0.52 0.73 0.88")
if err == nil {
t.Error("expected error for too few fields")
}
}

func TestParseLoadAvg_MalformedTasks(t *testing.T) {
_, err := parseLoadAvg("0.52 0.73 0.88 notvalid 12345")
if err == nil {
t.Error("expected error for malformed tasks field")
}
}

func TestParseFileNR(t *testing.T) {
got, err := parseFileNR("1024\t0\t9223372036854775807")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if got.Open != 1024 {
t.Errorf("Open: got %d, want 1024", got.Open)
}
if got.Max != 9223372036854775807 {
t.Errorf("Max: got %d, want 9223372036854775807", got.Max)
}
}

func TestParseFileNR_TooFewFields(t *testing.T) {
_, err := parseFileNR("1024 0")
if err == nil {
t.Error("expected error for too few fields")
}
}

func TestParseSockStat(t *testing.T) {
input := `sockets: used 123
TCP: inuse 31 orphan 0 tw 5 alloc 33 mem 5
UDP: inuse 7 mem 2
UDPLITE: inuse 0
RAW: inuse 1
FRAG: inuse 0 memory 0
`
got, err := parseSockStat(strings.NewReader(input))
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if got.TCPUsed != 31 {
t.Errorf("TCPUsed: got %d, want 31", got.TCPUsed)
}
if got.TCPOrphan != 0 {
t.Errorf("TCPOrphan: got %d, want 0", got.TCPOrphan)
}
if got.TCPTimeWait != 5 {
t.Errorf("TCPTimeWait: got %d, want 5", got.TCPTimeWait)
}
if got.UDPUsed != 7 {
t.Errorf("UDPUsed: got %d, want 7", got.UDPUsed)
}
if got.RAWUsed != 1 {
t.Errorf("RAWUsed: got %d, want 1", got.RAWUsed)
}
}

func TestParseSockStat_MissingProtocol(t *testing.T) {
// only TCP present; UDP and RAW fields should be zero, no error
input := "TCP: inuse 10 orphan 0 tw 2 alloc 10 mem 1\n"
got, err := parseSockStat(strings.NewReader(input))
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if got.UDPUsed != 0 {
t.Errorf("UDPUsed: got %d, want 0", got.UDPUsed)
}
}

func TestParseVMStat(t *testing.T) {
input := `pgfault 100
pgmajfault 2
pswpin 0
pswpout 3
pgpgin 500
pgpgout 800
some_other_key 999
`
got, err := parseVMStat(strings.NewReader(input))
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if got.PageFaults != 100 {
t.Errorf("PageFaults: got %d, want 100", got.PageFaults)
}
if got.MajorPageFaults != 2 {
t.Errorf("MajorPageFaults: got %d, want 2", got.MajorPageFaults)
}
if got.SwapIn != 0 {
t.Errorf("SwapIn: got %d, want 0", got.SwapIn)
}
if got.SwapOut != 3 {
t.Errorf("SwapOut: got %d, want 3", got.SwapOut)
}
if got.PageIn != 500 {
t.Errorf("PageIn: got %d, want 500", got.PageIn)
}
if got.PageOut != 800 {
t.Errorf("PageOut: got %d, want 800", got.PageOut)
}
}

func TestParseVMStat_MissingKeys(t *testing.T) {
// keys not present should be zero, not an error
got, err := parseVMStat(strings.NewReader("unrelated_key 42\n"))
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if got.PageFaults != 0 {
t.Errorf("PageFaults: got %d, want 0", got.PageFaults)
}
}

func TestParseDiskStats(t *testing.T) {
input := `   8       0 sda 1000 0 10000 500 2000 0 20000 1000 0 800 0 0 0 300
   8       1 sda1 500 0 5000 250 1000 0 10000 500 0 400 0 0 0 150
   7       0 loop0 100 0 1000 50 200 0 2000 100 0 80 0 0 0 30
 252       0 dm-0 800 0 8000 400 1600 0 16000 800 0 640 0 0 0 240
`
stats, err := parseDiskStats(strings.NewReader(input))
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
// loop0 should be filtered out
devices := make(map[string]domain.DiskStat)
for _, s := range stats {
devices[s.Device] = s
}
if _, ok := devices["loop0"]; ok {
t.Error("loop0 should be filtered out")
}
if _, ok := devices["sda"]; !ok {
t.Error("sda should be present")
}
if _, ok := devices["dm-0"]; !ok {
t.Error("dm-0 should be present")
}
sda := devices["sda"]
if sda.ReadsTotal != 1000 {
t.Errorf("sda ReadsTotal: got %d, want 1000", sda.ReadsTotal)
}
if sda.ReadBytes != 10000*512 {
t.Errorf("sda ReadBytes: got %d, want %d", sda.ReadBytes, 10000*512)
}
if sda.IOTimeMs != 800 {
t.Errorf("sda IOTimeMs: got %d, want 800", sda.IOTimeMs)
}
}

func TestParseDiskStats_TooFewFields(t *testing.T) {
// lines with fewer than 14 fields must be skipped without error
input := "   8       0 sda 1000 0 10000\n"
stats, err := parseDiskStats(strings.NewReader(input))
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(stats) != 0 {
t.Errorf("expected 0 stats, got %d", len(stats))
}
}

func TestParseNetDev(t *testing.T) {
input := `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo:  100000     500    0    0    0     0          0         0   100000     500    0    0    0     0       0          0
  eth0: 2000000   10000   1    2    0     0          0         0  1000000    5000    0    0    0     0       0          0
`
devs, err := parseNetDev(strings.NewReader(input))
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
byName := make(map[string]domain.NetDev)
for _, d := range devs {
byName[d.Interface] = d
}
if len(byName) != 2 {
t.Fatalf("expected 2 interfaces, got %d", len(byName))
}
eth0 := byName["eth0"]
if eth0.RxBytes != 2000000 {
t.Errorf("eth0 RxBytes: got %d, want 2000000", eth0.RxBytes)
}
if eth0.TxBytes != 1000000 {
t.Errorf("eth0 TxBytes: got %d, want 1000000", eth0.TxBytes)
}
if eth0.RxErrors != 1 {
t.Errorf("eth0 RxErrors: got %d, want 1", eth0.RxErrors)
}
if eth0.RxDrops != 2 {
t.Errorf("eth0 RxDrops: got %d, want 2", eth0.RxDrops)
}
}

func TestParseUptime(t *testing.T) {
	got, err := parseUptime("123456.78 98765.43")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 123456.78 {
		t.Errorf("got %f, want 123456.78", got)
	}
}

func TestParseUptime_InvalidFormat(t *testing.T) {
	if _, err := parseUptime(""); err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseLastUpgradeFromHistory(t *testing.T) {
	content := `Start-Date: 2026-05-20  10:00:00
Commandline: apt install curl
Requested-By: debian (1000)
Install: curl:amd64 (1.0)
End-Date: 2026-05-20  10:00:10

Start-Date: 2026-05-26  12:37:51
Commandline: apt upgrade -y
Requested-By: debian (1000)
Upgrade: bash:amd64 (5.2, 5.3)
End-Date: 2026-05-26  12:39:04

Start-Date: 2026-05-26  12:39:23
Commandline: apt autoremove
Requested-By: debian (1000)
Remove: old-pkg:amd64 (1.0)
End-Date: 2026-05-26  12:39:25
`
	f, err := os.CreateTemp("", "history")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()
	defer os.Remove(f.Name())

	ts, err := parseLastUpgradeFromHistory(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts == 0 {
		t.Error("expected non-zero upgrade timestamp")
	}
	// autoremove End-Date must not override the upgrade timestamp
	upgradeTs, _ := parseLastUpgradeFromHistory(f.Name())
	if upgradeTs == 0 {
		t.Error("upgrade timestamp should not be zero")
	}
}

func TestParseLastUpgradeFromHistory_NoUpgrade(t *testing.T) {
	content := `Start-Date: 2026-05-20  10:00:00
Commandline: apt install curl
End-Date: 2026-05-20  10:00:10
`
	f, err := os.CreateTemp("", "history")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()
	defer os.Remove(f.Name())

	ts, err := parseLastUpgradeFromHistory(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts != 0 {
		t.Errorf("expected 0 when no upgrade present, got %d", ts)
	}
}

func TestReadCgroupUnit(t *testing.T) {
cases := []struct {
content string
want    string
}{
{"0::/system.slice/cups.service\n", "cups.service"},
{"0::/user.slice/user-1000.slice/session-2.scope\n", "session-2.scope"},
{"0::/\n", "(untracked)"},
{"0::\n", "(untracked)"},
// cgroup v1 lines only — no "0::" line
{"12:memory:/system.slice/foo\n", "(untracked)"},
}
for _, c := range cases {
// write temp file since readCgroupUnit takes a path
f, err := os.CreateTemp("", "cgroup")
if err != nil {
t.Fatal(err)
}
f.WriteString(c.content)
f.Close()
got, err := readCgroupUnit(f.Name())
os.Remove(f.Name())
if err != nil {
t.Errorf("content %q: unexpected error: %v", c.content, err)
continue
}
if got != c.want {
t.Errorf("content %q: got %q, want %q", c.content, got, c.want)
}
}
}
