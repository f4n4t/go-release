package srrdb

import "errors"

var (
	// ErrCRCMismatch crc doesn't match
	ErrCRCMismatch = errors.New("rescene: crc error")

	// ErrBadBlock block not formatted properly
	ErrBadBlock = errors.New("rescene: block not properly formatted")

	// ErrBadFile file not properly formatted
	ErrBadFile = errors.New("rescene: file not properly formatted")

	// ErrBadData invalid data
	ErrBadData = errors.New("rescene: incorrect data")

	// ErrNoData data missing
	ErrNoData = errors.New("rescene: no data")

	// ErrDuplicateSFV file referenced twice in sfv
	ErrDuplicateSFV = errors.New("rescene: duplicate file in sfv")
)
