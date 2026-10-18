package release_test

import (
	"archive/zip"
	"os"
	"path/filepath"
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

func TestRelease_CheckZip_ExtractNFO(t *testing.T) {
	releaseDir := filepath.Join(t.TempDir(), "Zipped.NFO.Release-Group")
	require.NoError(t, os.Mkdir(releaseDir, 0o755))

	writeZip := func(name string, files map[string]string) {
		f, err := os.Create(filepath.Join(releaseDir, name))
		require.NoError(t, err)
		defer f.Close()

		zw := zip.NewWriter(f)
		for fileName, content := range files {
			w, err := zw.Create(fileName)
			require.NoError(t, err)
			_, err = w.Write([]byte(content))
			require.NoError(t, err)
		}
		require.NoError(t, zw.Close())
	}

	writeZip("testa.zip", map[string]string{
		"test.rar":    "rar-content",
		"file_id.diz": "[01/02]",
		"test.nfo":    "nfo-content",
	})
	writeZip("testb.zip", map[string]string{
		"test.r00":    "rar-content",
		"file_id.diz": "[02/02]",
	})

	releaseService := release.NewServiceBuilder().WithSkipPre(true).WithSkipMediaInfo(true).Build()
	rel, err := releaseService.Parse(releaseDir)
	require.NoError(t, err)

	require.NoError(t, releaseService.CheckZip(rel, true))
	require.NotNil(t, rel.NFO)
	assert.Equal(t, "test.nfo", rel.NFO.Name)
	assert.Equal(t, []byte("nfo-content"), rel.NFO.Content)
}
