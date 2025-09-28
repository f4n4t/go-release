package predbnet

import "fmt"

// Result is the struct that holds the json decoded result from predb.net.
type Result struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Data    Releases `json:"data"`
	Results int      `json:"results"`
	Page    int      `json:"page"`
	Time    string   `json:"time"`
}

// Release is the struct for a single pre
type Release struct {
	ID      int    `json:"id"`
	PreTime int64  `json:"pretime"`
	Release string `json:"release"`
	Section string `json:"section"`
	Files   int    `json:"files"`
	//Size    int64  `json:"size"`
	Status int    `json:"status"`
	Reason string `json:"reason"`
	Group  string `json:"group"`
	Genre  string `json:"genre"`
	URL    string `json:"url"`
	NFO    string `json:"nfo"`
	NFOImg string `json:"nfo_img"`
}

type Releases []Release

// Get searches for a specific pre in the Releases slice
func (r Releases) Get(name string) (Release, error) {
	for _, release := range r {
		if release.Release == name {
			return release, nil
		}
	}
	return Release{}, fmt.Errorf("%w for %s", ErrNothingFound, name)
}
