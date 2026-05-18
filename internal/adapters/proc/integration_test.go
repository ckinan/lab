//go:build integration

package proc_test

import (
	"testing"

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
