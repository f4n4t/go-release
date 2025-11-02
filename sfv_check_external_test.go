package release_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/f4n4t/go-release"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDir(t *testing.T, baseDir string, testFiles map[string][]byte) {
	for name, content := range testFiles {
		dir, file := filepath.Split(name)
		if dir != "" {
			require.NoError(t, os.MkdirAll(filepath.Join(baseDir, dir), 0755))
		}
		if file != "" {
			require.NoError(t, os.WriteFile(filepath.Join(baseDir, dir, file), content, 0666))
		}
	}
}

func TestRelease_CheckSFV(t *testing.T) {
	type test struct {
		name      string
		testFiles map[string][]byte
		wantErr   error
	}

	validTest := test{
		name: "valid input",
		testFiles: map[string][]byte{
			"test.r00": []byte("test-content-1\n"),
			"test.r01": []byte("test-content-2\n"),
			"test.rar": []byte("test-content-3\n"),
			"test.sfv": []byte("test.rar e4f6bb59\ntest.r00 d6c0d9db\ntest.r01 fded8a18\n"),
		},
	}

	tests := []test{
		validTest,
		{
			name: "invalid checksum",
			testFiles: map[string][]byte{
				"test.rar": []byte("test-content\n"),
				"test.sfv": []byte("test.rar ffffffff\n"),
			},
			wantErr: release.ErrSfvValidationFailed,
		},
		{
			name: "missing file",
			testFiles: map[string][]byte{
				"test.sfv": []byte("test.rar e4f6bb59\n"),
			},
			wantErr: os.ErrNotExist,
		},
		{
			name: "invalid sfv",
			testFiles: map[string][]byte{
				"test.rar": []byte("test-content\n"),
				"test.sfv": []byte("invalid"),
			},
			wantErr: release.ErrInvalidSfv,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			setupTestDir(t, tempDir, tt.testFiles)

			releaseService := release.NewServiceBuilder().WithSkipPre(true).WithSkipMediaInfo(true).Build()

			rel, err := releaseService.Parse(tempDir)
			require.NoError(t, err)

			gotErr := releaseService.CheckSFV(rel, false)
			assert.ErrorIs(t, gotErr, tt.wantErr)
		})
	}

	t.Run("CheckCancellation", func(t *testing.T) {
		tempDir := t.TempDir()
		setupTestDir(t, tempDir, validTest.testFiles)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// check if cancellation works
		releaseService := release.NewServiceBuilder().WithSkipPre(true).WithSkipMediaInfo(true).WithContext(ctx).Build()

		rel, err := releaseService.Parse(tempDir)
		require.NoError(t, err)

		cancel()

		gotErr := releaseService.CheckSFV(rel, false)
		assert.ErrorIs(t, gotErr, context.Canceled)
	})
}
