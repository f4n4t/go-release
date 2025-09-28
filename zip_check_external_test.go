package release_test

import (
	"testing"

	"github.com/f4n4t/go-release"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestRelease_CheckZip(t *testing.T) {
	tests := []struct {
		name    string
		folder  string
		wantErr bool
	}{
		{"valid zip release", "testdata/Zipped.Release-Group", false},
		{"invalid zip release", "testdata/Zipped.Missing.File.Release-Group", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releaseService := release.NewServiceBuilder().WithSkipPre(true).WithSkipMediaInfo(true).Build()
			rel, err := releaseService.Parse(tt.folder)
			require.NoError(t, err)

			gotErr := releaseService.CheckZip(rel, false)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			assert.NoError(t, gotErr)
		})
	}
}
