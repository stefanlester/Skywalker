package filesystems_test

import (
	"testing"

	"github.com/stefanlester/skywalker/filesystems"
)

func TestSizeToMB(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want float64
	}{
		{"zero", 0, 0},
		{"one kilobyte", 1024, 0.0009765625},
		{"one megabyte", 1 << 20, 1},
		{"five megabytes", 5 << 20, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filesystems.SizeToMB(tt.size); got != tt.want {
				t.Errorf("SizeToMB(%d) = %v; want %v", tt.size, got, tt.want)
			}
		})
	}
}
