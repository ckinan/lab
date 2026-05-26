package proc

import (
	"bufio"
	"fmt"
	"os"
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

	ts, err := parseLastUpgradeFromHistory(aptHistoryLog)
	if err != nil {
		return info, fmt.Errorf("parse apt history: %w", err)
	}
	info.LastUpgradeUnix = ts

	return info, nil
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

	var lastUpgrade int64
	var inUpgradeBlock bool

	scanner := bufio.NewScanner(f)
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
