package release_test

import (
	"os"
	"testing"

	"github.com/f4n4t/go-release"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNFO(t *testing.T) {
	tests := []struct {
		desc        string
		path        string
		expected    release.NFOFile
		expectedErr error
	}{
		{
			desc: "existing nfo",
			path: "testdata/with-nfo.mkv",
			expected: release.NFOFile{
				Name:    "test.nfo",
				Content: []byte("Das ist eine test nfo\nblib blab blub\n"),
			},
		},
		{
			desc:     "no nfo",
			path:     "testdata/without-nfo.mkv",
			expected: release.NFOFile{},
		},
		{
			desc:        "non-existing file",
			path:        "file-does-not-exist.mkv",
			expectedErr: os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := release.ParseNfoAttachment(tt.path)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, gotErr, tt.expectedErr)
				return
			}
			require.NoError(t, gotErr)

			assert.Equal(t, tt.expected, got)
		})
	}
}
