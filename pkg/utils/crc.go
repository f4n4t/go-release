package utils

import (
	"context"
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"os"
	"runtime"
	"sync"

	"github.com/f4n4t/go-release/pkg/progress"
	"github.com/vimeo/go-util/crc32combine"
)

// hashPool is a pool for crc32 hashers.
var hashPool = sync.Pool{
	New: func() any {
		return crc32.NewIEEE()
	},
}

var (
	ErrCRCMismatch = errors.New("crc mismatch")
)

type chunk struct {
	idx int
	crc uint32
	len int64
}

type fileChunk struct {
	startPos    int64
	chunkLength int64
}

type CheckCRC struct {
	file            string
	wantCRC         uint32
	bar             progress.Progress
	useParallelRead bool
	hashThreads     int
	ctx             context.Context
}

type CheckCRCBuilder struct {
	checkCRC CheckCRC
}

func NewCheckCRCBuilder(inputFile string, wantCRC uint32) *CheckCRCBuilder {
	cb := &CheckCRCBuilder{}
	cb.checkCRC.file = inputFile
	cb.checkCRC.wantCRC = wantCRC
	return cb
}

func (cb *CheckCRCBuilder) WithProgressBar(bar progress.Progress) *CheckCRCBuilder {
	cb.checkCRC.bar = bar
	return cb
}

func (cb *CheckCRCBuilder) WithParallelRead(parallelRead bool) *CheckCRCBuilder {
	cb.checkCRC.useParallelRead = parallelRead
	return cb
}

func (cb *CheckCRCBuilder) WithHashThreads(i int) *CheckCRCBuilder {
	cb.checkCRC.hashThreads = max(0, i)
	return cb
}

func (cb *CheckCRCBuilder) WithContext(ctx context.Context) *CheckCRCBuilder {
	cb.checkCRC.ctx = ctx
	return cb
}

func (cb *CheckCRCBuilder) Build() CheckCRC {
	if cb.checkCRC.ctx == nil {
		cb.checkCRC.ctx = context.Background()
	}
	return CheckCRC{
		file:            cb.checkCRC.file,
		wantCRC:         cb.checkCRC.wantCRC,
		bar:             cb.checkCRC.bar,
		useParallelRead: cb.checkCRC.useParallelRead,
		hashThreads:     cb.checkCRC.hashThreads,
		ctx:             cb.checkCRC.ctx,
	}
}

func (c CheckCRC) VerifyCRC32() error {
	var (
		fileCRC uint32
		err     error
	)

	if c.useParallelRead {
		fileCRC, err = GetCRC32Parallel(c.ctx, c.file, c.hashThreads, c.bar)
	} else {
		fileCRC, err = GetCRC32(c.ctx, c.file, c.bar)
	}

	if err != nil {
		return fmt.Errorf("%s: calculate crc32: %w", c.file, err)
	}

	if fileCRC != c.wantCRC {
		return fmt.Errorf("%s: %w", c.file, ErrCRCMismatch)
	}

	return nil
}

// GetCRC32Parallel returns the crc32 checksum of a file using multiple goroutines.
func GetCRC32Parallel(ctx context.Context, filePath string, hashThreads int, writers ...io.Writer) (uint32, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("file info: %w", err)
	} else if fileInfo.IsDir() {
		return 0, fmt.Errorf("file %s: directory not regular file", filePath)
	}

	var numWorkers int
	if hashThreads > 0 {
		numWorkers = hashThreads
	} else {
		numWorkers = runtime.GOMAXPROCS(0)
	}

	var (
		fileSize       = fileInfo.Size()
		chunkSize      = int64(1024 * 1024 * 10) // 10MB
		totalChunks    = (fileSize + int64(chunkSize) - 1) / int64(chunkSize)
		resultChan     = make(chan chunk, numWorkers)
		jobChan        = make(chan func() chunk, numWorkers)
		errChan        = make(chan error)
		chunkList      = make([]fileChunk, totalChunks)
		chunkIdx       = 0
		chunkRemaining = chunkSize
		fileRemaining  = fileSize
	)

	for fileRemaining > 0 {
		if chunkRemaining == 0 {
			chunkRemaining = chunkSize
		}

		toAllocate := min(fileRemaining, chunkRemaining)

		chunkList[chunkIdx] = fileChunk{
			startPos:    fileSize - fileRemaining,
			chunkLength: toAllocate,
		}

		chunkIdx++
		chunkRemaining -= toAllocate
		fileRemaining -= toAllocate
	}

	for range numWorkers {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return

				case job, ok := <-jobChan:
					if !ok {
						return
					}
					resultChan <- job()
				}
			}
		}()
	}

	go func() {
		defer close(jobChan)

		for chunkIdx, c := range chunkList {
			jobChan <- func() chunk {
				hasher := hashPool.Get().(hash.Hash32)
				defer hashPool.Put(hasher)
				hasher.Reset()

				f, err := os.Open(filePath)
				if err != nil {
					errChan <- err
					return chunk{}
				}
				defer f.Close()

				if _, err := f.Seek(c.startPos, io.SeekStart); err != nil {
					errChan <- err
					return chunk{}
				}

				writer := io.MultiWriter(append([]io.Writer{hasher}, writers...)...)

				written, err := io.Copy(writer, io.LimitReader(f, c.chunkLength))
				switch {
				case err != nil:
					errChan <- fmt.Errorf("%s: copy: %w", filePath, err)
					return chunk{}

				case written != c.chunkLength:
					// should never happen
					errChan <- fmt.Errorf("incomplete read: expected %d bytes, got %d", c.chunkLength, written)
					return chunk{}
				}

				crc := hasher.Sum32()

				return chunk{idx: chunkIdx, crc: crc, len: written}
			}
		}
	}()

	var (
		checkedLength int64
		resultCRC     uint32
		results       = make([]chunk, totalChunks)
	)

	for checkedLength < fileSize {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()

		case result := <-resultChan:
			checkedLength += result.len
			results[result.idx] = result

		case err := <-errChan:
			return 0, fmt.Errorf("crc32 calculation: %w", err)
		}
	}

	for _, crc := range results {
		resultCRC = crc32combine.CRC32Combine(crc32.IEEE, resultCRC, crc.crc, crc.len)
	}

	return resultCRC, nil
}

// GetCRC32 returns the crc32 checksum of a file.
func GetCRC32(ctx context.Context, filePath string, writers ...io.Writer) (uint32, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("file info: %w", err)
	} else if fileInfo.IsDir() {
		return 0, fmt.Errorf("file %s: directory not regular file", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	hasher := hashPool.Get().(hash.Hash32)
	defer hashPool.Put(hasher)
	hasher.Reset()

	writer := io.MultiWriter(append([]io.Writer{hasher}, writers...)...)

	if _, err := io.Copy(writer, NewReader(ctx, file)); err != nil {
		switch err {
		case context.Canceled, context.DeadlineExceeded:
			// canceled by user or timed out
			return 0, err

		default:
			return 0, fmt.Errorf("%s: copy: %w", filePath, err)
		}
	}

	return hasher.Sum32(), nil
}

type readerCtx struct {
	ctx context.Context
	r   io.Reader
}

func (r *readerCtx) Read(p []byte) (n int, err error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

// NewReader gets a context-aware io.Reader.
func NewReader(ctx context.Context, r io.Reader) io.Reader {
	return &readerCtx{
		ctx: ctx,
		r:   r,
	}
}
