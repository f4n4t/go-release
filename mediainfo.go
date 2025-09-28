package release

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const (
	// General represents the media information type for general metadata.
	General MediaInfoType = "General"

	// Video represents the media information type for video-specific metadata.
	Video MediaInfoType = "Video"

	// Audio represents the media information type for audio-specific metadata.
	Audio MediaInfoType = "Audio"

	// Text represents the media information type for text-based metadata (subtitles).
	Text MediaInfoType = "Text"

	// Menu represents the media information type for menu-specific metadata (chapters).
	Menu MediaInfoType = "Menu"
)

// MediaInfoType defines a string type used to specify the category of media information metadata.
type MediaInfoType string

type MediaInfo struct {
	CreatingLibrary CreatingLibrary `json:"creatingLibrary"`
	Media           Media           `json:"media"`
}

type CreatingLibrary struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	URL     string `json:"url"`
}

type Media struct {
	Ref    string           `json:"@ref"`
	Tracks []MediaInfoTrack `json:"track"`
}

type MediaInfoTrack struct {
	Type                           string              `json:"@type"`
	UniqueID                       string              `json:"UniqueID"`
	VideoCount                     string              `json:"VideoCount,omitempty"`
	AudioCount                     string              `json:"AudioCount,omitempty"`
	TextCount                      string              `json:"TextCount,omitempty"`
	Format                         string              `json:"Format"`
	FormatVersion                  string              `json:"Format_Version,omitempty"`
	FileSize                       string              `json:"FileSize,omitempty"`
	Duration                       string              `json:"Duration"`
	OverallBitRate                 string              `json:"OverallBitRate,omitempty"`
	FrameRate                      string              `json:"FrameRate"`
	FrameCount                     string              `json:"FrameCount"`
	StreamSize                     string              `json:"StreamSize"`
	IsStreamable                   string              `json:"IsStreamable,omitempty"`
	EncodedDate                    string              `json:"Encoded_Date,omitempty"`
	EncodedApplication             string              `json:"Encoded_Application,omitempty"`
	EncodedLibrary                 string              `json:"Encoded_Library,omitempty"`
	StreamOrder                    string              `json:"StreamOrder,omitempty"`
	ID                             string              `json:"ID,omitempty"`
	FormatProfile                  string              `json:"Format_Profile,omitempty"`
	FormatLevel                    string              `json:"Format_Level,omitempty"`
	FormatSettingsCABAC            string              `json:"Format_Settings_CABAC,omitempty"`
	FormatSettingsRefFrames        string              `json:"Format_Settings_RefFrames,omitempty"`
	CodecID                        string              `json:"CodecID,omitempty"`
	BitRate                        string              `json:"BitRate,omitempty"`
	Width                          string              `json:"Width,omitempty"`
	Height                         string              `json:"Height,omitempty"`
	StoredHeight                   string              `json:"Stored_Height,omitempty"`
	SampledWidth                   string              `json:"Sampled_Width,omitempty"`
	SampledHeight                  string              `json:"Sampled_Height,omitempty"`
	PixelAspectRatio               string              `json:"PixelAspectRatio,omitempty"`
	DisplayAspectRatio             string              `json:"DisplayAspectRatio,omitempty"`
	FrameRateMode                  string              `json:"FrameRate_Mode,omitempty"`
	ColorSpace                     string              `json:"ColorSpace,omitempty"`
	ChromaSubsampling              string              `json:"ChromaSubsampling,omitempty"`
	BitDepth                       string              `json:"BitDepth,omitempty"`
	ScanType                       string              `json:"ScanType,omitempty"`
	Delay                          string              `json:"Delay,omitempty"`
	Default                        string              `json:"Default,omitempty"`
	Forced                         string              `json:"Forced,omitempty"`
	ColourDescriptionPresent       string              `json:"colour_description_present,omitempty"`
	ColourDescriptionPresentSource string              `json:"colour_description_present_Source,omitempty"`
	ColourRange                    string              `json:"colour_range,omitempty"`
	ColourRangeSource              string              `json:"colour_range_Source,omitempty"`
	ColourPrimaries                string              `json:"colour_primaries,omitempty"`
	ColourPrimariesSource          string              `json:"colour_primaries_Source,omitempty"`
	TransferCharacteristics        string              `json:"transfer_characteristics,omitempty"`
	TransferCharacteristicsSource  string              `json:"transfer_characteristics_Source,omitempty"`
	MatrixCoefficients             string              `json:"matrix_coefficients,omitempty"`
	MatrixCoefficientsSource       string              `json:"matrix_coefficients_Source,omitempty"`
	TypeOrder                      string              `json:"typeorder,omitempty"`
	FormatCommercialIfAny          string              `json:"Format_Commercial_IfAny,omitempty"`
	FormatSettingsEndianness       string              `json:"Format_Settings_Endianness,omitempty"`
	FormatAdditionalFeatures       string              `json:"Format_AdditionalFeatures,omitempty"`
	BitRateMode                    string              `json:"BitRate_Mode,omitempty"`
	Channels                       string              `json:"Channels,omitempty"`
	ChannelPositions               string              `json:"ChannelPositions,omitempty"`
	ChannelLayout                  string              `json:"ChannelLayout,omitempty"`
	SamplesPerFrame                string              `json:"SamplesPerFrame,omitempty"`
	SamplingRate                   string              `json:"SamplingRate,omitempty"`
	SamplingCount                  string              `json:"SamplingCount,omitempty"`
	CompressionMode                string              `json:"Compression_Mode,omitempty"`
	DelaySource                    string              `json:"Delay_Source,omitempty"`
	StreamSizeProportion           string              `json:"StreamSize_Proportion,omitempty"`
	Language                       string              `json:"Language,omitempty"`
	ServiceKind                    string              `json:"ServiceKind,omitempty"`
	Extra                          MediaInfoTrackExtra `json:"extra"`
	ElementCount                   string              `json:"ElementCount,omitempty"`
	Title                          string              `json:"Title,omitempty"`
}

type MediaInfoTrackExtra struct {
	Attachments             string `json:"Attachments,omitempty"`
	ComplexityIndex         string `json:"ComplexityIndex"`
	NumberOfDynamicObjects  string `json:"NumberOfDynamicObjects"`
	BedChannelCount         string `json:"BedChannelCount"`
	BedChannelConfiguration string `json:"BedChannelConfiguration"`
	Bsid                    string `json:"bsid"`
	Dialnorm                string `json:"dialnorm"`
	Compr                   string `json:"compr"`
	Acmod                   string `json:"acmod"`
	Lfeon                   string `json:"lfeon"`
	DialnormAverage         string `json:"dialnorm_Average"`
	DialnormMinimum         string `json:"dialnorm_Minimum"`
	ComprAverage            string `json:"compr_Average"`
	ComprMinimum            string `json:"compr_Minimum"`
	ComprMaximum            string `json:"compr_Maximum"`
	ComprCount              string `json:"compr_Count"`
}

// GetAttachmentNames retrieves the names of attachments filtered by optional file extensions from the collection of tracks.
// Extensions need to be a list of lowercase file extensions with the leading dot, e.g. ".nfo", ".srt", ".jpg".
// If no extensions are provided, all attachment names are returned.
func (m *MediaInfo) GetAttachmentNames(extensions ...string) []string {
	if len(m.Media.Tracks) == 0 {
		return nil
	}

	var names []string

	shouldIncludeFile := func(fileName string) bool {
		if len(extensions) == 0 {
			return true
		}

		return slices.Contains(extensions, strings.ToLower(filepath.Ext(fileName)))
	}

	for _, track := range m.Media.Tracks {
		if track.Type != string(General) {
			continue
		}

		if strings.TrimSpace(track.Extra.Attachments) == "" {
			continue
		}

		for _, attachment := range strings.Split(track.Extra.Attachments, "/") {
			fileName := strings.TrimSpace(attachment)

			if shouldIncludeFile(fileName) {
				names = append(names, fileName)
			}
		}
	}

	return names
}

// HasAnyLanguage checks if any audio tracks in the collection match the specified languages (case-insensitive).
func (m *MediaInfo) HasAnyLanguage(languages ...string) bool {
	if len(m.Media.Tracks) == 0 {
		return false
	}

	for _, track := range m.Media.Tracks {
		if track.Type != string(Audio) {
			continue
		}

		exists := slices.ContainsFunc(languages, func(s string) bool {
			return strings.EqualFold(s, track.Language)
		})
		if exists {
			return true
		}
	}
	return false
}

func (m *MediaInfo) GetNearestResolution() Resolution {
	if len(m.Media.Tracks) == 0 {
		return ""
	}

	type resolutionDimension struct {
		resolution Resolution
		width      int
		height     int
	}

	standardResolutions := []resolutionDimension{
		{SD, 640, 480},  // VGA
		{SD, 854, 480},  // FWVGA
		{SD, 720, 576},  // Standard Definition (PAL)
		{SD, 720, 480},  // Standard Definition (NTSC)
		{SD, 960, 540},  // qHD
		{HD, 960, 720},  // HD720
		{HD, 1280, 720}, // HD (720p)
		{HD, 1280, 800}, // WXGA
		{HD, 1366, 768}, // WXGA (Widescreen Extended Graphics Array)
		{HD, 1152, 720}, // Proportional widescreen 720p
		{HD, 1280, 768}, // WXGA (16:9 aspect ratio)
		{HD, 1280, 800}, // WXGA (16:10 aspect ratio)
		{FHD, 1440, 900},
		{FHD, 1440, 1080}, // Anamorphic Full HD
		{FHD, 1600, 900},  // HD+
		{FHD, 1920, 1080}, // Full HD (1080p)
		{UHD, 3840, 2160}, // Ultra HD (4K)
	}

	var (
		width, height int
		videoFound    bool
	)

	for _, track := range m.Media.Tracks {
		if track.Type == string(Video) {
			widthVal, widthErr := strconv.Atoi(track.Width)
			heightVal, heightErr := strconv.Atoi(track.Height)

			if widthErr == nil && heightErr == nil && widthVal > 0 && heightVal > 0 {
				width = widthVal
				height = heightVal
				videoFound = true
				break
			}
		}
	}

	if !videoFound {
		return ""
	}

	// UHD is the highest resolution, so check for that first
	if height >= 2160 {
		return UHD
	}

	var (
		closestResolution Resolution
		minDistance       = float64(1<<31 - 1) // set big minDistance
	)

	for _, std := range standardResolutions {
		widthDiff := float64(width - std.width)
		heightDiff := float64(height - std.height)
		distance := (widthDiff * widthDiff) + (heightDiff * heightDiff)

		if distance < minDistance {
			minDistance = distance
			closestResolution = std.resolution
			if minDistance == 0 {
				break
			}
		}
	}

	return closestResolution
}

// MediaInfoBinary checks for the existence of tsmedia or mediainfo-rar in Path.
func MediaInfoBinary() (string, error) {
	for _, binary := range []string{"tsmedia", "mediainfo-rar", "mediainfo"} {
		if binaryPath, err := exec.LookPath(binary); err == nil && binaryPath != "" {
			return binaryPath, nil
		}
	}

	return "", errors.New("no binary for mediainfo generation found")
}

// GenerateMediaInfo calls tsmedia or mediainfo-rar to generate mediainfo output for the biggest file in release.
// returns the JSON output and MediaInfo, potentially an error.
func GenerateMediaInfo(mediaFile string) ([]byte, *MediaInfo, error) {
	binaryPath, err := MediaInfoBinary()
	if err != nil {
		return nil, nil, err
	}

	var args []string

	switch filepath.Base(binaryPath) {
	case "tsmedia":
		args = []string{"-q", "-o", "JSON", "--", mediaFile}

	case "mediainfo-rar":
		args = []string{"--Output=JSON", "--", mediaFile}

	case "mediainfo":
		args = []string{"--Output=JSON", "--", mediaFile}

	default:
		return nil, nil, fmt.Errorf("unknown mediainfo binary: %s", binaryPath)
	}

	jsonOutput, err := exec.Command(binaryPath, args...).Output()
	if err != nil {
		return nil, nil, fmt.Errorf("error running mediainfo: %w", err)
	}

	mediaInfo := &MediaInfo{}

	if err := json.Unmarshal(jsonOutput, &mediaInfo); err != nil {
		return nil, nil, fmt.Errorf("unmarshal mediainfo: %w", err)
	}

	return jsonOutput, mediaInfo, nil
}
