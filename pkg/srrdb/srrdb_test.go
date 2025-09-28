package srrdb_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"testing"

	"github.com/f4n4t/go-release/pkg/srrdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_SrrContent(t *testing.T) {
	//t.Skip()
	testFile := "testfiles/test3.srr"

	srr, err := srrdb.LoadFromFile(testFile)
	require.NoError(t, err)

	for _, p := range srr.StoredFiles {
		fmt.Printf("%s\n", p.Path)
	}

	for _, p := range srr.PackedFiles {
		fmt.Printf("%s\n", p.Path)
	}

	fmt.Println()

	testFile = "testfiles/omg.srr"
	srr, err = srrdb.LoadFromFile(testFile)
	require.NoError(t, err)

	for _, p := range srr.StoredFiles {
		fmt.Printf("%s\n", p.Path)
	}

	fmt.Println()

	for _, p := range srr.PackedFiles {
		fmt.Printf("%s\n", p.Path)
	}
}

func TestService_GetInformation(t *testing.T) {
	tests := []struct {
		name        string
		release     string
		expected    srrdb.Release
		expectedErr bool
	}{
		{
			name:    "Existing Release",
			release: "Notfall.Krankenhaus.Kliniken.vor.dem.Finanzkollaps.German.DOKU.1080p.WEB.H264-UTOPiA",
			expected: srrdb.Release{
				Name: "Notfall.Krankenhaus.Kliniken.vor.dem.Finanzkollaps.German.DOKU.1080p.WEB.H264-UTOPiA",
				Files: []srrdb.File{
					{Name: "utopia-nk-web1080p.nfo", Size: 7360, CRC: "C334A5FF"},
					{Name: "Sample/utopia-nk-web1080p-sample.mkv", Size: 26973666, CRC: "DD9FB3D5"},
					{Name: "utopia-nk-web1080p.sfv", Size: 363, CRC: "290424B1"},
					{Name: "utopia-nk-web1080p.rar", Size: 100000000, CRC: "8BDCCFFB"},
					{Name: "utopia-nk-web1080p.r00", Size: 100000000, CRC: "84309539"},
					{Name: "utopia-nk-web1080p.r01", Size: 100000000, CRC: "43EF8598"},
					{Name: "utopia-nk-web1080p.r02", Size: 100000000, CRC: "D9C35249"},
					{Name: "utopia-nk-web1080p.r03", Size: 100000000, CRC: "8402A3A8"},
					{Name: "utopia-nk-web1080p.r04", Size: 100000000, CRC: "01E88EA0"},
					{Name: "utopia-nk-web1080p.r05", Size: 100000000, CRC: "2681167C"},
					{Name: "utopia-nk-web1080p.r06", Size: 100000000, CRC: "617844C0"},
					{Name: "utopia-nk-web1080p.r07", Size: 100000000, CRC: "805DA3E8"},
					{Name: "utopia-nk-web1080p.r08", Size: 100000000, CRC: "1B16E2D5"},
					{Name: "utopia-nk-web1080p.r09", Size: 57991226, CRC: "3017B362"},
				},
				ArchivedFiles: []srrdb.ArchivedFile{
					{Name: "utopia-nk-web1080p.mkv", Size: 1057990137, CRC: "AE8CF45D"},
				},
				Adds: []srrdb.Add{},
			},
		},
		{
			name:        "Non-existent Release",
			release:     "This.Release.Does.Not.Exist",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := srrdb.GetInformation(tt.release)
			if tt.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestService_GetFile(t *testing.T) {
	tests := []struct {
		name        string
		releaseName string
		fileName    string
		sha256Sum   string
		expectedErr bool
	}{
		{
			name:        "Existing file",
			releaseName: "Notfall.Krankenhaus.Kliniken.vor.dem.Finanzkollaps.German.DOKU.1080p.WEB.H264-UTOPiA",
			fileName:    "utopia-nk-web1080p.nfo",
			sha256Sum:   "ed78bbd8f7d04eafb35a7d8909686cac366fc44e8644dc531d819555add22f43",
		},
		{
			name:        "Non-existing file",
			releaseName: "Notfall.Krankenhaus.Kliniken.vor.dem.Finanzkollaps.German.DOKU.1080p.WEB.H264-UTOPiA",
			fileName:    "utopia-nk-web1080p.nfo32",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dl := srrdb.DownloadRelease{
				Name: tt.releaseName,
				File: tt.fileName,
			}
			data, err := srrdb.GetFile(dl)
			if tt.expectedErr {
				fmt.Println(err)
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			rdr := bytes.NewReader(data)

			h := sha256.New()
			if _, err := io.Copy(h, rdr); err != nil {
				t.Error(err)
				return
			}

			actual := fmt.Sprintf("%x", h.Sum(nil))

			assert.Equal(t, tt.sha256Sum, actual)
		})
	}
}
