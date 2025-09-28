package release

import (
	"strconv"
	"testing"

	"github.com/f4n4t/go-dtree"
	"github.com/f4n4t/go-release/pkg/progress"
	"github.com/f4n4t/go-release/pkg/srrdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelease_VerifySingleSRR(t *testing.T) {
	validSRR := srrdb.Release{
		Name: "Test.Release",
		ArchivedFiles: srrdb.ArchivedFiles{
			{
				Name: "test.mkv",
				Size: 13,
				CRC:  "d61538ea",
			},
		},
	}

	tests := []struct {
		name      string
		testFiles map[string][]byte
		inputSRR  srrdb.Release
		fastCheck bool
		wantErr   error
	}{
		{
			name: "valid input",
			testFiles: map[string][]byte{
				"test.mkv": []byte("test-content\n"),
			},
			inputSRR:  validSRR,
			fastCheck: false,
		},
		{
			name: "valid input (fast check)",
			testFiles: map[string][]byte{
				"test.mkv": []byte("test-content\n"),
			},
			inputSRR:  validSRR,
			fastCheck: true,
		},
		{
			name: "invalid size",
			testFiles: map[string][]byte{
				"test.mkv": []byte("test-content\n"),
			},
			inputSRR: srrdb.Release{
				Name: "Test.Release",
				ArchivedFiles: srrdb.ArchivedFiles{
					{
						Name: "test.mkv",
						Size: 4,
						CRC:  "d61538ea",
					},
				},
			},
			fastCheck: false,
			wantErr:   ErrSrrValidationFailed,
		},
		{
			name: "invalid checksum (fast check enabled)",
			testFiles: map[string][]byte{
				"test.mkv": []byte("test-content\n"),
			},
			inputSRR: srrdb.Release{
				Name: "Test.Release",
				ArchivedFiles: srrdb.ArchivedFiles{
					{
						Name: "test.mkv",
						Size: 13,
						CRC:  "ffffffff",
					},
				},
			},
			fastCheck: true,
		},
		{
			name: "invalid checksum",
			testFiles: map[string][]byte{
				"test.mkv": []byte("test-content\n"),
			},
			inputSRR: srrdb.Release{
				Name: "Test.Release",
				ArchivedFiles: srrdb.ArchivedFiles{
					{
						Name: "test.mkv",
						Size: 13,
						CRC:  "ffffffff",
					},
				},
			},
			fastCheck: false,
			wantErr:   ErrSrrValidationFailed,
		},
		{
			name: "invalid checksum syntax",
			testFiles: map[string][]byte{
				"test.mkv": []byte("test-content\n"),
			},
			inputSRR: srrdb.Release{
				Name: "Test.Release",
				ArchivedFiles: srrdb.ArchivedFiles{
					{
						Name: "test.mkv",
						Size: 13,
						CRC:  "zzzzzzzz",
					},
				},
			},
			fastCheck: false,
			wantErr:   strconv.ErrSyntax,
		},
		{
			name: "missing file",
			testFiles: map[string][]byte{
				"another-file.mkv": []byte("blub\n"),
			},
			inputSRR:  validSRR,
			fastCheck: false,
			wantErr:   dtree.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			setupTestDir(t, tempDir, tt.testFiles)

			releaseService := NewServiceBuilder().WithSkipPre(true).WithSkipMediaInfo(true).Build()

			rel, err := releaseService.Parse(tempDir)
			require.NoError(t, err)

			gotErr := releaseService.verifySingleSRR(rel, tt.inputSRR, &progress.NoOpProgressBar{}, false, tt.fastCheck)
			assert.ErrorIs(t, gotErr, tt.wantErr)
		})
	}
}
