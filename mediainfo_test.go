package release

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNearestResolution(t *testing.T) {
	tests := []struct {
		desc      string
		mediaInfo *MediaInfo
		expected  Resolution
	}{
		{
			desc: "no video tracks",
			mediaInfo: &MediaInfo{
				Media: Media{
					Tracks: []MediaInfoTrack{
						{Type: "Audio"},
					},
				},
			},
			expected: "",
		},
		{
			desc: "invalid dimensions",
			mediaInfo: &MediaInfo{
				Media: Media{
					Tracks: []MediaInfoTrack{
						{Type: "Video", Width: "-1", Height: "-1"},
					},
				},
			},
			expected: "",
		},
		{
			desc: "standard HD resolution",
			mediaInfo: &MediaInfo{
				Media: Media{
					Tracks: []MediaInfoTrack{
						{Type: "Video", Width: "1280", Height: "720"},
					},
				},
			},
			expected: HD,
		},
		{
			desc: "anamorphic full hd 4:3",
			mediaInfo: &MediaInfo{
				Media: Media{
					Tracks: []MediaInfoTrack{
						{Type: "Video", Width: "1440", Height: "1080"},
					},
				},
			},
			expected: FHD,
		},
		{
			desc: "non-standard resolution closer to FHD",
			mediaInfo: &MediaInfo{
				Media: Media{
					Tracks: []MediaInfoTrack{
						{Type: "Video", Width: "1900", Height: "1080"},
					},
				},
			},
			expected: FHD,
		},
		{
			desc: "non-standard resolution closer to UHD",
			mediaInfo: &MediaInfo{
				Media: Media{
					Tracks: []MediaInfoTrack{
						{Type: "Video", Width: "3700", Height: "2100"},
					},
				},
			},
			expected: UHD,
		},
		{
			desc: "qHD resolution",
			mediaInfo: &MediaInfo{
				Media: Media{
					Tracks: []MediaInfoTrack{
						{Type: "Video", Width: "960", Height: "540"},
					},
				},
			},
			expected: SD,
		},
		{
			desc: "resolution below SD",
			mediaInfo: &MediaInfo{
				Media: Media{
					Tracks: []MediaInfoTrack{
						{Type: "Video", Width: "640", Height: "480"},
					},
				},
			},
			expected: SD,
		},
		{
			desc: "multiple tracks with valid video",
			mediaInfo: &MediaInfo{
				Media: Media{
					Tracks: []MediaInfoTrack{
						{Type: "Audio", Width: "0", Height: "0"},
						{Type: "Video", Width: "1920", Height: "1080"},
						{Type: "Audio", Width: "0", Height: "0"},
					},
				},
			},
			expected: FHD,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			result := test.mediaInfo.GetNearestResolution()
			assert.Equal(t, test.expected, result)
		})
	}
}
