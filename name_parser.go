package release

import (
	_ "embed"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Resolution represents the video resolution quality
type Resolution string

const (
	SD  Resolution = "sd"
	HD  Resolution = "720p"
	FHD Resolution = "1080p"
	UHD Resolution = "2160p"
)

// Section represents the category of a release
type Section string

// Application categories
const (
	AppsMisc    Section = "apps-misc"
	AppsMacOS   Section = "apps-macos"
	AppsLinux   Section = "apps-linux"
	AppsWindows Section = "apps-windows"
)

// Game categories
const (
	GamesWindows     Section = "games-windows"
	GamesMacOS       Section = "games-macos"
	GamesLinux       Section = "games-linux"
	GamesPlaystation Section = "games-playstation"
	GamesNintendo    Section = "games-nintendo"
	GamesXbox        Section = "games-xbox"
)

// Audio categories
const (
	AudioBooks Section = "abooks"
	AudioFLAC  Section = "flac"
	AudioMP3   Section = "mp3"
	AudioVideo Section = "mvid"
)

// Video categories
const (
	Movies Section = "movies"
	TV     Section = "tv"
	TVPack Section = "tv-pack"
	Sport  Section = "sport"
)

// Adult content categories
const (
	XXX          Section = "xxx"
	XXXImagesets Section = "xxx-imagesets"
	XXXClips     Section = "xxx-clips"
	XXXDVD       Section = "xxx-dvd"
	XXXPack      Section = "xxx-pack"
	XXXMovies    Section = "xxx-movies"
)

// Miscellaneous categories
const (
	Tutorials Section = "tutorials"
	Mobile    Section = "mobile"
	Ebooks    Section = "ebooks"
	Unknown   Section = "unknown"
)

// sportSections is a text file that holds patterns for sport sections
//
//go:embed sport_patterns.txt
var sportSections []byte

// languages is a slice with all the languages to check for in the release name
var languages = []string{"danish", "dutch", "finnish", "french", "german", "norwegian", "spanish", "swedish", "hebrew"}

// sectionRegexes holds patterns for identifying different section types
var sectionRegexes = struct {
	musicSource, videoSource, videoCodec, oldVideo, xxxImageset, tv, ebook, game, gameSection, mobile, tutorial, macOS, linux *regexp.Regexp
}{
	musicSource: regexp.MustCompile(`(?i)[_-](web|sat|dvbs|cable|\d*cd[mrs]?|cdep|dvd(rip)?|mbluray|vinyl|vls|tape|sacd)[_-]`),
	videoSource: regexp.MustCompile(`(?i)\.(atv|dtv|hdtv|dvd[59]|bdrip|uhdbdrip|bluray|hddvd|web|vhs|hd2dvd)(rip)?[.-]`),
	videoCodec:  regexp.MustCompile(`(?i)[._-]([xh]26[45]|avc|hevc|vp9|divx|xvid|mpeg2|mp4|vc1|wmv)[._-]`),
	oldVideo:    regexp.MustCompile(`(?i)[._-](hevc|avc|xvid|divx|vc1|[xh][._]?26[45]|m?dvd[59]?r?|mp4|mpeg2|m?bluray|hd2?dvd|720[ip]|1080[ip]|2160[ip]|xxx)([._-]|$)`),
	xxxImageset: regexp.MustCompile(`(?i)xxx[._]imageset`),
	tv:          regexp.MustCompile(`(?i)[._]s\d{2}[de]\d{2,}|s\d{2}|\d{4}[._-]\d{2}[._-]\d{2}|[._]\dx\d{2}[._]|[._]s\d{4}e\d{2}[._]|[._]e\d{2,}[._]|[._]d\d{2}[._]`),
	ebook:       regexp.MustCompile(`(?i)[._-](ebook|epub|pdf|cbr|cbz)([._-]|$)`),
	game:        regexp.MustCompile(`(?i)[._-](ps[1-5]|xbox(one|360)?|nsw|wiiu?|linux)[._-]`),
	gameSection: regexp.MustCompile(`(?i)^(games|ps\d|playstation|wii|xbox|x360|nsw|nintendo|nds)`),
	mobile:      regexp.MustCompile(`(?i)[._-]android[._-]|apk$`),
	tutorial:    regexp.MustCompile(`(?i)([._-]tutorials?|lectures?[._-])|^udemy[._-]`),
	macOS:       regexp.MustCompile(`(?i)[._-]macosx?[._-]`),
	linux:       regexp.MustCompile(`(?i)[._-]linux[._-]`),
}

// videoRegexes holds patterns for identifying video content types
var videoRegexes = struct {
	xxx, imageSet, clips, dvd, pack, tvPack, noTvPack, mvid, noSport *regexp.Regexp
}{
	xxx:      regexp.MustCompile(`(?i)[._]xxx[._]?`),
	imageSet: regexp.MustCompile(`(?i)[._]imagesets?[._-]?`),
	clips:    regexp.MustCompile(`(?i)(\d{2}[._]){3}|[._]\d{4}[._]`),
	dvd:      regexp.MustCompile(`(?i)[._]dvd[59r]?([._-]|$)`),
	pack:     regexp.MustCompile(`(?i)[._]pack[._-]`),
	tvPack:   regexp.MustCompile(`(?i)[._](s\d{2})[._]`),
	noTvPack: regexp.MustCompile(`(?i)[._]complete.*(bluray|dvd)`),
	mvid:     regexp.MustCompile(`(?i)-\d{4}-|[._-](mbluray|[ck]on[cz]ert)[._-]`),
	noSport:  regexp.MustCompile(`(?i)[._-](do[ck]u(mentation)?|(s(taffel)?\d+)?e(pisode)?\d+)[._-]`),
}

// audioRegexes holds patterns for identifying audio content types
var audioRegexes = struct {
	flac, aBook *regexp.Regexp
}{
	flac:  regexp.MustCompile(`(?i)[_-]flac[_-]`),
	aBook: regexp.MustCompile(`(?i)[_-](abook|audiobook|hoerbuch)`),
}

// gameRegexes holds patterns for identifying game platforms
var gameRegexes = struct {
	wii, playStation *regexp.Regexp
}{
	wii:         regexp.MustCompile(`(?i)wiiu?|nsw`),
	playStation: regexp.MustCompile(`(?i)[._-](ps|playstation)[1-5][._-]`),
}

// resRegexes holds patterns for identifying video resolutions
var resRegexes = struct {
	fhd, ultraHD *regexp.Regexp
}{
	fhd:     regexp.MustCompile(`(?i)complete[._-]m?bluray|[._-]fhd(2[45]p)?[._-]`),
	ultraHD: regexp.MustCompile(`(?i)([._-]uhd[._-]m?bluray)|[._]4k[._]bluray`),
}

// ParseSection tries to determine the section for the given release name
func (s *Service) ParseSection(name string, preInfo *Pre) Section {
	name = strings.ToLower(name)
	preSection := ""

	if preInfo != nil {
		preSection = strings.ToLower(preInfo.Section)
	}

	// Try primary section detection
	section := s.detectPrimarySection(name, preSection)

	// Try fallback detection if primary detection failed
	if section == Unknown {
		section = s.detectFallbackSection(name)
	}

	return section
}

// detectPrimarySection attempts to identify the section based on common patterns
func (s *Service) detectPrimarySection(name string, preSection string) Section {
	switch {
	case sectionRegexes.xxxImageset.MatchString(name):
		return XXXImagesets
	case sectionRegexes.musicSource.MatchString(name):
		return parseAudio(name)
	case sectionRegexes.oldVideo.MatchString(name):
		return s.parseVideo(name, preSection)
	case sectionRegexes.ebook.MatchString(name):
		return Ebooks
	case preSection != "":
		return s.detectSectionFromPreSection(name, preSection)
	}

	return Unknown
}

// detectSectionFromPreSection uses pre-release information to help identify the section
func (s *Service) detectSectionFromPreSection(name string, preSection string) Section {
	switch {
	case sectionRegexes.gameSection.MatchString(preSection):
		return parseGame(name, true)
	case slices.Contains([]string{"0day", "apps"}, preSection):
		return parseApp(name)
	case slices.Contains([]string{"abooks", "abook", "mp3", "flac"}, preSection):
		return parseAudio(name)
	}
	// slices.ContainsFunc([]string{"abook", "mp3", "flac"}, func(s string) bool {
	// 	return strings.Contains(s, strings.ToLower(pre.Section))
	// })

	return Unknown
}

// detectFallbackSection tries alternative detection methods
func (s *Service) detectFallbackSection(name string) Section {
	switch {
	case sectionRegexes.tutorial.MatchString(name):
		return Tutorials
	case sectionRegexes.mobile.MatchString(name):
		return Mobile
	case sectionRegexes.game.MatchString(name):
		return parseGame(name, false)
	}

	return Unknown
}

// parseXXXContent identifies the specific type of adult content
func parseXXXContent(name string) Section {
	switch {
	case videoRegexes.imageSet.MatchString(name):
		return XXXImagesets
	case videoRegexes.clips.MatchString(name):
		return XXXClips
	case videoRegexes.dvd.MatchString(name):
		return XXXDVD
	case videoRegexes.pack.MatchString(name):
		return XXXPack
	default:
		return XXXMovies
	}
}

// parseVideo identifies the specific type of video content
func (s *Service) parseVideo(name string, preSection string) Section {
	// Check for adult content first
	if videoRegexes.xxx.MatchString(name) {
		return parseXXXContent(name)
	}

	// Check for sports content
	if !videoRegexes.noSport.MatchString(name) && s.isSport(name) {
		return Sport
	}

	// Check for music video content
	if slices.Contains([]string{"music", "mbluray", "mvid"}, preSection) ||
		videoRegexes.mvid.MatchString(name) {
		return AudioVideo
	}

	// Check for TV pack content
	if videoRegexes.tvPack.MatchString(name) && !videoRegexes.noTvPack.MatchString(name) {
		return TVPack
	}

	// Check for TV content
	if sectionRegexes.tv.MatchString(name) {
		return TV
	}

	// Default to Movies
	return Movies
}

// parseAudio identifies the specific type of audio content
func parseAudio(name string) Section {
	switch {
	case audioRegexes.aBook.MatchString(name):
		return AudioBooks
	case audioRegexes.flac.MatchString(name):
		return AudioFLAC
	case sectionRegexes.videoCodec.MatchString(name):
		return AudioVideo
	default:
		return AudioMP3
	}
}

// parseGame identifies the specific gaming platform
func parseGame(name string, hasPreSection bool) Section {
	switch {
	case strings.Contains(name, "xbox"):
		return GamesXbox
	case gameRegexes.wii.MatchString(name):
		return GamesNintendo
	case gameRegexes.playStation.MatchString(name):
		return GamesPlaystation
	case sectionRegexes.linux.MatchString(name):
		return GamesLinux
	case sectionRegexes.macOS.MatchString(name):
		return GamesMacOS
	case hasPreSection:
		return GamesWindows
	}

	return Unknown
}

// parseApp identifies the specific type of application
func parseApp(name string) Section {
	switch {
	case strings.Contains(name, "crossplatform"):
		return AppsMisc
	case sectionRegexes.macOS.MatchString(name):
		return AppsMacOS
	case sectionRegexes.linux.MatchString(name):
		return AppsLinux
	default:
		return AppsWindows
	}
}

// isSport checks if the name contains any sport pattern
func (s *Service) isSport(name string) bool {
	patterns := s.sportPatterns

	if len(sportSections) > 0 {
		patterns = append(patterns, strings.Split(string(sportSections), "\n")...)
	}

	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		pattern := regexp.MustCompile(fmt.Sprintf("(?i)^%s[._-]", p))
		if pattern.MatchString(name) {
			return true
		}
	}

	return false
}

// ParseResolution determines the video resolution from the release name
func ParseResolution(name string) Resolution {
	name = strings.ToLower(name)

	// Direct resolution matching
	for _, res := range []Resolution{UHD, FHD, HD} {
		if strings.Contains(name, string(res)) {
			return res
		} else if strings.Contains(name, string(res[:len(res)-1]+"i")) {
			// Check for 720i, 1080i, 2160i variants
			return res
		}
	}

	// Pattern-based resolution detection
	switch {
	case resRegexes.fhd.MatchString(name):
		return FHD
	case resRegexes.ultraHD.MatchString(name):
		return UHD
	default:
		return SD
	}
}

// ParseLanguage identifies the language from the release name
func ParseLanguage(name string) string {
	name = strings.ToLower(name)

	for _, lang := range languages {
		if strings.Contains(name, lang) && !strings.Contains(name, ".subbed.") {
			return lang
		}
	}

	return ""
}
