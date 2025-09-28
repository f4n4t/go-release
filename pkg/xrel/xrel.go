package xrel

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
	apiBaseURL  = "https://xrel-api.nfos.to/v2/release/info.json"
	userAgent   = "go-release"
	httpTimeout = 10 * time.Second
)

var (
	ErrNothingFound = errors.New("nothing found")
)

// GetWithContext performs an HTTP GET request to fetch release data by name, using the provided context for cancellation.
func GetWithContext(ctx context.Context, name string) (Release, error) {
	if name == "" {
		return Release{}, errors.New("search name cannot be empty")
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

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return Release{}, fmt.Errorf("%w for %s", ErrNothingFound, name)
		default:
			return Release{}, fmt.Errorf("unknown status code: %s", http.StatusText(resp.StatusCode))
		}
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Release{}, fmt.Errorf("decode json: %w", err)
	}

	return release, nil
}

// Get retrieves release information for the given directory name by making a request to the xrel.to API.
func Get(name string) (Release, error) {
	if name == "" {
		return Release{}, errors.New("search name cannot be empty")
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

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return Release{}, fmt.Errorf("%w for %s", ErrNothingFound, name)
		default:
			return Release{}, fmt.Errorf("unknown status code: %s", http.StatusText(resp.StatusCode))
		}
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Release{}, fmt.Errorf("decode json: %w", err)
	}

	return release, nil
}

// buildSearchRequest constructs an HTTP GET request to search for a directory name using the xrel API.
func buildSearchRequest(name string) (*http.Request, error) {
	v := url.Values{
		"dirname": []string{name},
	}

	searchURL := apiBaseURL + "?" + v.Encode()

	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	return req, nil
}
