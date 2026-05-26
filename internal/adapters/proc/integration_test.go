//go:build integration

package proc_test

import (
	"testing"
	"time"

	"github.com/ckinan/cktop/internal/adapters/proc"
)

func TestProcMemoryReader_RealSystem(t *testing.T) {
	r := proc.ProcMemoryReader{}
	mem, err := r.ReadMemory()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mem.Total == 0 {
		t.Error("Total memory should be > 0 on a real machine")
	}
	if mem.Used == 0 {
		t.Error("Used memory should be > 0 on a real machine")
	}
}

func TestProcCPUReader_RealSystem(t *testing.T) {
	r := proc.NewProcCPUReader()
	// first call seeds the baseline - always returns 0
	_, err := r.ReadCPU()
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	// second call returns the actual delta
	cpu, err := r.ReadCPU()
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if cpu < 0 || cpu > 100 {
		t.Errorf("CPU percent should be between 0-100, got %.2f", cpu)
	}
}

func TestProcProcessReader_RealSystem(t *testing.T) {
	r := proc.NewProcProcessReader()
	procs, err := r.ReadProcesses()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(procs) == 0 {
		t.Error("expected at least one process on a real machine")
	}
}

func TestProcLoadAvgReader_RealSystem(t *testing.T) {
r := proc.ProcLoadAvgReader{}
v, err := r.ReadLoadAvg()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if v.Avg1m < 0 {
t.Errorf("Avg1m should be >= 0, got %f", v.Avg1m)
}
if v.Total == 0 {
t.Error("Total tasks should be > 0 on a real machine")
}
}

func TestProcFileNRReader_RealSystem(t *testing.T) {
r := proc.ProcFileNRReader{}
v, err := r.ReadFileNR()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if v.Open == 0 {
t.Error("Open fds should be > 0 on a real machine")
}
if v.Max == 0 {
t.Error("Max fds should be > 0")
}
}

func TestProcSockStatReader_RealSystem(t *testing.T) {
r := proc.ProcSockStatReader{}
_, err := r.ReadSockStat()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
}

func TestProcVMStatReader_RealSystem(t *testing.T) {
r := proc.ProcVMStatReader{}
v, err := r.ReadVMStat()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if v.PageFaults == 0 {
t.Error("pgfault should be > 0 on a running machine")
}
}

func TestProcDiskStatsReader_RealSystem(t *testing.T) {
r := proc.ProcDiskStatsReader{}
stats, err := r.ReadDiskStats()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(stats) == 0 {
t.Error("expected at least one disk device")
}
}

func TestProcNetDevReader_RealSystem(t *testing.T) {
r := proc.ProcNetDevReader{}
devs, err := r.ReadNetDev()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(devs) == 0 {
t.Error("expected at least one network interface")
}
}

func TestProcCPUCoresReader_RealSystem(t *testing.T) {
r := proc.NewProcCPUCoresReader()
// first call seeds baseline
_, err := r.ReadCPUCores()
if err != nil {
t.Fatalf("unexpected error on first call: %v", err)
}
time.Sleep(100 * time.Millisecond)
cores, err := r.ReadCPUCores()
if err != nil {
t.Fatalf("unexpected error on second call: %v", err)
}
if len(cores) == 0 {
t.Error("expected at least one CPU core")
}
for name, pct := range cores {
if pct < 0 || pct > 100 {
t.Errorf("core %s: percent out of range: %.2f", name, pct)
}
}
}

func TestProcServiceStatsReader_RealSystem(t *testing.T) {
r := proc.NewProcServiceStatsReader()
// first call seeds baseline
_, err := r.ReadServiceStats()
if err != nil {
t.Fatalf("unexpected error on first call: %v", err)
}
stats, err := r.ReadServiceStats()
if err != nil {
t.Fatalf("unexpected error on second call: %v", err)
}
if len(stats) == 0 {
t.Error("expected at least one service/unit")
}
}
