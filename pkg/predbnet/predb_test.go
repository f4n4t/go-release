package predbnet_test

import (
	"testing"
	"time"

	"github.com/f4n4t/go-release/pkg/predbnet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Get(t *testing.T) {
	t.Skip("skipping test")
	tests := []struct {
		name        string
		rlsName     string
		wantRelease predbnet.Release
		wantErr     error
	}{
		{
			name:    "[PreNet] Cable.Guy.Die.Nervensaege.1996.GERMAN.AC3D.DL.1080p.BluRay.x264-TVP",
			rlsName: "Cable.Guy.Die.Nervensaege.1996.GERMAN.AC3D.DL.1080p.BluRay.x264-TVP",
			wantRelease: predbnet.Release{
				Release: "Cable.Guy.Die.Nervensaege.1996.GERMAN.AC3D.DL.1080p.BluRay.x264-TVP",
				Group:   "TVP",
				Section: "X264-HD",
				PreTime: 1302295015,
			},
		},
		{
			name:    "[PreNet] Survivor.Cesko.a.Slovensko.S01E07.CZECH.1080p.WEB.H264-SDTV",
			rlsName: "Survivor.Cesko.a.Slovensko.S01E07.CZECH.1080p.WEB.H264-SDTV",
			wantRelease: predbnet.Release{
				Release: "Survivor.Cesko.a.Slovensko.S01E07.CZECH.1080p.WEB.H264-SDTV",
				Group:   "SDTV",
				Section: "TV-WEB-HD-X264",
				PreTime: 1685896713,
			},
		},
		{
			name:        "[PreNet] Empty name",
			rlsName:     "",
			wantRelease: predbnet.Release{},
			wantErr:     predbnet.ErrEmptyName,
		},
		{
			name:        "[PreNet] Does not exists",
			rlsName:     "this.release.does.not.exist.asdf",
			wantRelease: predbnet.Release{},
			wantErr:     predbnet.ErrNothingFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := predbnet.Get(tt.rlsName)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.wantRelease.Release, actual.Release)
			assert.Equal(t, tt.wantRelease.Group, actual.Group)
			assert.Equal(t, tt.wantRelease.Section, actual.Section)
			assert.Equal(t, tt.wantRelease.PreTime, actual.PreTime)
			time.Sleep(2 * time.Second)
		})
	}
}
