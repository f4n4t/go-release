package xrel_test

import (
	"testing"

	"github.com/f4n4t/go-release/pkg/xrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Get(t *testing.T) {
	t.Skip("skipping test")
	tests := []struct {
		name        string
		rlsName     string
		wantRelease xrel.Release
		wantErr     error
	}{
		{
			name:    "[PreXrel] John.Wick.Kapitel.4.2023.German.DL.AC3.Dubbed.1080p.WEB.H264-PsO",
			rlsName: "John.Wick.Kapitel.4.2023.German.DL.AC3.Dubbed.1080p.WEB.H264-PsO",
			wantRelease: xrel.Release{
				Dirname:   "John.Wick.Kapitel.4.2023.German.DL.AC3.Dubbed.1080p.WEB.H264-PsO",
				GroupName: "PsO",
				ExtInfo:   xrel.ExtInfo{Type: "movie"},
				Time:      1684807815,
			},
		},
		{
			name:        "[PreXrel] Does not exists",
			rlsName:     "this.release.does.not.exist.asdf",
			wantRelease: xrel.Release{},
			wantErr:     xrel.ErrNothingFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := xrel.Get(tt.rlsName)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.wantRelease.Dirname, actual.Dirname)
			assert.Equal(t, tt.wantRelease.GroupName, actual.GroupName)
			assert.Equal(t, tt.wantRelease.ExtInfo.Type, actual.ExtInfo.Type)
			assert.Equal(t, tt.wantRelease.Time, actual.Time)
		})
	}
}
