package utils_test

import (
	"testing"

	"github.com/f4n4t/go-release/pkg/utils"
)

func TestConvertByteSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{"test 10B", 10, "10 B"},
		{"test 1MB", 1048576, "1.0 MiB"},
		{"test 1GB", 1073741824, "1.0 GiB"},
		{"test 10.2GB", 10952166605, "10.2 GiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.Bytes(tt.size)
			if got != tt.want {
				t.Errorf("convert.ConvertByteSize() expected %s, got %s", tt.want, got)
			}
		})
	}
}
