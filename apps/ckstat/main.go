package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ckinan/lab/libs/proc"
	"github.com/ckinan/lab/libs/util"
)

type snapshot struct {
	CPU    float64      `json:"cpu_percent"`
	Memory proc.Memory `json:"memory"`
}

func main() {
	output   := flag.String("o", "text", "output format: text or json")
	interval := flag.Duration("interval", time.Second, "CPU measurement window (two /proc/stat reads)")
	flag.Parse()

	cpuReader := proc.NewProcCPUReader()
	if _, err := cpuReader.ReadCPU(); err != nil {
		log.Fatal(err)
	}
	time.Sleep(*interval)
	cpu, err := cpuReader.ReadCPU()
	if err != nil {
		log.Fatal(err)
	}

	mem, err := proc.ProcMemoryReader{}.ReadMemory()
	if err != nil {
		log.Fatal(err)
	}

	s := snapshot{CPU: cpu, Memory: mem}

	switch *output {
	case "json":
		printJSON(s)
	default:
		printText(s)
	}
}

func printText(s snapshot) {
	m := s.Memory
	usedPct := float64(m.Used) / float64(m.Total) * 100
	availPct := float64(m.Available) / float64(m.Total) * 100

	fmt.Printf("CPU\n")
	fmt.Printf("  %-12s %9.1f%%\n", "usage", s.CPU)
	fmt.Println()

	fmt.Println("Memory")
	fmt.Printf("  %-12s %10s\n", "total", util.HumanBytes(m.Total))
	fmt.Printf("  %-12s %10s  %.0f%%\n", "used", util.HumanBytes(m.Used), usedPct)
	fmt.Printf("  %-12s %10s  %.0f%%\n", "available", util.HumanBytes(m.Available), availPct)
	fmt.Printf("  %-12s %10s\n", "free", util.HumanBytes(m.Free))
	fmt.Printf("  %-12s %10s\n", "cached", util.HumanBytes(m.Cached))
	fmt.Printf("  %-12s %10s\n", "buffers", util.HumanBytes(m.Buffers))
	fmt.Printf("  %-12s %10s\n", "shared", util.HumanBytes(m.Shmem))

	fmt.Println()

	if m.SwapTotal > 0 {
		swapUsedPct := float64(m.SwapUsed) / float64(m.SwapTotal) * 100
		fmt.Println("Swap")
		fmt.Printf("  %-12s %10s\n", "total", util.HumanBytes(m.SwapTotal))
		fmt.Printf("  %-12s %10s  %.0f%%\n", "used", util.HumanBytes(m.SwapUsed), swapUsedPct)
		fmt.Printf("  %-12s %10s\n", "free", util.HumanBytes(m.SwapFree))
	}
}

func printJSON(s snapshot) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		log.Fatal(err)
	}
}
