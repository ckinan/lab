package proc

import "fmt"

type Collector struct {
	mem  MemoryReader
	proc ProcessReader
	cpu  CPUReader
}

func NewCollector(mem MemoryReader, proc ProcessReader, cpu CPUReader) *Collector {
	return &Collector{
		mem:  mem,
		proc: proc,
		cpu:  cpu,
	}
}

func (c *Collector) Collect() (Snapshot, error) {
	mem, err := c.mem.ReadMemory()
	if err != nil {
		return Snapshot{}, fmt.Errorf("reading memory: %w", err)
	}

	processes, err := c.proc.ReadProcesses()
	if err != nil {
		return Snapshot{}, fmt.Errorf("reading processes: %w", err)
	}

	cpu, err := c.cpu.ReadCPU()
	if err != nil {
		cpu = 0 // non-fatal: show 0% rather than crash
	}

	return Snapshot{
		CPU:       cpu,
		Memory:    mem,
		Processes: processes,
	}, nil
}
