package release

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/f4n4t/go-release/pkg/progress"
	"github.com/f4n4t/go-release/pkg/utils"
)

var (
	// sfvRegex is the compiled regex to extract the name and crc from the sfv files.
	sfvRegex = regexp.MustCompile(`(?m)^\s*(?P<name>[^;\s]+)\s+(?P<crc>[a-fA-F0-9]{8})`)
)

var (
	// ErrSfvValidationFailed indicates that an SFV validation process has failed.
	ErrSfvValidationFailed = errors.New("sfv check failed")

	// ErrEmptySfv indicates that the SFV file being processed is empty.
	ErrEmptySfv = errors.New("empty sfv file")

	// ErrInvalidSfv indicates that the provided SFV file is invalid or does not conform to expected formatting rules.
	ErrInvalidSfv = errors.New("invalid sfv file")
)

// sfvFile represents a file with metadata including name, path, CRC checksum, and size.
type sfvFile struct {
	name string
	path string
	crc  uint32
	size int64
}

// sfvFiles represents a collection of sfvFile objects, allowing operations on multiple files with associated metadata.
type sfvFiles []sfvFile

// TotalSize calculates and returns the combined size of all files in the sfvFiles collection.
func (sf sfvFiles) TotalSize() int64 {
	var totalSize int64
	for _, f := range sf {
		totalSize += f.size
	}
	return totalSize
}

// CheckSFV verifies the integrity of files against SFV checksums and logs the results.
// It processes all ".sfv" files associated with the provided Info object.
func (s *Service) CheckSFV(rel *Info, showProgress bool) error {
	startTime := time.Now()

	success := true

	for _, sfv := range rel.Root.GetFiles(".sfv") {
		s.log.Info().Str("sfvFile", sfv.Info.Name).Msg("starting sfv check")

		passed, err := s.performSFVCheck(rel, sfv.FullPath, showProgress)
		if err != nil {
			return fmt.Errorf("perform sfv check %s: %w", sfv.Info.Name, err)
		}

		if !passed {
			s.log.Error().Str("sfvFile", sfv.Info.Name).Msg("check failed")
			success = false
			continue
		}

		s.log.Info().Str("sfvFile", sfv.Info.Name).Msg("check passed")
	}

	if !success {
		return ErrSfvValidationFailed
	}

	s.log.Info().Str("dur", time.Since(startTime).String()).Msg("sfv checks complete")

	return nil
}

// performSFVCheck checks the integrity of files listed in an SFV file by comparing their CRC values with local files.
func (s *Service) performSFVCheck(rel *Info, sfvPath string, showProgress bool) (bool, error) {
	useParallelRead, err := s.useParallelRead(rel.Root.FullPath)
	if err != nil {
		return false, err
	}

	filesFromSFV, err := getFilesFromSFV(sfvPath)
	if err != nil {
		return false, fmt.Errorf("get files from sfv: %w", err)
	}

	if len(filesFromSFV) == 0 {
		return false, ErrEmptySfv
	}

	var (
		passed    = true
		totalSize = filesFromSFV.TotalSize()
		bar       = progress.NewProgressBar(showProgress, totalSize, true)
	)

	for _, sfvFile := range filesFromSFV {
		localFile, err := rel.Root.GetFileByAbsolutePath(sfvFile.path)
		if err != nil {
			return false, fmt.Errorf("get file: %w", err)
		}

		crcChecker := utils.NewCheckCRCBuilder(localFile.FullPath, sfvFile.crc).
			WithParallelRead(useParallelRead).
			WithProgressBar(bar).
			WithContext(s.ctx).
			WithHashThreads(s.hashThreads).Build()

		if err := crcChecker.VerifyCRC32(); err != nil {
			passed = false

			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return false, err
			}

			s.log.Error().Err(err).Msg("verification failed")
			// continue to check every file
		}
	}

	return passed, nil
}

// getFilesFromSFV parses an SFV file, extracts file information and CRC values, and returns the corresponding sfvFiles.
func getFilesFromSFV(sfvPath string) (sfvFiles, error) {
	content, err := os.ReadFile(sfvPath)
	if err != nil {
		return nil, fmt.Errorf("read sfv file: %w", err)
	}

	matches := sfvRegex.FindAllStringSubmatch(string(content), -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: no matches found in sfv file", ErrInvalidSfv)
	}

	files := make(sfvFiles, 0, len(matches))
	sfvDir := filepath.Dir(sfvPath)

	for _, match := range matches {
		file, err := processSFVEntry(sfvDir, match[1], match[2])
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

// processSFVEntry parses an SFV entry, validates file existence, and creates an sfvFile object with metadata.
func processSFVEntry(baseDir, fileName, crcStr string) (sfvFile, error) {
	filePath := filepath.Join(baseDir, fileName)

	fInfo, err := os.Stat(filePath)
	if err != nil {
		return sfvFile{}, fmt.Errorf("stat file %s: %w", fileName, err)
	}

	crcValue, err := strconv.ParseUint(crcStr, 16, 32)
	if err != nil {
		return sfvFile{}, fmt.Errorf("parse crc: %w", err)
	}

	return sfvFile{
		name: fileName,
		path: filePath,
		crc:  uint32(crcValue),
		size: fInfo.Size(),
	}, nil
}
