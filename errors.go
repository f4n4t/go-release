package release

import "errors"

var (
	// ErrEmptyFolder is the error that Parse will return if the folder is empty.
	ErrEmptyFolder = errors.New("empty folder")

	// ErrFileNotFound is the error returned when a specified file cannot be located.
	ErrFileNotFound = errors.New("file not found")

	// ErrForbiddenFiles is the error that Parse will return on forbidden files.
	ErrForbiddenFiles = errors.New("forbidden files")

	// ErrForbiddenExtension is returned when a file has a prohibited or unsupported extension.
	ErrForbiddenExtension = errors.New("forbidden extension")

	// ErrForbiddenCharacters is the error returned when a file contains prohibited or invalid characters.
	ErrForbiddenCharacters = errors.New("forbidden characters")

	// ErrEmptyFile is the error returned when a file is empty.
	ErrEmptyFile = errors.New("empty file")
)
