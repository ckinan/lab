package util_test

import (
	"testing"

	"github.com/ckinan/lab/internal/util"
)

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{"bytes", 15, "15 B"},
		{"mebibytes", 1 << 20, "1.00 MiB"},
		{"gibibytes", 1 << 30, "1.00 GiB"},
		{"zero", 0, "0 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := util.HumanBytes(tt.input)
			if got != tt.want {
				t.Errorf("HumanBytes(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
