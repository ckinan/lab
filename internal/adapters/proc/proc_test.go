package proc

import (
	"strings"
	"testing"

	"github.com/ckinan/cktop/internal/domain"
)

func TestParseMeminfo(t *testing.T) {
	input := `MemTotal:       16384000 kB
MemFree:         2048000 kB
MemAvailable:    8192000 kB
Buffers:          512000 kB
`
	mem, err := parseMeminfo(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := domain.Memory{
		Total:     16384000 * 1024,
		Available: 8192000 * 1024,
		Used:      (16384000 - 8192000) * 1024,
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
