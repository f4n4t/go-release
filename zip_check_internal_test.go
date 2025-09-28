package release

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessDizFile(t *testing.T) {
	tests := []struct {
		name          string
		content       []byte
		expectedCount archiveCount
		expectedError error
	}{
		{
			name:          "valid content",
			content:       []byte("test [01/05]"),
			expectedCount: archiveCount{current: 1, total: 5},
			expectedError: nil,
		},
		{
			name:          "no count",
			content:       []byte("no archive count"),
			expectedCount: archiveCount{},
			expectedError: ErrNoFileCountInDiz,
		},
		{
			name:          "empty file",
			content:       nil,
			expectedCount: archiveCount{},
			expectedError: ErrNoFileCountInDiz,
		},
		{
			name:          "invalid format",
			content:       []byte("test [01/XX]"),
			expectedCount: archiveCount{},
			expectedError: ErrNoFileCountInDiz,
		},
		{
			name:          "invalid format",
			content:       []byte("test [XX/02]"),
			expectedCount: archiveCount{},
			expectedError: ErrNoFileCountInDiz,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := processDizContent(tt.content)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestValidateArchiveCollection(t *testing.T) {
	tests := []struct {
		name          string
		input         archiveResult
		expectedError error
	}{
		{
			name: "valid single file",
			input: archiveResult{
				archives: []archiveInfo{
					{size: 100},
					{size: 100},
				},
				expectedTotal: 2,
			},
			expectedError: nil,
		},
		{
			name: "valid two sizes",
			input: archiveResult{
				archives: []archiveInfo{
					{size: 100},
					{size: 200},
				},
				expectedTotal: 2,
			},
			expectedError: nil,
		},
		{
			name: "invalid expected count",
			input: archiveResult{
				archives: []archiveInfo{
					{size: 100},
				},
				expectedTotal: 2,
			},
			expectedError: ErrZipValidationFailed,
		},
		{
			name: "invalid sizes",
			input: archiveResult{
				archives: []archiveInfo{
					{size: 100},
					{size: 200},
					{size: 300},
				},
				expectedTotal: 3,
			},
			expectedError: ErrZipValidationFailed,
		},
		{
			name: "empty archives",
			input: archiveResult{
				archives:      []archiveInfo{},
				expectedTotal: 0,
			},
			expectedError: nil,
		},
		{
			name: "invalid no archives with count",
			input: archiveResult{
				archives:      []archiveInfo{},
				expectedTotal: 1,
			},
			expectedError: ErrZipValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArchiveCollection(tt.input)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestProcessZipContents(t *testing.T) {
	tests := []struct {
		name        string
		setupZip    func(string, string) error
		extractNFO  bool
		wantArchive archiveInfo
		wantNFO     NFOFile
		wantErr     error
	}{
		{
			name: "valid case with nfo extraction",
			setupZip: func(tmpPath, filename string) error {
				return createTestZip(tmpPath, filename, map[string][]byte{
					"file.diz":    []byte("test [01/03]"),
					"file.nfo":    []byte("nfo-content"),
					"archive.rar": []byte("rar-content"),
				})
			},
			extractNFO: true,
			wantArchive: archiveInfo{
				name:    "archive.rar",
				size:    uint64(len([]byte("rar-content"))),
				current: 1,
				total:   3,
			},
			wantNFO: NFOFile{
				Name:    "file.nfo",
				Content: []byte("nfo-content"),
			},
			wantErr: nil,
		},
		{
			name: "valid case without nfo extraction",
			setupZip: func(tmpPath, filename string) error {
				return createTestZip(tmpPath, filename, map[string][]byte{
					"file.diz":    []byte("test [01/03]"),
					"file.nfo":    []byte("nfo-content"),
					"archive.rar": []byte("rar-content"),
				})
			},
			extractNFO: false,
			wantArchive: archiveInfo{
				name:    "archive.rar",
				size:    uint64(len([]byte("rar-content"))),
				current: 1,
				total:   3,
			},
			wantNFO: NFOFile{},
			wantErr: nil,
		},
		{
			name: "no archive in zip",
			setupZip: func(tmpPath, filename string) error {
				return createTestZip(tmpPath, filename, map[string][]byte{
					"file.diz": []byte("test [01/03]"),
					"file.nfo": []byte("nfo-content"),
				})
			},
			extractNFO:  true,
			wantArchive: archiveInfo{},
			wantNFO:     NFOFile{},
			wantErr:     ErrNoArchiveInZip,
		},
		{
			name: "no file count in diz",
			setupZip: func(tmpPath, filename string) error {
				return createTestZip(tmpPath, filename, map[string][]byte{
					"file.diz":    []byte("no file count"),
					"archive.rar": []byte("rar-content"),
				})
			},
			extractNFO:  true,
			wantArchive: archiveInfo{},
			wantNFO:     NFOFile{},
			wantErr:     ErrNoFileCountInDiz,
		},
		{
			name: "different rar syntax",
			setupZip: func(tmpPath, filename string) error {
				return createTestZip(tmpPath, filename, map[string][]byte{
					"file.diz":  []byte("test [01/03]"),
					"part1.r01": []byte("rar-part"),
				})
			},
			extractNFO: true,
			wantArchive: archiveInfo{
				name:    "part1.r01",
				size:    uint64(len([]byte("rar-part"))),
				current: 1,
				total:   3,
			},
			wantNFO: NFOFile{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tempZipFile := filepath.Join(tempDir, "test.zip")

			err := tt.setupZip(tempDir, "test.zip")
			require.NoError(t, err, "error creating test zip file")

			zipReader, err := zip.OpenReader(tempZipFile)
			require.NoError(t, err, "error opening test zip file")

			gotArchive, gotNFO, err := processZipContents(zipReader, tt.extractNFO)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.wantArchive, gotArchive)
			assert.Equal(t, tt.wantNFO, gotNFO)
		})
	}
}

func TestProcessZipFiles(t *testing.T) {
	tests := []struct {
		name       string
		zipFiles   map[string]map[string][]byte
		wantResult archiveResult
		wantErr    error
	}{
		{
			name: "valid single file",
			zipFiles: map[string]map[string][]byte{
				"test.zip": {
					"file.diz":    []byte("test [01/01]"),
					"file.nfo":    []byte("nfo-content"),
					"archive.rar": []byte("rar-content"),
				},
			},
			wantResult: archiveResult{
				archives: []archiveInfo{{
					name:    "archive.rar",
					size:    uint64(len([]byte("rar-content"))),
					current: 1,
					total:   1,
				}},
				expectedTotal: 1,
				nfoFile: NFOFile{
					Name:    "file.nfo",
					Content: []byte("nfo-content"),
				},
			},
		},
		{
			name: "valid multi file",
			zipFiles: map[string]map[string][]byte{
				"testa.zip": {
					"archive.r00": []byte("rar-content"),
					"file.diz":    []byte("test [01/03]"),
					"file.nfo":    []byte("nfo-content"),
				},
				"testb.zip": {
					"archive.r01": []byte("rar-content"),
					"file.diz":    []byte("test [02/03]"),
					"file.nfo":    []byte("nfo-content"),
				},
				"testc.zip": {
					"archive.rar": []byte("rar-content"),
					"file.diz":    []byte("test [03/03]"),
					"file.nfo":    []byte("nfo-content"),
				},
			},
			wantResult: archiveResult{
				archives: []archiveInfo{
					{
						name:    "archive.r00",
						size:    uint64(len([]byte("rar-content"))),
						current: 1,
						total:   3,
					},
					{
						name:    "archive.r01",
						size:    uint64(len([]byte("rar-content"))),
						current: 2,
						total:   3,
					},
					{
						name:    "archive.rar",
						size:    uint64(len([]byte("rar-content"))),
						current: 3,
						total:   3,
					},
				},
				expectedTotal: 3,
				nfoFile: NFOFile{
					Name:    "file.nfo",
					Content: []byte("nfo-content"),
				},
			},
		},
		{
			name: "missing file count",
			zipFiles: map[string]map[string][]byte{
				"test.zip": {
					"file.diz":    []byte("no file count"),
					"file.nfo":    []byte("nfo-content"),
					"archive.rar": []byte("rar-content"),
				},
			},
			wantErr: ErrNoFileCountInDiz,
		},
		{
			name: "missing archive",
			zipFiles: map[string]map[string][]byte{
				"test.zip": {
					"file.diz": []byte("test [01/01]"),
					"file.nfo": []byte("nfo-content"),
				},
			},
			wantErr: ErrNoArchiveInZip,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			setupZips(t, tempDir, tt.zipFiles)

			var testFiles []string
			for k := range tt.zipFiles {
				testFiles = append(testFiles, filepath.Join(tempDir, k))
			}

			gotResult, err := processZipFiles(testFiles)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			// check archives first, because they change ordering
			assert.ElementsMatch(t, tt.wantResult.archives, gotResult.archives)

			tt.wantResult.archives = nil
			gotResult.archives = nil

			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}

func createTestZip(path, filename string, files map[string][]byte) error {
	zipFile, err := os.Create(filepath.Join(path, filename))
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for name, content := range files {
		fileWriter, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		_, err = fileWriter.Write(content)
		if err != nil {
			return err
		}
	}

	return nil
}

func setupZips(t *testing.T, tmpPath string, zipFiles map[string]map[string][]byte) {
	for k, v := range zipFiles {
		err := createTestZip(tmpPath, k, v)
		require.NoError(t, err, "error creating test zip file")
	}
}
