package predbnet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	apiBaseURL  = "https://api.predb.net/"
	userAgent   = "go-release"
	httpTimeout = 5 * time.Second
)

var (
	ErrEmptyName    = errors.New("empty search name")
	ErrNothingFound = errors.New("nothing found")
)

// GetWithContext retrieves the release information by its name using an HTTP request, utilizing the provided context.
func GetWithContext(ctx context.Context, name string) (Release, error) {
	if name == "" {
		return Release{}, ErrEmptyName
	}

	req, err := buildSearchRequest(name)
	if err != nil {
		return Release{}, fmt.Errorf("build http request: %w", err)
	}

	req = req.WithContext(ctx)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Release{}, fmt.Errorf("%w for %s", ErrNothingFound, name)
	} else if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("unknown status code: %s", http.StatusText(resp.StatusCode))
	}

	var result Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Release{}, fmt.Errorf("decode json: %w", err)
	}

	return result.Data.Get(name)
}

// Get searches for available pre on predb.net
func Get(name string) (Release, error) {
	if name == "" {
		return Release{}, ErrEmptyName
	}

	req, err := buildSearchRequest(name)
	if err != nil {
		return Release{}, fmt.Errorf("build http request: %w", err)
	}

	client := &http.Client{
		Timeout: httpTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Release{}, fmt.Errorf("%w for %s", ErrNothingFound, name)
	} else if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("unknown status code: %s", http.StatusText(resp.StatusCode))
	}

	var result Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Release{}, fmt.Errorf("decode json: %w", err)
	}

	return result.Data.Get(name)
}

// buildSearchRequest constructs and returns an HTTP GET request for searching a name on the predb.net API.
func buildSearchRequest(name string) (*http.Request, error) {
	v := url.Values{}
	v.Add("q", name)
	// use "type search", because "type pre" has longer load times
	v.Add("type", "search")

	searchURL := apiBaseURL + "?" + v.Encode()

	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	return req, nil
}
