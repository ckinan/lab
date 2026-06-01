package proc_test

import (
	"errors"
	"testing"

	"github.com/ckinan/lab/libs/proc"
)

type MockMemoryReader struct {
	Memory proc.Memory
	Err    error
}

func (m MockMemoryReader) ReadMemory() (proc.Memory, error) {
	return m.Memory, m.Err
}

type MockProcessReader struct {
	Processes []proc.Process
	Err       error
}

func (m MockProcessReader) ReadProcesses() ([]proc.Process, error) {
	return m.Processes, m.Err
}

type MockCPUReader struct {
	CPU float64
	Err error
}

func (m MockCPUReader) ReadCPU() (float64, error) {
	return m.CPU, m.Err
}

func TestCollector_Collect(t *testing.T) {
	tests := []struct {
		name       string
		memReader  proc.MemoryReader
		procReader proc.ProcessReader
		cpuReader  proc.CPUReader
		wantErr    bool
		wantCPU    float64
	}{
		{
			name:       "returns full snapshot on success",
			memReader:  MockMemoryReader{Memory: proc.Memory{Total: 1000, Used: 500}},
			procReader: MockProcessReader{Processes: []proc.Process{{Pid: 1}}},
			cpuReader:  MockCPUReader{CPU: 42.5},
			wantErr:    false,
			wantCPU:    42.5,
		},
		{
			name:       "returns error when memory reader fails",
			memReader:  MockMemoryReader{Err: errors.New("disk error")},
			procReader: MockProcessReader{},
			cpuReader:  MockCPUReader{},
			wantErr:    true,
		},
		{
			name:       "returns error when process reader fails",
			memReader:  MockMemoryReader{Memory: proc.Memory{Total: 1000}},
			procReader: MockProcessReader{Err: errors.New("proc error")},
			cpuReader:  MockCPUReader{},
			wantErr:    true,
		},
		{
			name:       "CPU error is non-fatal, returns 0% instead",
			memReader:  MockMemoryReader{Memory: proc.Memory{Total: 1000}},
			procReader: MockProcessReader{},
			cpuReader:  MockCPUReader{Err: errors.New("cpu read failed")},
			wantErr:    false,
			wantCPU:    0, // degraded gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := proc.NewCollector(tt.memReader, tt.procReader, tt.cpuReader)
			snap, err := c.Collect()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if snap.CPU != tt.wantCPU {
				t.Errorf("CPU = %.2f, want %.2f", snap.CPU, tt.wantCPU)
			}
		})
	}
}
