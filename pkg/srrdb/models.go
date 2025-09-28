package srrdb

import (
	"errors"
	"strconv"
	"strings"
)

type FileLike interface {
	GetSize() int64
}

type Release struct {
	Name          string        `json:"name"`
	Files         Files         `json:"files"`
	ArchivedFiles ArchivedFiles `json:"archived-files"`
	Adds          Adds          `json:"adds"`
}

type Files []File
type ArchivedFiles []ArchivedFile
type Adds []Add

type File struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	CRC  string `json:"crc"`
}

func (f File) GetSize() int64 {
	return f.Size
}

type ArchivedFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	CRC  string `json:"crc"`
}

func (af ArchivedFile) GetSize() int64 {
	return af.Size
}

type Add struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	CRC  string `json:"crc"`
	ID   int    `json:"id"`
}

func (a Add) GetSize() int64 {
	return a.Size
}

func TotalSize[T FileLike](items []T) int64 {
	var totalSize int64
	for _, item := range items {
		totalSize += item.GetSize()
	}
	return totalSize
}

type DownloadRelease struct {
	Name string
	File string
	ID   int
}

// buildURL generates a download URL based on the DownloadRelease fields and predefined URL templates.
func (dr DownloadRelease) buildURL() (string, error) {
	var dlURL string

	switch {
	case dr.ID > 0:
		if dr.Name == "" || dr.File == "" {
			return "", errors.New("both name and file must be present")
		}

		dlURL = strings.ReplaceAll(DownloadAddURL, "{release}", dr.Name)
		dlURL = strings.ReplaceAll(dlURL, "{id}", strconv.Itoa(dr.ID))
		dlURL = strings.ReplaceAll(dlURL, "{file}", dr.File)

	case dr.File != "":
		if dr.Name == "" {
			return "", errors.New("name must be present")
		}

		dlURL = strings.ReplaceAll(DownloadURL, "{release}", dr.Name)
		dlURL = strings.ReplaceAll(dlURL, "{file}", dr.File)

	case dr.Name != "":
		dlURL = strings.ReplaceAll(DownloadSrrURL, "{release}", dr.Name)

	default:
		return "", errors.New("no valid input")
	}

	return dlURL, nil
}
