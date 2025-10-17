package release_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/f4n4t/go-dtree"
	"github.com/f4n4t/go-release"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func compareRelease(t *testing.T, expected, got release.Info) {
	assert.Equal(t, expected.Name, got.Name, "name mismatch")
	assert.Equal(t, expected.Group, got.Group, "group mismatch")
	assert.Equal(t, expected.Size, got.Size, "size mismatch")
	assert.Equal(t, expected.Extensions, got.Extensions, "extensions mismatch")
	assert.Equal(t, expected.Language, got.Language, "language mismatch")
	assert.Equal(t, expected.TagResolution, got.TagResolution, "resolution mismatch")
	if expected.BiggestFile != nil {
		require.NotNil(t, got.BiggestFile)
		assert.Equal(t, expected.BiggestFile.Info.Name, got.BiggestFile.Info.Name, "biggest file name mismatch")
		assert.Equal(t, expected.BiggestFile.Info.Size, got.BiggestFile.Info.Size, "biggest file size mismatch")
	}
	assert.Equal(t, expected.ProductTitle, got.ProductTitle, "product title mismatch")
	assert.Equal(t, expected.ProductYear, got.ProductYear, "product year mismatch")
	assert.Equal(t, expected.Section, got.Section, "section mismatch")
	assert.Equal(t, expected.ImdbID, got.ImdbID, "imdb id mismatch")
	if expected.NFO != nil {
		require.NotNil(t, got.NFO)
		assert.Equal(t, expected.NFO.Name, got.NFO.Name, "nfo name mismatch")
		assert.Equal(t, expected.NFO.Content, got.NFO.Content, "nfo content mismatch")
	}
	assert.Equal(t, expected.ArchiveCount, got.ArchiveCount, "archive count mismatch")
	assert.Equal(t, expected.SfvCount, got.SfvCount, "sfv count mismatch")
	compareEpisodes(t, expected.Episodes, got.Episodes)
}

func compareForbiddenFiles(t *testing.T, expected, got []release.ForbiddenFile) {
	assert.Equal(t, len(expected), len(got), "number of forbidden files mismatch")
	for i, e := range expected {
		assert.Equal(t, e.Info.Name, got[i].Info.Name, "forbidden file name mismatch")
		assert.Equal(t, e.Error, got[i].Error, "forbidden file error mismatch")
	}
}

func compareEpisodes(t *testing.T, expected, got []release.Episode) {
	assert.Equal(t, len(expected), len(got), "number of episodes mismatch")
	for i, e := range expected {
		assert.Equal(t, e.Number, got[i].Number, "episode number mismatch")
		assert.Equal(t, e.Name, got[i].Name, "episode name mismatch")
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		desc        string
		root        string
		testFiles   map[string][]byte
		ignore      []string
		expected    release.Info
		expectedErr error
	}{
		{
			desc: "unpacked release",
			root: "Unpacked.1967.German.1080p.BluRay.x264-Group",
			testFiles: map[string][]byte{
				"Unpacked.1967.German.1080p.BluRay.x264-Group/unpacked-group.mkv":      []byte("should.be.the.biggest.file.here\n"),
				"Unpacked.1967.German.1080p.BluRay.x264-Group/unpacked-group.nfo":      []byte("imdb.com/title/tt0123456\n"),
				"Unpacked.1967.German.1080p.BluRay.x264-Group/Subs/unpacked-group.idx": []byte("ab\n"),
				"Unpacked.1967.German.1080p.BluRay.x264-Group/Subs/unpacked-group.sub": []byte("ab\n"),
			},
			expected: release.Info{
				Name:  "Unpacked.1967.German.1080p.BluRay.x264-Group",
				Group: "Group",
				Size:  32 + 25 + 3 + 3, // total file size
				Extensions: map[string]int{
					".mkv": 1,
					".nfo": 1,
					".idx": 1,
					".sub": 1,
				},
				Language:      "german",
				TagResolution: release.FHD,
				BiggestFile: &dtree.Node{
					Info: &dtree.FileInfo{
						Name: "unpacked-group.mkv",
						Size: 32,
					},
				},
				ProductTitle: "Unpacked",
				ProductYear:  1967,
				Section:      release.Movies,
				ImdbID:       123456,
				NFO: &release.NFOFile{
					Name:    "unpacked-group.nfo",
					Content: []byte("imdb.com/title/tt0123456\n"),
				},
			},
		},
		{
			desc: "tv pack release",
			root: "TVPack.1967.S01.German.1080p.BluRay.x264-Group",
			testFiles: map[string][]byte{
				"TVPack.1967.S01.German.1080p.BluRay.x264-Group/s01e01-group.mkv": []byte("abcde"),
				"TVPack.1967.S01.German.1080p.BluRay.x264-Group/s01e02-group.mkv": []byte("abcd"),
				"TVPack.1967.S01.German.1080p.BluRay.x264-Group/s01e03-group.mkv": []byte("abc"),
				"TVPack.1967.S01.German.1080p.BluRay.x264-Group/meta.jpg":         []byte("ab"),
			},
			expected: release.Info{
				Name:  "TVPack.1967.S01.German.1080p.BluRay.x264-Group",
				Group: "Group",
				Size:  5 + 4 + 3 + 2, // total file size
				Extensions: map[string]int{
					".mkv": 3,
					".jpg": 1,
				},
				Language:      "german",
				TagResolution: release.FHD,
				BiggestFile: &dtree.Node{
					Info: &dtree.FileInfo{
						Name: "s01e01-group.mkv",
						Size: 5,
					},
				},
				ProductTitle: "TVPack",
				ProductYear:  1967,
				Section:      release.TVPack,
				Episodes: []release.Episode{
					{Number: 1, Name: "s01e01-group.mkv"},
					{Number: 2, Name: "s01e02-group.mkv"},
					{Number: 3, Name: "s01e03-group.mkv"},
				},
			},
		},
		{
			desc: "tv release hidden as pack",
			root: "TV.1967.S01.German.1080p.BluRay.x264-Group",
			testFiles: map[string][]byte{
				"TV.1967.S01.German.1080p.BluRay.x264-Group/s01e01-group.mkv": []byte("abcde"),
				"TV.1967.S01.German.1080p.BluRay.x264-Group/meta.jpg":         []byte("ab"),
			},
			expected: release.Info{
				Name:  "TV.1967.S01.German.1080p.BluRay.x264-Group",
				Group: "Group",
				Size:  5 + 2, // total file size
				Extensions: map[string]int{
					".mkv": 1,
					".jpg": 1,
				},
				Language:      "german",
				TagResolution: release.FHD,
				BiggestFile: &dtree.Node{
					Info: &dtree.FileInfo{
						Name: "s01e01-group.mkv",
						Size: 5,
					},
				},
				ProductTitle: "TV",
				ProductYear:  1967,
				Section:      release.TV,
				Episodes: []release.Episode{
					{Number: 1, Name: "s01e01-group.mkv"},
				},
			},
		},
		{
			desc: "tv retail release hidden as pack",
			root: "TV.1967.S01.German.AC3.COMPLETE.DVD.MPEG2-GROUP",
			testFiles: map[string][]byte{
				"TV.1967.S01.German.AC3.COMPLETE.DVD.MPEG2-GROUP/VIDEOS_TS/VIDEO_TS.BUP": []byte("a"),
				"TV.1967.S01.German.AC3.COMPLETE.DVD.MPEG2-GROUP/VIDEOS_TS/VIDEO_TS.IFO": []byte("ab"),
				"TV.1967.S01.German.AC3.COMPLETE.DVD.MPEG2-GROUP/VIDEOS_TS/VTS_01_0.BUP": []byte("abc"),
				"TV.1967.S01.German.AC3.COMPLETE.DVD.MPEG2-GROUP/VIDEOS_TS/VTS_01_0.IFO": []byte("abcd"),
				"TV.1967.S01.German.AC3.COMPLETE.DVD.MPEG2-GROUP/VIDEOS_TS/VTS_01_0.VOB": []byte("abcde"),
			},
			expected: release.Info{
				Name:  "TV.1967.S01.German.AC3.COMPLETE.DVD.MPEG2-GROUP",
				Group: "GROUP",
				Size:  1 + 2 + 3 + 4 + 5, // total file size
				Extensions: map[string]int{
					".vob": 1,
					".bup": 2,
					".ifo": 2,
				},
				Language:      "german",
				TagResolution: release.SD,
				BiggestFile: &dtree.Node{
					Info: &dtree.FileInfo{
						Name: "VTS_01_0.VOB",
						Size: 5,
					},
				},
				ProductTitle: "TV",
				ProductYear:  1967,
				Section:      release.TV,
				Episodes:     []release.Episode{},
			},
		},
		{
			desc: "tv pack retail release",
			root: "TV.1967.S01.German.AC3.COMPLETE.DVD.MPEG2-GROUP",
			testFiles: map[string][]byte{
				"TV.1967.S01.German.AC3.COMPLETE.DVD.MPEG2-GROUP/TV.1967.S01D01.German.AC3.COMPLETE.DVD.MPEG2-GROUP/VIDEOS_TS/VTS_01_0.VOB": []byte("ab"),
				"TV.1967.S01.German.AC3.COMPLETE.DVD.MPEG2-GROUP/TV.1967.S01D02.German.AC3.COMPLETE.DVD.MPEG2-GROUP/VIDEOS_TS/VTS_01_0.VOB": []byte("abc"),
				"TV.1967.S01.German.AC3.COMPLETE.DVD.MPEG2-GROUP/TV.1967.S01D03.German.AC3.COMPLETE.DVD.MPEG2-GROUP/VIDEOS_TS/VTS_01_0.VOB": []byte("abcd"),
			},
			expected: release.Info{
				Name:  "TV.1967.S01.German.AC3.COMPLETE.DVD.MPEG2-GROUP",
				Group: "GROUP",
				Size:  2 + 3 + 4, // total file size
				Extensions: map[string]int{
					".vob": 3,
				},
				Language:      "german",
				TagResolution: release.SD,
				BiggestFile: &dtree.Node{
					Info: &dtree.FileInfo{
						Name: "VTS_01_0.VOB",
						Size: 4,
					},
				},
				ProductTitle: "TV",
				ProductYear:  1967,
				Section:      release.TVPack,
				Episodes: []release.Episode{
					{Number: 1, Name: "VTS_01_0.VOB"},
					{Number: 2, Name: "VTS_01_0.VOB"},
					{Number: 3, Name: "VTS_01_0.VOB"},
				},
			},
		},
		{
			desc: "forbidden files",
			root: "Forbidden.1967.German.1080p.BluRay.x264-Group",
			testFiles: map[string][]byte{
				"Forbidden.1967.German.1080p.BluRay.x264-Group/empty.nfo":              nil,
				"Forbidden.1967.German.1080p.BluRay.x264-Group/forbidden-char-äöü.nfo": []byte("imdb.com/title/tt0123456\n"),
				"Forbidden.1967.German.1080p.BluRay.x264-Group/release-group.mkv":      []byte("should.be.the.biggest.file.here\n"),
				"Forbidden.1967.German.1080p.BluRay.x264-Group/release.nzb":            []byte("abc\n"),
				"Forbidden.1967.German.1080p.BluRay.x264-Group/empty-dir/":             nil,
			},
			expected: release.Info{
				Name:  "Forbidden.1967.German.1080p.BluRay.x264-Group",
				Group: "Group",
				Size:  0 + 25 + 32 + 4, // total file size
				Extensions: map[string]int{
					".nfo": 2,
					".mkv": 1,
					".nzb": 1,
				},
				Language:      "german",
				TagResolution: release.FHD,
				BiggestFile: &dtree.Node{
					Info: &dtree.FileInfo{
						Name: "release-group.mkv",
						Size: 32,
					},
				},
				ProductTitle: "Forbidden",
				ProductYear:  1967,
				Section:      release.Movies,
				NFO: &release.NFOFile{
					Name:    "forbidden-char-äöü.nfo",
					Content: []byte("imdb.com/title/tt0123456\n"),
				},
				ImdbID: 123456,
				ForbiddenFiles: []release.ForbiddenFile{
					{
						Info: &dtree.FileInfo{
							Name: "empty.nfo",
							Size: 0,
						},
						Error: release.ErrEmptyFile,
					},
					{
						Info: &dtree.FileInfo{
							Name: "forbidden-char-äöü.nfo",
							Size: 4,
						},
						Error: release.ErrForbiddenCharacters,
					},
					{
						Info: &dtree.FileInfo{
							Name: "release.nzb",
							Size: 4,
						},
						Error: release.ErrForbiddenExtension,
					},
					{
						Info: &dtree.FileInfo{
							Name: "empty-dir",
							Size: 4096,
						},
						Error: release.ErrEmptyFolder,
					},
				},
			},
			expectedErr: release.ErrForbiddenFiles,
		},
		{
			desc: "ignore files and folders",
			root: "Ignore.1967.German.1080p.BluRay.x264-Group",
			testFiles: map[string][]byte{
				"Ignore.1967.German.1080p.BluRay.x264-Group/release-group.rar": []byte("should.be.the.biggest.file.here\n"),
				"Ignore.1967.German.1080p.BluRay.x264-Group/release-group.sfv": []byte("packed-group.rar  605ec0d\n"),
				"Ignore.1967.German.1080p.BluRay.x264-Group/release-group.nzb": []byte("asd\n"),
				"Ignore.1967.German.1080p.BluRay.x264-Group/Sample/":           nil,
			},
			ignore: []string{"Sample", "*.sfv", "*.nzb"},
			expected: release.Info{
				Name:  "Ignore.1967.German.1080p.BluRay.x264-Group",
				Group: "Group",
				Size:  32, // total file size
				Extensions: map[string]int{
					".rar": 1,
				},
				ArchiveCount:  1,
				Language:      "german",
				TagResolution: release.FHD,
				BiggestFile: &dtree.Node{
					Info: &dtree.FileInfo{
						Name: "release-group.rar",
						Size: 32,
					},
				},
				ProductTitle: "Ignore",
				ProductYear:  1967,
				Section:      release.Movies,
			},
		},
		{
			desc: "packed release",
			root: "Packed.1967.German.1080p.BluRay.x264-Group",
			testFiles: map[string][]byte{
				"Packed.1967.German.1080p.BluRay.x264-Group/packed-group.rar": []byte("should.be.the.biggest.file.here\n"),
				"Packed.1967.German.1080p.BluRay.x264-Group/packed-group.nfo": []byte("imdb.com/title/tt0123456\n"),
				"Packed.1967.German.1080p.BluRay.x264-Group/packed-group.sfv": []byte("packed-group.rar  605ec0d\n"),
			},
			expected: release.Info{
				Name:  "Packed.1967.German.1080p.BluRay.x264-Group",
				Group: "Group",
				Size:  32 + 25 + 26, // total file size
				Extensions: map[string]int{
					".rar": 1,
					".nfo": 1,
					".sfv": 1,
				},
				ArchiveCount:  1,
				SfvCount:      1,
				Language:      "german",
				TagResolution: release.FHD,
				BiggestFile: &dtree.Node{
					Info: &dtree.FileInfo{
						Name: "packed-group.rar",
						Size: 32,
					},
				},
				ProductTitle: "Packed",
				ProductYear:  1967,
				Section:      release.Movies,
				ImdbID:       123456,
				NFO: &release.NFOFile{
					Name:    "packed-group.nfo",
					Content: []byte("imdb.com/title/tt0123456\n"),
				},
			},
		},
		{
			desc: "single file release",
			root: "Single.1967.German.2160p.BluRay.x264-Group.mkv",
			testFiles: map[string][]byte{
				"Single.1967.German.2160p.BluRay.x264-Group.mkv": []byte("should.be.the.biggest.file.here\n"),
			},
			expected: release.Info{
				Name:  "Single.1967.German.2160p.BluRay.x264-Group",
				Group: "Group",
				Size:  32,
				Extensions: map[string]int{
					".mkv": 1,
				},
				Language:      "german",
				TagResolution: release.UHD,
				BiggestFile: &dtree.Node{
					Info: &dtree.FileInfo{
						Name: "Single.1967.German.2160p.BluRay.x264-Group.mkv",
						Size: 32,
					},
				},
				ProductTitle: "Single",
				ProductYear:  1967,
				Section:      release.Movies,
			},
		},
		{
			desc: "empty main folder",
			root: "Empty.1976.German.1080p.BluRay.x264-Group",
			testFiles: map[string][]byte{
				"Empty.1976.German.1080p.BluRay.x264-Group/": nil,
			},
			expected:    release.Info{},
			expectedErr: release.ErrEmptyFolder,
		},
		{
			desc:        "non-existing folder",
			root:        "Non-Existent.1976.German.1080p.BluRay.x264-Group",
			testFiles:   map[string][]byte{},
			expected:    release.Info{},
			expectedErr: os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tempDir := t.TempDir()
			setupTestDir(t, tempDir, tt.testFiles)
			releaseDir := filepath.Join(tempDir, tt.root)

			// disable logger
			zerolog.SetGlobalLevel(zerolog.FatalLevel)

			releaseService := release.NewServiceBuilder().WithSkipPre(true).WithSkipMediaInfo(true).Build()
			gotRelease, gotErr := releaseService.Parse(releaseDir, tt.ignore...)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, gotErr, tt.expectedErr)
				if len(tt.expected.ForbiddenFiles) > 0 {
					compareForbiddenFiles(t, tt.expected.ForbiddenFiles, gotRelease.ForbiddenFiles)
				}
				if errors.Is(gotErr, release.ErrForbiddenFiles) {
					compareRelease(t, tt.expected, *gotRelease)
				}
				return
			}
			assert.NoError(t, gotErr)
			compareRelease(t, tt.expected, *gotRelease)
		})
	}
}

func TestInfo_HasMetaFiles(t *testing.T) {
	tests := []struct {
		desc         string
		inputRelease release.Info
		inputIgnore  []string
		expected     bool
	}{
		{
			desc: "got jpg",
			inputRelease: release.Info{
				Extensions: map[string]int{".jpg": 1},
			},
			expected: true,
		},
		{
			desc: "got png",
			inputRelease: release.Info{
				Extensions: map[string]int{".png": 1},
			},
			expected: true,
		},
		{
			desc: "ignore jpg",
			inputRelease: release.Info{
				Extensions: map[string]int{".jpg": 1},
			},
			inputIgnore: []string{".jpg"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := tt.inputRelease.HasMetaFiles(tt.inputIgnore...)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestInfo_HasAnyExtension(t *testing.T) {
	tests := []struct {
		desc            string
		inputRelease    release.Info
		inputExtensions []string
		expected        bool
	}{
		{
			desc: "valid input",
			inputRelease: release.Info{
				Extensions: map[string]int{".nfo": 1},
			},
			inputExtensions: []string{".nfo"},
			expected:        true,
		},
		{
			desc: "valid input, 2 extensions",
			inputRelease: release.Info{
				Extensions: map[string]int{".nfo": 1},
			},
			inputExtensions: []string{".mkv", ".nfo"},
			expected:        true,
		},
		{
			desc: "no match",
			inputRelease: release.Info{
				Extensions: map[string]int{".nfo": 1},
			},
			inputExtensions: []string{".rar"},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := tt.inputRelease.HasAnyExtension(tt.inputExtensions...)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestInfo_HasExtensions(t *testing.T) {
	tests := []struct {
		desc            string
		inputRelease    release.Info
		inputExtensions []string
		expected        bool
	}{
		{
			desc: "valid input",
			inputRelease: release.Info{
				Extensions: map[string]int{".nfo": 1},
			},
			inputExtensions: []string{".nfo"},
			expected:        true,
		},
		{
			desc: "valid input, 2 extension",
			inputRelease: release.Info{
				Extensions: map[string]int{".nfo": 1, ".mkv": 1},
			},
			inputExtensions: []string{".nfo", ".mkv"},
			expected:        true,
		},
		{
			desc: "2 extensions, but only one match",
			inputRelease: release.Info{
				Extensions: map[string]int{".nfo": 1},
			},
			inputExtensions: []string{".nfo", ".mkv"},
			expected:        false,
		},
		{
			desc: "no match",
			inputRelease: release.Info{
				Extensions: map[string]int{".nfo": 1},
			},
			inputExtensions: []string{".mkv"},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := tt.inputRelease.HasExtensions(tt.inputExtensions...)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestInfo_HasAnyLanguages(t *testing.T) {
	tests := []struct {
		desc           string
		inputRelease   release.Info
		inputLanguages []string
		expected       bool
	}{
		{
			desc: "find german by section",
			inputRelease: release.Info{
				Language: "de",
			},
			inputLanguages: []string{"german", "de"},
			expected:       true,
		},
		{
			desc: "find german by mediainfo",
			inputRelease: release.Info{
				MediaInfo: &release.MediaInfo{
					Media: release.Media{
						Tracks: []release.MediaInfoTrack{
							{Type: string(release.Audio), Language: "german"},
						},
					},
				},
			},
			inputLanguages: []string{"german", "de"},
			expected:       true,
		},
		{
			desc:           "no language found",
			inputRelease:   release.Info{},
			inputLanguages: []string{"german", "de"},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := tt.inputRelease.HasAnyLanguage(tt.inputLanguages...)
			assert.Equal(t, tt.expected, got)
		})
	}
}
