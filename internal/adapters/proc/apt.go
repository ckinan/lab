package proc

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ckinan/lab/internal/domain"
)

const aptHistoryLog = "/var/log/apt/history.log"
const aptListsDir = "/var/lib/apt/lists"

type AptReader struct{}

func (r AptReader) ReadAptInfo() (domain.AptInfo, error) {
	info := domain.AptInfo{}

	if fi, err := os.Stat(aptListsDir); err == nil {
		info.LastUpdateUnix = fi.ModTime().Unix()
	}

	ts, err := lastUpgradeTimestamp()
	if err != nil {
		return info, fmt.Errorf("parse apt history: %w", err)
	}
	info.LastUpgradeUnix = ts

	return info, nil
}

// lastUpgradeTimestamp returns the most recent "apt upgrade" end timestamp
// across the current history.log and all rotated history.log.*.gz files.
func lastUpgradeTimestamp() (int64, error) {
	var latest int64

	ts, err := parseLastUpgradeFromHistory(aptHistoryLog)
	if err != nil {
		return 0, err
	}
	if ts > latest {
		latest = ts
	}

	gz, _ := filepath.Glob(aptHistoryLog + ".*.gz")
	for _, path := range gz {
		ts, err := parseLastUpgradeFromHistoryGZ(path)
		if err != nil {
			continue
		}
		if ts > latest {
			latest = ts
		}
	}

	return latest, nil
}

func parseLastUpgradeFromHistory(path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	return scanUpgradeTimestamp(f)
}

func parseLastUpgradeFromHistoryGZ(path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return 0, fmt.Errorf("gzip %s: %w", path, err)
	}
	defer gr.Close()

	return scanUpgradeTimestamp(gr)
}

func scanUpgradeTimestamp(r io.Reader) (int64, error) {
	var lastUpgrade int64
	var inUpgradeBlock bool

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "Commandline:"):
			cmd := strings.TrimSpace(strings.TrimPrefix(line, "Commandline:"))
			inUpgradeBlock = strings.Contains(cmd, "upgrade") && !strings.Contains(cmd, "autoremove")
		case strings.HasPrefix(line, "End-Date:") && inUpgradeBlock:
			raw := strings.Join(strings.Fields(strings.TrimPrefix(line, "End-Date:")), " ")
			if t, err := time.ParseInLocation("2006-01-02 15:04:05", raw, time.Local); err == nil {
				lastUpgrade = t.Unix()
			}
			inUpgradeBlock = false
		case line == "":
			inUpgradeBlock = false
		}
	}

	return lastUpgrade, scanner.Err()
}
