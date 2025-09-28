package srrdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	ReleaseURL     = "https://api.srrdb.com/v1/details/{release}"
	DownloadURL    = "https://www.srrdb.com/download/file/{release}/{file}"
	DownloadSrrURL = "https://www.srrdb.com/download/srr/{release}"
	DownloadAddURL = "https://www.srrdb.com/download/temp/{release}/{id}/{file}"
)

var (
	// DownloadableExtensions defines file extensions that are allowed for download.
	DownloadableExtensions = []string{".nfo", ".m3u", ".jpg", ".sfv"}

	// httpTimeout is the time limit for the http client
	httpTimeout = time.Second * 30

	// ErrFileNotFound indicates that the requested file could not be found during an operation.
	ErrFileNotFound = fmt.Errorf("file not found")
)

// GetInformation fetches and decodes release information from a remote API based on the given release name.
func GetInformation(name string) (Release, error) {
	client := &http.Client{
		Timeout: httpTimeout,
	}

	releaseURL := strings.ReplaceAll(ReleaseURL, "{release}", name)

	resp, err := client.Get(releaseURL)
	if err != nil {
		return Release{}, fmt.Errorf("get release information: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("getting release information failed, status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return Release{}, fmt.Errorf("read response body: %w", err)
	}

	if len(content) == 0 || bytes.Contains(content, []byte("The SRR file does not exist.")) ||
		bytes.Equal(content, []byte("[]")) {
		return Release{}, fmt.Errorf("%s does not exist", releaseURL)
	}

	var info Release

	if err := json.Unmarshal(content, &info); err != nil {
		return Release{}, fmt.Errorf("decode json: %w", err)
	}

	return info, nil
}

// GetFile retrieves the content of a file for the given DownloadRelease configuration via HTTP request.
// It dynamically generates the URL based on the provided release name, file, and ID details.
func GetFile(rel DownloadRelease) ([]byte, error) {
	dlURL, err := rel.buildURL()
	if err != nil {
		return nil, fmt.Errorf("failed to build download URL: %w", err)
	}

	client := &http.Client{
		Timeout: httpTimeout,
	}

	resp, err := client.Get(dlURL)
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getting file failed, status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if len(content) == 0 || bytes.Contains(content, []byte("The SRR file does not exist.")) ||
		bytes.Equal(content, []byte("[]")) {
		return nil, ErrFileNotFound
	}

	return content, nil
}

// GetSrrFile retrieves and unmarshals an SRR file based on the given release name, returning an SrrFile object.
func GetSrrFile(releaseName string) (*SrrFile, error) {
	content, err := GetFile(DownloadRelease{Name: releaseName})
	if err != nil {
		return nil, err
	}

	var srr SrrFile

	if err := srr.Unmarshal(content); err != nil {
		return nil, fmt.Errorf("decode srr file: %w", err)
	}

	return &srr, nil
}

// LoadFromFile reads a file from the given path and unmarshals its content into an SrrFile.
func LoadFromFile(srrFile string) (*SrrFile, error) {
	content, err := os.ReadFile(srrFile)
	if err != nil {
		return nil, err
	}

	var srr SrrFile

	if err := srr.Unmarshal(content); err != nil {
		return nil, fmt.Errorf("decode srr file: %w", err)
	}

	return &srr, nil
}
