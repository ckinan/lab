package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ckinan/cktop/internal/adapters/proc"
	"github.com/ckinan/cktop/internal/domain"
	"github.com/ckinan/cktop/internal/util"
)

func main() {
	output := flag.String("o", "text", "output format: text or json")
	flag.Parse()

	mem, err := proc.ProcMemoryReader{}.ReadMemory()
	if err != nil {
		log.Fatal(err)
	}

	switch *output {
	case "json":
		printJSON(mem)
	default:
		printText(mem)
	}
}

func printText(m domain.Memory) {
	usedPct := float64(m.Used) / float64(m.Total) * 100
	availPct := float64(m.Available) / float64(m.Total) * 100

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

func printJSON(m domain.Memory) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		log.Fatal(err)
	}
}
