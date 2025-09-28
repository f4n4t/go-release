package release

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	// rarFilesPattern is a compiled regular expression that matches file extensions like .rar
	// or sequential archive parts (.rXX, .sXX, .tXX).
	rarFilesPattern = regexp.MustCompile(`^\.([rst]\d+|rar)$`)

	// archiveCountPattern recognizes archive counts in formats like [n/m], (n|m), or {n/m},
	// where n is the current archive number and m is the total archive count.
	// Examples: [01/10], (3|8), {05/12}
	archiveCountPattern = regexp.MustCompile(`[\[\](){}](\d+)[/|](\d+)[\[\](){}]`)
)

var (
	// ErrNoFileCountInDiz represents an error indicating that no file count was found in the .diz metadata.
	ErrNoFileCountInDiz = errors.New("no file count in .diz")

	// ErrNoArchiveInZip represents an error indicating that no archive file was found within the .zip file.
	ErrNoArchiveInZip = errors.New("no archive in .zip")

	// ErrZipValidationFailed represents an error indicating that a check or validation process for a zip file has failed.
	ErrZipValidationFailed = errors.New("zip check failed")
)

type archiveResult struct {
	archives      []archiveInfo
	expectedTotal int
	nfoFile       NFOFile
}

type archiveInfo struct {
	name    string
	size    uint64
	current int
	total   int
}

type archiveCount struct {
	current int
	total   int
}

func (s *Service) CheckZip(rel *Info, extractNFO bool) error {
	var nfoFile NFOFile

	zipFilesByDir := make(map[string][]string)

	for _, file := range rel.Root.GetFiles(".zip") {
		dirName := file.Parent.Info.Name
		zipFilesByDir[dirName] = append(zipFilesByDir[dirName], file.FullPath)
	}

	for dir, files := range zipFilesByDir {
		s.log.Info().Str("folder", dir).Msg("checking zip files")

		result, err := processZipFiles(files)
		if err != nil {
			return err
		}

		if err := validateArchiveCollection(result); err != nil {
			return err
		}

		s.log.Info().Str("folder", dir).Msg("zip check complete")
	}

	if extractNFO && len(nfoFile.Content) > 0 {
		rel.NFO = &nfoFile
	}

	return nil
}

// processZipFiles processes a list of zip file paths to extract archive metadata and locate a valid NFO file.
func processZipFiles(files []string) (archiveResult, error) {
	var (
		nfoFile            NFOFile
		archives           []archiveInfo
		totalExpectedFiles int
	)

	for _, file := range files {
		zipReader, err := zip.OpenReader(file)
		if err != nil {
			return archiveResult{}, fmt.Errorf("read zip file: %w", err)
		}
		// we close zipReader in processZipContents

		extractNFO := len(nfoFile.Content) == 0

		archiveInfo, nfo, err := processZipContents(zipReader, extractNFO)
		if err != nil {
			return archiveResult{}, err
		}

		if extractNFO && len(nfo.Content) > 0 {
			nfoFile = nfo
		}

		if totalExpectedFiles == 0 {
			totalExpectedFiles = archiveInfo.total
		}

		archives = append(archives, archiveInfo)
	}

	result := archiveResult{
		archives:      archives,
		expectedTotal: totalExpectedFiles,
		nfoFile:       nfoFile,
	}

	return result, nil
}

// processZipContents extracts archive and metadata information from a zip file, including NFO content and file count.
func processZipContents(zipReader *zip.ReadCloser, extractNFO bool) (archiveInfo, NFOFile, error) {
	var (
		archiveCount archiveCount
		archive      archiveInfo
		nfoFile      NFOFile
	)

	defer zipReader.Close()

	for _, zipEntry := range zipReader.File {
		ext := strings.ToLower(filepath.Ext(zipEntry.Name))

		if ext == ".diz" {
			content, err := readInnerZip(zipEntry)
			if err != nil {
				return archiveInfo{}, NFOFile{}, err
			}

			count, err := processDizContent(content)
			if err != nil {
				return archiveInfo{}, NFOFile{}, err
			}

			if count.current > 0 {
				archiveCount = count
			}
		} else if ext == ".nfo" && extractNFO {
			content, err := readInnerZip(zipEntry)
			if err != nil {
				return archiveInfo{}, NFOFile{}, err
			}

			if len(content) > 0 {
				nfoFile = NFOFile{
					Name:    zipEntry.Name,
					Content: content,
				}
				extractNFO = false
			}
		} else if rarFilesPattern.MatchString(ext) {
			archive = archiveInfo{
				name: zipEntry.Name,
				size: zipEntry.UncompressedSize64,
			}
		}
	}

	if archiveCount.current == 0 || archiveCount.total == 0 {
		return archiveInfo{}, NFOFile{}, ErrNoFileCountInDiz
	} else if archive == (archiveInfo{}) {
		return archiveInfo{}, NFOFile{}, ErrNoArchiveInZip
	}

	archive.current = archiveCount.current
	archive.total = archiveCount.total

	return archive, nfoFile, nil
}

// readInnerZip reads the content of a zip file entry and returns it as a byte slice or an error if unsuccessful.
func readInnerZip(zipEntry *zip.File) ([]byte, error) {
	f, err := zipEntry.Open()
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", zipEntry.Name, err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", zipEntry.Name, err)
	}

	return content, nil
}

// processDizContent extracts the current and total archive numbers from the given .diz file content or returns an error.
func processDizContent(content []byte) (archiveCount, error) {
	matches := archiveCountPattern.FindSubmatch(content)
	if matches == nil {
		return archiveCount{}, ErrNoFileCountInDiz
	}

	current, err := strconv.Atoi(string(matches[1]))
	if err != nil {
		return archiveCount{}, fmt.Errorf("convert string to int: %w", err)
	}

	total, err := strconv.Atoi(string(matches[2]))
	if err != nil {
		return archiveCount{}, fmt.Errorf("convert string to int: %w", err)
	}

	return archiveCount{current: current, total: total}, nil
}

// validateArchiveCollection validates if the archive collection matches the expected totals and size constraints.
func validateArchiveCollection(result archiveResult) error {
	if len(result.archives) != result.expectedTotal {
		return fmt.Errorf("%w: expected %d archives, got %d",
			ErrZipValidationFailed, result.expectedTotal, len(result.archives))
	}

	uniqueSizes := make(map[uint64]struct{})
	for _, content := range result.archives {
		uniqueSizes[content.size] = struct{}{}
	}

	if len(uniqueSizes) > 2 {
		return fmt.Errorf("%w: more than 2 different archive sizes, something is wrong", ErrZipValidationFailed)
	}

	return nil
}
