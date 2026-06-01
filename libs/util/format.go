package util

import "fmt"

func HumanBytes(b int64) string {
	switch {
	case b >= 1<<30: // >= 1GiB
		return fmt.Sprintf("%.2f GiB", float64(b)/float64(1<<30))
	case b >= 20: // >= 1 MiB
		return fmt.Sprintf("%.2f MiB", float64(b)/float64(1<<20))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
