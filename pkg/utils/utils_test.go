package utils_test

import (
	"testing"

	"github.com/f4n4t/go-release/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetJSONString(t *testing.T) {
	tests := []struct {
		name       string
		obj        any
		expected   string
		withIndent bool
	}{
		{"nil", nil, "null", true},
		{"empty object", struct{}{}, "{}", true},
		{"simple object", struct{ Name string }{"test"}, "{\n\t\"Name\": \"test\"\n}", true},
		{"nil with jsonLog", nil, "null", false},
		{"empty object with jsonLog", struct{}{}, "{}", false},
		{"simple object with jsonLog", struct{ Name string }{"test"}, "{\"Name\":\"test\"}", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := utils.GetJSONString(tt.obj, tt.withIndent)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
