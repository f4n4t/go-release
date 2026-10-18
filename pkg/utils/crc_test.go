package utils_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/f4n4t/go-release/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCRC32Parallel_NoGoroutineLeakOnCancel(t *testing.T) {
	// sparse file with more chunks (10MB each) than workers, so the job producer would block
	testFile := filepath.Join(t.TempDir(), "test.bin")
	f, err := os.Create(testFile)
	require.NoError(t, err)
	require.NoError(t, f.Truncate(50*1024*1024))
	require.NoError(t, f.Close())

	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = utils.GetCRC32Parallel(ctx, testFile, 1)
	assert.ErrorIs(t, err, context.Canceled)

	// poll manually, assert.Eventually spawns its own goroutine
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	assert.LessOrEqual(t, runtime.NumGoroutine(), before, "goroutines leaked")
}

func TestGetCRC32Parallel_MatchesSequential(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "test.bin")
	content := make([]byte, 25*1024*1024)
	for i := range content {
		content[i] = byte(i % 251)
	}
	require.NoError(t, os.WriteFile(testFile, content, 0o644))

	want, err := utils.GetCRC32(context.Background(), testFile)
	require.NoError(t, err)

	got, err := utils.GetCRC32Parallel(context.Background(), testFile, 2)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
