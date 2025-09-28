package release

import (
	"fmt"

	"github.com/f4n4t/go-release/pkg/utils"
)

// useParallelRead determines if the parallel file reading method should be used based on the service's read mode and storage type.
// It takes the release path as input and assesses the parallel read capability based on predefined modes or SSD detection.
func (s *Service) useParallelRead(releasePath string) (bool, error) {
	switch s.parallelFileRead {
	case ParallelFileReadDisabled:
		s.log.Debug().Msg("disabling parallel method for reading files")
		return false, nil

	case ParallelFileReadEnabled:
		s.log.Debug().Msg("forcing parallel method for reading files")
		return true, nil

	case ParallelFileReadAuto:
		if utils.IsSSD(releasePath) {
			s.log.Debug().Msg("detected ssd, using faster parallel method for reading files")
			return true, nil
		}

		s.log.Debug().Msg("could not detect ssd, using traditional method for reading files")
		return false, nil

	default:
		return false, fmt.Errorf("invalid parallel read mode: %d", s.parallelFileRead)
	}
}
