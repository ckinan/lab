package proc

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

)

// ProcPSIReader reads system-wide PSI from /proc/pressure/.
type ProcPSIReader struct{}

func (r ProcPSIReader) ReadPSI() (SystemPSI, error) {
	var psi SystemPSI
	var err error

	psi.CPU, err = readPSIResource("/proc/pressure/cpu")
	if err != nil {
		return psi, fmt.Errorf("read cpu pressure: %w", err)
	}
	psi.Memory, err = readPSIResource("/proc/pressure/memory")
	if err != nil {
		return psi, fmt.Errorf("read memory pressure: %w", err)
	}
	psi.IO, err = readPSIResource("/proc/pressure/io")
	if err != nil {
		return psi, fmt.Errorf("read io pressure: %w", err)
	}
	return psi, nil
}

// CgroupPSIReader reads per-service PSI from /sys/fs/cgroup/system.slice/*.service/.
type CgroupPSIReader struct{}

func (r CgroupPSIReader) ReadServicePSI() ([]ServicePSI, error) {
	matches, err := filepath.Glob("/sys/fs/cgroup/system.slice/*.service")
	if err != nil {
		return nil, fmt.Errorf("glob cgroup services: %w", err)
	}

	results := make([]ServicePSI, 0, len(matches))
	for _, dir := range matches {
		unit := filepath.Base(dir)
		svcPSI := ServicePSI{Unit: unit}
		svcPSI.CPU, _ = readPSIResource(filepath.Join(dir, "cpu.pressure"))
		svcPSI.Memory, _ = readPSIResource(filepath.Join(dir, "memory.pressure"))
		svcPSI.IO, _ = readPSIResource(filepath.Join(dir, "io.pressure"))
		results = append(results, svcPSI)
	}
	return results, nil
}

func readPSIResource(path string) (PSIResource, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return PSIResource{}, nil
		}
		return PSIResource{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var res PSIResource
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val := parsePSIValues(fields[1:])
		switch fields[0] {
		case "some":
			res.Some = val
		case "full":
			res.Full = val
		}
	}
	return res, scanner.Err()
}

func parsePSIValues(fields []string) PSIValue {
	var val PSIValue
	for _, f := range fields {
		kv := strings.SplitN(f, "=", 2)
		if len(kv) != 2 {
			continue
		}
		v, err := strconv.ParseFloat(kv[1], 64)
		if err != nil {
			continue
		}
		switch kv[0] {
		case "avg10":
			val.Avg10 = v
		case "avg60":
			val.Avg60 = v
		case "avg300":
			val.Avg300 = v
		}
	}
	return val
}
