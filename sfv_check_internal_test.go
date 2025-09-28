package release

import (
	"fmt"
	"github.com/rs/zerolog"
	"hash/crc32"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessSFVEntry(t *testing.T) {
	tests := []struct {
		desc          string
		setupTestFile func(t *testing.T, filePath string)
		baseDir       string
		fileName      string
		crcStr        string
		wantFile      sfvFile
		wantErr       error
	}{
		{
			desc: "valid input",
			setupTestFile: func(t *testing.T, filePath string) {
				f, err := os.Create(filePath)
				require.NoError(t, err)
				f.WriteString("test-content\n")
			},
			fileName: "test.rar",
			crcStr:   "d61538ea",
			wantFile: sfvFile{
				name: "test.rar",
				path: "",
				crc:  3591715050,
				size: 13,
			},
		},
		{
			desc:          "missing file",
			setupTestFile: func(t *testing.T, filePath string) {},
			fileName:      "test.rar",
			crcStr:        "d61538ea",
			wantErr:       os.ErrNotExist,
		},
		{
			desc: "invalid crc",
			setupTestFile: func(t *testing.T, filePath string) {
				f, err := os.Create(filePath)
				require.NoError(t, err)
				f.WriteString("test-content")
			},
			fileName: "test.rar",
			crcStr:   "zzz",
			wantErr:  strconv.ErrSyntax,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, tt.fileName)

			tt.setupTestFile(t, filePath)

			gotFile, err := processSFVEntry(tempDir, tt.fileName, tt.crcStr)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			tt.wantFile.path = filePath

			assert.Equal(t, tt.wantFile, gotFile)
		})
	}
}

func TestGetFilesFromSFV(t *testing.T) {
	tests := []struct {
		desc           string
		sfvName        string
		testFiles      map[string][]byte
		setupTestFiles func(t *testing.T, baseDir, sfvName string, testFiles map[string][]byte)
		wantResult     sfvFiles
		wantErr        error
	}{
		{
			desc:    "valid input",
			sfvName: "test.sfv",
			testFiles: map[string][]byte{
				"test.rar": []byte("test-content\n"),
			},
			setupTestFiles: func(t *testing.T, baseDir, sfvName string, testFiles map[string][]byte) {
				sfv, err := os.Create(filepath.Join(baseDir, sfvName))
				require.NoError(t, err)
				for name, content := range testFiles {
					os.WriteFile(filepath.Join(baseDir, name), content, 0666)
					sfv.WriteString(fmt.Sprintf("%s %x\n", name, crc32.ChecksumIEEE(content)))
				}
			},
			wantResult: []sfvFile{
				{
					name: "test.rar",
					path: "",
					crc:  3591715050,
					size: 13,
				},
			},
		},
		{
			desc:           "missing sfv file",
			sfvName:        "test.sfv",
			setupTestFiles: func(t *testing.T, baseDir, sfvName string, testFiles map[string][]byte) {},
			wantErr:        os.ErrNotExist,
		},
		{
			desc:    "invalid sfv",
			sfvName: "test.sfv",
			setupTestFiles: func(t *testing.T, baseDir, sfvName string, testFiles map[string][]byte) {
				_, err := os.Create(filepath.Join(baseDir, sfvName))
				require.NoError(t, err)
			},
			wantErr: ErrInvalidSfv,
		},
		{
			desc:    "missing archive",
			sfvName: "test.sfv",
			testFiles: map[string][]byte{
				"test.rar": []byte("test-content\n"),
			},
			setupTestFiles: func(t *testing.T, baseDir, sfvName string, testFiles map[string][]byte) {
				sfv, err := os.Create(filepath.Join(baseDir, sfvName))
				require.NoError(t, err)
				for name, content := range testFiles {
					sfv.WriteString(fmt.Sprintf("%s %x\n", name, crc32.ChecksumIEEE(content)))
				}
			},
			wantErr: os.ErrNotExist,
		},
	}

	zerolog.SetGlobalLevel(zerolog.FatalLevel)

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tempDir := t.TempDir()
			tt.setupTestFiles(t, tempDir, tt.sfvName, tt.testFiles)

			sfvPath := filepath.Join(tempDir, tt.sfvName)

			gotFiles, err := getFilesFromSFV(sfvPath)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			for i := range gotFiles {
				// ignore the path here
				gotFiles[i].path = ""
			}

			assert.Equal(t, tt.wantResult, gotFiles)
		})
	}
}

func TestRelease_PerformSFVCheck(t *testing.T) {
	const testSFVName = "test.sfv"

	tests := []struct {
		desc       string
		files      map[string][]byte
		wantResult bool
		wantErr    error
	}{
		{
			desc: "valid input",
			files: map[string][]byte{
				"test.rar":  []byte("test-content\n"),
				testSFVName: []byte("test.rar d61538ea\n"),
			},
			wantResult: true,
		},
		{
			desc: "valid input uppercase checksum",
			files: map[string][]byte{
				"test.rar":  []byte("test-content\n"),
				testSFVName: []byte("test.rar D61538EA\n"),
			},
			wantResult: true,
		},
		{
			desc: "invalid checksum",
			files: map[string][]byte{
				"test.rar":  []byte("test-content\n"),
				testSFVName: []byte("test.rar 11111111\n"),
			},
			wantResult: false,
		},
		{
			desc: "missing file",
			files: map[string][]byte{
				testSFVName: []byte("test.rar 11111111\n"),
			},
			wantErr: os.ErrNotExist,
		},
		{
			desc: "invalid sfv",
			files: map[string][]byte{
				testSFVName: []byte("invalid"),
				"test.rar":  []byte("test-content\n"),
			},
			wantErr: ErrInvalidSfv,
		},
		// test for empty sfv not possible, rel service doesn't allow empty files
	}

	zerolog.SetGlobalLevel(zerolog.FatalLevel)

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tmpDir := t.TempDir()

			setupTestDir(t, tmpDir, tt.files)

			sfvPath := filepath.Join(tmpDir, testSFVName)

			releaseService := NewServiceBuilder().WithSkipPre(true).WithSkipMediaInfo(true).Build()

			rel, err := releaseService.Parse(tmpDir)
			require.NoError(t, err)

			gotResult, err := releaseService.performSFVCheck(rel, sfvPath, false)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}

func setupTestDir(t *testing.T, baseDir string, testFiles map[string][]byte) {
	for name, content := range testFiles {
		dir, file := filepath.Split(name)
		if dir != "" {
			require.NoError(t, os.MkdirAll(filepath.Join(baseDir, dir), 0755))
		}
		require.NoError(t, os.WriteFile(filepath.Join(baseDir, dir, file), content, 0666))
	}
}
