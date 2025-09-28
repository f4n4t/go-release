package release

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/f4n4t/go-release/pkg/progress"
	"github.com/f4n4t/go-release/pkg/srrdb"
	"github.com/f4n4t/go-release/pkg/utils"
)

var (
	// ErrSrrValidationFailed indicates that the SRR validation process failed due to size or CRC mismatch.
	ErrSrrValidationFailed = errors.New("srr check failed")

	// ErrNoSRRFile indicates that no corresponding SRR file was found on the SRR database.
	ErrNoSRRFile = errors.New("nothing found on srrdb")
)

// CheckSRR validates the SRR integrity of a release using provided information and options.
func (s *Service) CheckSRR(rel *Info, showProgress bool, fastCheck bool) error {
	startTime := time.Now()

	useParallelRead, err := s.useParallelRead(rel.Root.FullPath)
	if err != nil {
		return err
	}

	var releases []string

	for _, f := range rel.MediaFiles {
		releases = append(releases, f.Parent.Info.Name)
	}

	srrdbReleases, err := s.fetchSRRInformation(releases)
	if err != nil {
		return err
	}

	var totalSize int64
	for _, release := range srrdbReleases {
		totalSize += srrdb.TotalSize(release.ArchivedFiles)
	}

	bar := progress.NewProgressBar(showProgress, totalSize, true)

	s.log.Info().Str("totalSize", utils.Bytes(totalSize)).Msg("starting srr check")

	for _, srr := range srrdbReleases {
		if err := s.verifySingleSRR(rel, srr, bar, useParallelRead, fastCheck); err != nil {
			bar.Cancel()
			return fmt.Errorf("verify srr %s: %w", srr.Name, err)
		}
	}

	_ = bar.Finish()

	s.log.Info().Str("dur", time.Since(startTime).String()).Msg("checked srr")

	return nil
}

// fetchSRRInformation retrieves SRR information for a list of release names from the SRR database.
// It logs errors for individual releases that fail to retrieve and skips them.
// Returns an error if no SRR records are successfully retrieved.
func (s *Service) fetchSRRInformation(releaseNames []string) ([]srrdb.Release, error) {
	srrdbReleases := make([]srrdb.Release, 0, len(releaseNames))

	for _, releaseName := range releaseNames {
		srr, err := srrdb.GetInformation(releaseName)
		if err != nil {
			s.log.Error().Err(err).Str("release", releaseName).Msg("no srr record retrieved")
			continue
		}
		srrdbReleases = append(srrdbReleases, srr)
	}

	if len(srrdbReleases) == 0 {
		return nil, ErrNoSRRFile
	}

	return srrdbReleases, nil
}

// verifySingleSRR validates the integrity of a single SRR file by comparing its metadata with local files.
func (s *Service) verifySingleSRR(rel *Info, srr srrdb.Release, bar progress.Progress, useParallelRead bool, fastCheck bool) error {
	for _, fs := range srr.ArchivedFiles {
		localFile, err := rel.Root.GetFile(fs.Name)
		if err != nil {
			return fmt.Errorf("get file: %w", err)
		}

		if localFile.Info.Size != fs.Size {
			return fmt.Errorf("%w: size mismatch", ErrSrrValidationFailed)
		}

		if fastCheck {
			bar.Set64(localFile.Info.Size)
			continue
		}

		srrCRC, err := strconv.ParseUint(fs.CRC, 16, 32)
		if err != nil {
			return fmt.Errorf("parse crc: %w", err)
		}

		crcChecker := utils.NewCheckCRCBuilder(localFile.FullPath, uint32(srrCRC)).
			WithParallelRead(useParallelRead).
			WithProgressBar(bar).
			WithHashThreads(s.hashThreads).Build()

		if err := crcChecker.VerifyCRC32(); err != nil {
			return fmt.Errorf("%w: crc mismatch", ErrSrrValidationFailed)
		}
	}

	s.log.Debug().Str("srr", srr.Name).Msg("check passed")

	return nil
}
