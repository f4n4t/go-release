package release

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/f4n4t/go-dtree"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	Module = "release"
)

// ParallelFileRead is a type that defines if the files should be read in parallel.
type ParallelFileRead int

const (
	// ParallelFileReadDisabled disables parallel file reading (better for hdds).
	ParallelFileReadDisabled ParallelFileRead = iota
	// ParallelFileReadEnabled enables parallel reading of files (improves performance on ssds).
	ParallelFileReadEnabled
	// ParallelFileReadAuto enables it if system is linux and ssd is detected.
	ParallelFileReadAuto
)

// Regexes is a struct with all compiled regexes used for parsing.
var Regexes = struct {
	Archive, Media, IMDb, CleanTitle, Year, EpisodeSpecial, BadChars, MetaFolders, MetaFiles, Group *regexp.Regexp
}{
	Archive:     regexp.MustCompile(`(?i)\.(rar|r\d{2}|\d{3}|zip|xz|tar|gz)$`),
	Media:       regexp.MustCompile(`(?i)\.(avi|mkv|mpe?g|wm[av]|vob|m2ts|divx|xvid|mp[34]|flac|jpe?g|gif|png|img|iso|aac|m4a|ogg|opus|ra|alac|ape|wv|wav|aiff?|pcm|au|snd)$`),
	IMDb:        regexp.MustCompile(`(?i)imdb.+/(tt\d+)`),
	CleanTitle:  regexp.MustCompile(`(?i)[._-](2160p|1080p|720p|s\d+([ed]\d+)?|e\d+|german|complete|d[vl]|hdr|web(rip)?|internal|ac3d?|unrated|uncut|remastered|stream)[._-].+$`),
	Year:        regexp.MustCompile(`[._ -](\d{4})[._ -]`),
	BadChars:    regexp.MustCompile(`(?i)[^a-z0-9()\[\].\-_+]`),
	MetaFolders: regexp.MustCompile(`(?i)^(sample|subs|proof|cover|extras?|addon)$`),
	MetaFiles:   regexp.MustCompile(`(?i)\.(png|jpe?g|sfv|gif|txt)$`),
	Group:       regexp.MustCompile(`(?i)-([a-z0-9]+(_iNT)?)$`),
}

var (
	// mediaInfoSections contains the sections for which mediainfo will be generated.
	mediaInfoSections = []Section{TV, TVPack, Movies, AudioVideo, Sport, AudioBooks, AudioFLAC, AudioMP3}

	// ForbiddenExtensions holds all the forbidden extensions.
	ForbiddenExtensions = []string{".nzb", ".par2", ".url", ".html", ".srr", ".srs"}

	// PictureExtensions is a struct with known picture extensions.
	PictureExtensions = []string{".jpg", ".jpeg", ".png", ".gif"}

	// AudioExtensions is a struct with known audio extensions.
	AudioExtensions = []string{".mp3", ".aac", ".m4a", ".ogg", ".opus", ".wma", ".ra", ".flac", ".alac", ".ape", ".wv",
		".wav", ".aiff", ".aif", ".pcm", ".au", ".snd"}
)

// skipType defines a type used to represent various skipping behaviors for files or directories.
type skipType int

const (
	// skipNothing represents the default skip type where no skipping is applied.
	skipNothing skipType = iota
	// skipFile indicates that the current file should be skipped during processing.
	skipFile
	// skipDir indicates that the current directory should be skipped during processing.
	skipDir
)

type Service struct {
	log              zerolog.Logger
	sportPatterns    []string
	skipPre          bool
	skipMediaInfo    bool
	parallelFileRead ParallelFileRead
	hashThreads      int
	preInfo          *Pre
}

// ServiceBuilder is a builder for the Service.
type ServiceBuilder struct {
	service Service
}

// NewServiceBuilder creates a new ServiceBuilder.
func NewServiceBuilder() *ServiceBuilder {
	sb := &ServiceBuilder{}
	sb.service.log = log.Logger.With().Str("module", Module).Logger()
	return sb
}

// WithSportPatterns sets the sport patterns.
func (s *ServiceBuilder) WithSportPatterns(patterns []string) *ServiceBuilder {
	s.service.sportPatterns = patterns
	return s
}

// WithSkipPre sets the skipPre flag to enable or disable searching for pre-information.
func (s *ServiceBuilder) WithSkipPre(skip bool) *ServiceBuilder {
	s.service.skipPre = skip
	return s
}

// WithSkipMediaInfo sets the skipMediaInfo flag to enable or disable mediainfo generation.
func (s *ServiceBuilder) WithSkipMediaInfo(skip bool) *ServiceBuilder {
	s.service.skipMediaInfo = skip
	return s
}

// WithPreInfo sets the preInfo in advance and skips the pre-search.
func (s *ServiceBuilder) WithPreInfo(preInfo *Pre) *ServiceBuilder {
	if preInfo == nil {
		return s
	}
	s.service.preInfo = preInfo
	s.service.skipPre = true
	return s
}

// WithParallelFileRead sets the parallelFileRead flag to enable or disable parallel file read mode.
// In this package it is used for the CRC32 calculation.
func (s *ServiceBuilder) WithParallelFileRead(i int) *ServiceBuilder {
	switch ParallelFileRead(i) {
	case ParallelFileReadAuto:
		s.service.parallelFileRead = ParallelFileReadAuto
	case ParallelFileReadEnabled:
		s.service.parallelFileRead = ParallelFileReadEnabled
	case ParallelFileReadDisabled:
		s.service.parallelFileRead = ParallelFileReadDisabled
	default:
		// should never happen when we use the constants
		panic(fmt.Sprintf("invalid parallel file read mode: %d", i))
	}

	return s
}

// WithHashThreads sets the number of threads to use for CRC32 checking.
func (s *ServiceBuilder) WithHashThreads(i int) *ServiceBuilder {
	s.service.hashThreads = max(0, i)
	return s
}

// Build creates a new Service from the builder.
func (s *ServiceBuilder) Build() *Service {
	return &Service{
		log:              s.service.log,
		sportPatterns:    s.service.sportPatterns,
		skipPre:          s.service.skipPre,
		skipMediaInfo:    s.service.skipMediaInfo,
		parallelFileRead: s.service.parallelFileRead,
		hashThreads:      s.service.hashThreads,
		preInfo:          s.service.preInfo,
	}
}

// Info represents the main struct with all the additional information.
type Info struct {
	// ArchiveCount is the total count of archive files (files which matched the archive pattern).
	ArchiveCount int `json:"archive_count"`
	// BiggestFile is the largest file found in the release.
	BiggestFile *dtree.Node `json:"-"`
	// Episodes is a slice with all matched Episodes (only media files).
	Episodes []Episode `json:"episodes"`
	// Extensions is a map with all the found file extensions and their count.
	Extensions map[string]int `json:"extensions"`
	// BaseDir is the base directory path of the release.
	BaseDir string `json:"base_dir"`
	// Root is the root node of the directory tree.
	Root *dtree.Node `json:"-"`
	// ForbiddenFiles is a slice with all the files that are either empty files or folders, have bad chars or matched the ForbiddenExtensions slice.
	ForbiddenFiles ForbiddenFiles `json:"-"`
	// Group is the name of the release group (final part of the release after the -).
	Group string `json:"group"`
	// ImdbID is the parsed IMDB ID from the NFO file.
	ImdbID int `json:"imdb_id"`
	// MediaFiles is a slice with all the media files (files that matched the media pattern).
	MediaFiles MediaFiles `json:"-"`
	// MediaInfo is only generated if mediainfo is found in a path.
	MediaInfo *MediaInfo `json:"-"`
	// MediaInfoJSON contains the raw JSON output from mediainfo.
	MediaInfoJSON []byte `json:"-"`
	// Name is the release name (basename of directory or file without extension).
	Name string `json:"name"`
	// PreInfo is a pointer to the Pre information if something is found.
	PreInfo *Pre `json:"-"`
	// ProductTitle is the title without all the additional meta-tags.
	ProductTitle string `json:"product_title"`
	// ProductYear is the year found in the release name.
	ProductYear int `json:"product_year"`
	// Section is the parsed section category of the release.
	Section Section `json:"section"`
	// SfvCount is the count of all the .sfv files.
	SfvCount int `json:"sfv_count"`
	// Size is the total size of the release in bytes.
	Size int64 `json:"size"`
	// Language is the parsed language tag from the release name.
	Language string `json:"language"`
	// TagResolution is the parsed resolution tag from the release name.
	TagResolution Resolution `json:"tag_resolution"`
	// IsSingleFile is true when the root is a file rather than a directory.
	IsSingleFile bool `json:"single_file"`
	// NFO holds the name and content of an NFO file if one is found.
	NFO *NFOFile `json:"-"`
	// parents map is used internally to build the directory tree.
	parents map[string]*dtree.Node
}

func (i *Info) HasNuke() bool {
	return i.PreInfo != nil && i.PreInfo.Nuke != ""
}

func (i *Info) GetPre() (*Pre, bool) {
	if i.PreInfo != nil {
		return i.PreInfo, true
	}
	return nil, false
}

type MediaFiles []*dtree.Node

func (mf MediaFiles) GetByExtensions(extensions ...string) []*dtree.Node {
	if len(extensions) == 0 {
		return nil
	}

	var mediaFiles []*dtree.Node

	for _, f := range mf {
		exists := slices.ContainsFunc(extensions, func(e string) bool {
			return strings.EqualFold(e, f.Info.Extension)
		})
		if exists {
			mediaFiles = append(mediaFiles, f)
		}
	}

	return mediaFiles
}

// ForbiddenFile represents a forbidden file like empty files or not allowed extension.
type ForbiddenFile struct {
	FullPath string
	Info     *dtree.FileInfo
	Error    error
}

type ForbiddenFiles []ForbiddenFile

// Names returns only the base names of the forbidden files as a slice.
func (ff *ForbiddenFiles) Names() []string {
	forbiddenList := make([]string, len(*ff))
	for _, f := range *ff {
		forbiddenList = append(forbiddenList, f.Info.Name)
	}
	return forbiddenList
}

// addFile adds a new file to the forbidden files slice.
func (ff *ForbiddenFiles) addFile(fullPath string, fileInfo *dtree.FileInfo, fileError error) {
	*ff = append(*ff, ForbiddenFile{
		FullPath: fullPath,
		Info:     fileInfo,
		Error:    fileError,
	})
}

// isForbidden checks for forbidden extensions
func isForbidden(fi *dtree.FileInfo) bool {
	return slices.Contains(ForbiddenExtensions, fi.Extension)
}

// Episode represents a single episode in a series.
type Episode struct {
	Number int         `json:"number"`
	Name   string      `json:"name"`
	File   *dtree.Node `json:"-"`
}

// NFOFile contains a single nfo file with content and filename.
type NFOFile struct {
	Name    string
	Content []byte
}

// Parse processes a directory structure, extracts information, and builds a tree representation of its contents.
func (s *Service) Parse(root string, ignore ...string) (*Info, error) {
	info, err := s.initReleaseInfo(root)
	if err != nil {
		return nil, err
	}

	walkFunc := func(path string, fileInfo fs.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		return s.processPath(info, path, dtree.FileInfoFromInterface(fileInfo), ignore)
	}

	if err := filepath.Walk(info.BaseDir, walkFunc); err != nil {
		return nil, err
	}

	info.Root = dtree.BuildFileTree(info.parents)

	if info.Root == nil || (info.Root.Info.IsDir && len(info.Root.Children) == 0) {
		return nil, ErrEmptyFolder
	}

	s.checkForEmptySubfolders(info, info.Root)

	// sort media files by name
	sort.Slice(info.MediaFiles, func(i, j int) bool {
		return info.MediaFiles[i].Info.Name < info.MediaFiles[j].Info.Name
	})

	if s.preInfo != nil {
		info.PreInfo = s.preInfo
	} else if !s.skipPre {
		info.PreInfo = s.GetPre(info.Name)
	}

	info.Section = s.ParseSection(info.Name, info.PreInfo)

	// search for episode numbers
	if info.Section == TVPack && len(info.Root.Children) > 1 {
		info.Episodes = getEpisodes(info.MediaFiles, info.Root)

		if !s.skipPre && info.PreInfo == nil && len(info.Episodes) > 1 {
			firstChild := info.Root.Children[0]
			if firstChild.Info.IsDir {
				s.log.Debug().Str("name", firstChild.Info.Name).
					Msg("trying to search for pre information with sub folder name")
				info.PreInfo = s.GetPre(firstChild.Info.Name)
			}
		}
	}

	// unusual group name != [a-z0-9]
	if info.Group == "" && info.PreInfo != nil {
		if g := strings.TrimSpace(info.PreInfo.Group); g != "" {
			info.Group = g
		}
	}

	if !s.skipMediaInfo && slices.Contains(mediaInfoSections, info.Section) && len(info.MediaFiles) > 0 {
		s.tryGenerateMediaInfo(info)
	}

	if info.MediaInfo != nil {
		// get nfo from .mkv container, uses https://github.com/remko/go-mkvparse
		if info.NFO == nil {
			s.tryExtractNFO(info)
		}
	}

	s.log.Debug().Str("Name", info.Name).
		Any("Section", info.Section).
		Msg("parsed release")

	if len(info.ForbiddenFiles) > 0 {
		return info, ErrForbiddenFiles
	}

	return info, nil
}

// tryGenerateMediaInfo attempts to generate MediaInfo for the provided context and logs relevant actions or errors.
func (s *Service) tryGenerateMediaInfo(info *Info) {
	var mediaFile *dtree.Node

	if info.ArchiveCount > 1 {
		mediaFile, _ = getRarForMediaInfo(info.Root)
	} else if len(info.Episodes) > 1 {
		mediaFile = info.Episodes[0].File
	} else if slices.Contains([]Section{AudioMP3, AudioFLAC, AudioBooks}, info.Section) {
		files := info.MediaFiles.GetByExtensions(AudioExtensions...)
		if len(files) > 0 {
			mediaFile = files[0]
		}
	} else {
		mediaFile = info.BiggestFile
	}

	if mediaFile == nil {
		s.log.Debug().Msg("no media file found for mediainfo generation")
		return
	}

	if mediaFile.Parent != nil && slices.Contains([]string{"STREAM", "VIDEO_TS"}, mediaFile.Parent.Info.Name) {
		if mediaFile.Parent.Parent != nil {
			s.log.Debug().Str("mediaFile", mediaFile.Parent.Parent.FullPath).
				Msg("using parent folder for mediainfo")
			mediaFile = mediaFile.Parent.Parent
		}
	}

	s.log.Debug().Str("mediaFile", mediaFile.FullPath).Msg("generating mediainfo...")

	mediaInfoJSON, mediaInfo, err := GenerateMediaInfo(mediaFile.FullPath)
	if err != nil {
		s.log.Error().Err(err).Str("mediaFile", mediaFile.FullPath).Msg("error generating mediainfo")
		return
	}

	info.MediaInfoJSON = mediaInfoJSON
	info.MediaInfo = mediaInfo
}

// tryExtractNFO extracts an NFO file from the mkv container if present and sets it in the provided Info context.
func (s *Service) tryExtractNFO(info *Info) {
	if info.MediaInfo == nil {
		return
	}

	if info.BiggestFile.Info.Extension != ".mkv" ||
		len(info.MediaInfo.GetAttachmentNames(".nfo")) == 0 {
		// no .mkv or no .nfo in container
		return
	}

	mkvNFO, err := ParseNfoAttachment(info.BiggestFile.FullPath)
	if err != nil {
		s.log.Error().Err(err).Str("mediaFile", info.BiggestFile.Info.Name).Msg("failed to parse nfo file")
		return
	}

	if len(mkvNFO.Content) > 0 {
		s.log.Debug().Str("nfoName", mkvNFO.Name).
			Str("mediaFile", info.BiggestFile.Info.Name).Msg("extracted nfo from mkv")
		info.NFO = &mkvNFO
	}
}

// checkForEmptySubfolders checks recursively for empty subfolders and logs or adds them to the forbidden files list.
func (s *Service) checkForEmptySubfolders(info *Info, node *dtree.Node) {
	for _, child := range node.Children {
		if child.Info.IsDir {
			if len(child.Children) == 0 {
				s.log.Error().Str("folder", child.FullPath).Err(ErrEmptyFolder).Msg("")
				info.ForbiddenFiles.addFile(child.FullPath, child.Info, ErrEmptyFolder)
			} else {
				s.checkForEmptySubfolders(info, child)
			}
		}
	}
}

// initReleaseInfo initializes a new Info struct from a file or directory path.
// It resolves the absolute path, extracts essential metadata like name, group, and year,
// and prepares the basic structure for further processing.
func (s *Service) initReleaseInfo(root string) (*Info, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("get absolute path: %w", err)
	}

	rootFileInfo, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("get file info: %w", err)
	}

	var (
		rlsName      = filepath.Base(absRoot)
		isSingleFile = !rootFileInfo.IsDir()
	)

	if isSingleFile {
		// remove extension from name
		rlsName = strings.TrimSuffix(rlsName, filepath.Ext(rlsName))
	}

	info := &Info{
		parents:       make(map[string]*dtree.Node),
		Extensions:    make(map[string]int),
		BaseDir:       absRoot,
		Name:          rlsName,
		Language:      ParseLanguage(rlsName),
		TagResolution: ParseResolution(rlsName),
		ProductTitle:  cleanTitle(rlsName),
		IsSingleFile:  isSingleFile,
	}

	if m := Regexes.Group.FindStringSubmatch(info.Name); m != nil {
		info.Group = m[1]
	}

	if m := Regexes.Year.FindAllStringSubmatch(info.Name, -1); m != nil {
		if len(m) > 1 {
			info.ProductYear, _ = strconv.Atoi(m[1][1])
		} else {
			info.ProductYear, _ = strconv.Atoi(m[0][1])
		}
	}

	return info, nil
}

// cleanTitle removes unnecessary metadata from the release name to extract a clean product title.
func cleanTitle(releaseName string) string {
	// Characters that should be replaced with spaces
	replacer := strings.NewReplacer(
		".", " ",
		"_", " ",
		"-", " ",
	)

	cleanedTitle := releaseName

	// Step 1: Remove last year
	if m := Regexes.Year.FindAllString(cleanedTitle, -1); m != nil {
		var idx int
		if len(m) > 1 {
			idx = strings.Index(cleanedTitle, m[1])
		} else {
			idx = strings.Index(cleanedTitle, m[0])
		}
		cleanedTitle = releaseName[0:idx]
	}

	// Step 2: Remove metadata using regex
	cleanedTitle = Regexes.CleanTitle.ReplaceAllString(cleanedTitle, "")

	// Step 3: Replace separators with spaces
	cleanedTitle = replacer.Replace(cleanedTitle)

	// Step 4: Trim trailing punctuation and spaces
	cleanedTitle = strings.TrimRightFunc(cleanedTitle, func(r rune) bool {
		return r == '.' || r == '_' || r == ' '
	})

	// Step 5: Remove empty segments and join with spaces
	titleParts := strings.Fields(cleanedTitle)

	return strings.Join(titleParts, " ")
}

// processPath processes a given file or directory path, handling errors, skips, forbidden criteria, and context updates.
func (s *Service) processPath(info *Info, path string, fileInfo *dtree.FileInfo, ignore []string) error {
	if len(ignore) > 0 {
		skip, err := s.checkIgnoreList(info, path, fileInfo, ignore)
		if err != nil {
			return fmt.Errorf("check ignore list: %w", err)
		}
		switch skip {
		case skipFile:
			return nil
		case skipDir:
			return fs.SkipDir
		default:
			// skipNothing
		}
	}

	if Regexes.BadChars.MatchString(fileInfo.Name) {
		info.ForbiddenFiles.addFile(path, fileInfo, ErrForbiddenCharacters)
		s.log.Error().Str("name", fileInfo.Name).Err(ErrForbiddenCharacters).Msg("")
	}

	node := &dtree.Node{
		FullPath: path,
		Info:     fileInfo,
	}

	info.parents[path] = node

	if !fileInfo.IsDir {
		info.Size += fileInfo.Size
		info.Extensions[fileInfo.Extension] += 1

		// empty file
		if fileInfo.Size == 0 {
			info.ForbiddenFiles.addFile(path, fileInfo, ErrEmptyFile)
			s.log.Error().Str("name", fileInfo.Name).Err(ErrEmptyFile).Msg("")
		}

		if err := s.checkFileExtension(info, node); err != nil {
			return fmt.Errorf("check file extension: %w", err)
		}

		if info.BiggestFile == nil || fileInfo.Size > info.BiggestFile.Info.Size {
			info.BiggestFile = node
		}
	}

	return nil
}

// checkIgnoreList evaluates if a file or directory should be skipped based on the provided ignore-patterns.
func (s *Service) checkIgnoreList(info *Info, path string, fileInfo *dtree.FileInfo, ignore []string) (skipType, error) {
	var (
		relPath string
		err     error
	)

	if info.IsSingleFile {
		relPath = filepath.Base(fileInfo.Name)
	} else {
		relPath, err = filepath.Rel(info.BaseDir, path)
		if err != nil {
			return skipNothing, fmt.Errorf("get relative path: %w", err)
		}
	}

	skip, err := canSkip(relPath, ignore, true)
	if err != nil {
		return skipNothing, fmt.Errorf("check ignore list: %w", err)
	}

	if !skip {
		return skipNothing, nil
	}

	if fileInfo.IsDir {
		s.log.Info().Str("folder", fileInfo.Name).Msg("ignoring directory")
		return skipDir, nil
	}

	s.log.Info().Str("file", fileInfo.Name).Msg("ignoring file")

	return skipFile, nil
}

// maxNFOSize is the maximum size of a nfo file that will be parsed.
const maxNFOSize int64 = 10 * 1024 * 1024 // 10MB

// checkFileExtension processes files based on their extension and updates the context.
// For .nfo files, only the first one is stored, but later ones are still checked for missing IMDB IDs.
func (s *Service) checkFileExtension(info *Info, node *dtree.Node) error {
	switch {
	case isForbidden(node.Info):
		info.ForbiddenFiles.addFile(node.FullPath, node.Info, ErrForbiddenExtension)
		s.log.Error().Str("name", node.Info.Name).Err(ErrForbiddenExtension).Msg("")

	case node.Info.Extension == ".sfv":
		info.SfvCount++

	case node.Info.Extension == ".nfo":
		if node.Info.Size == 0 || (info.ImdbID > 0 && info.NFO != nil) {
			break
		} else if node.Info.Size > maxNFOSize {
			s.log.Warn().Msg("nfo is bigger than 10MB, skip parsing")
			break
		}

		nfoContent, err := os.ReadFile(node.FullPath)
		if err != nil {
			return fmt.Errorf("read nfo file: %w", err)
		}

		if info.NFO == nil {
			info.NFO = &NFOFile{
				Name:    node.Info.Name,
				Content: nfoContent,
			}
		}

		if info.ImdbID == 0 {
			if m := Regexes.IMDb.FindSubmatch(nfoContent); m != nil {
				ttID := string(bytes.TrimLeft(m[1], "tt"))
				info.ImdbID, _ = strconv.Atoi(ttID)
			}
		}

	case Regexes.Archive.MatchString(node.Info.Extension):
		info.ArchiveCount++

	case Regexes.Media.MatchString(node.Info.Extension):
		info.MediaFiles = append(info.MediaFiles, node)
	}

	return nil
}

// canSkip is a helper function to check if the file or folder can be ignored.
func canSkip(path string, pattern []string, ignoreCase bool) (bool, error) {
	for _, p := range pattern {
		var (
			name  string
			match bool
			err   error
		)

		if strings.Contains(p, string(filepath.Separator)) {
			name = path
		} else {
			name = filepath.Base(path)
		}

		if ignoreCase {
			match, err = filepath.Match(strings.ToLower(p), strings.ToLower(name))
		} else {
			match, err = filepath.Match(p, name)
		}
		if err != nil {
			return false, fmt.Errorf("pattern %s: %w", p, err)
		}

		if match {
			return true, nil
		}
	}

	return false, nil
}

// HasMetaFiles checks extensions against Regexes.MetaFiles.
func (rel *Info) HasMetaFiles(ignore ...string) bool {
	for e := range rel.Extensions {
		if Regexes.MetaFiles.MatchString(e) && !slices.Contains(ignore, e) {
			return true
		}
	}

	return false
}

// HasAnyExtension checks if any of the given extensions are found.
func (rel *Info) HasAnyExtension(extensions ...string) bool {
	for _, ext := range extensions {
		if _, ok := rel.Extensions[ext]; ok {
			return true
		}
	}

	return false
}

// HasExtensions checks if all the given extensions are found.
func (rel *Info) HasExtensions(extensions ...string) bool {
	for _, ext := range extensions {
		if _, ok := rel.Extensions[ext]; !ok {
			return false
		}
	}

	return true
}

// HasAnyLanguage checks if any of the given languages are found.
func (rel *Info) HasAnyLanguage(languages ...string) bool {
	exists := slices.ContainsFunc(languages, func(s string) bool {
		return strings.EqualFold(s, rel.Language)
	})

	if exists {
		return true
	}

	if rel.MediaInfo != nil {
		if rel.MediaInfo.HasAnyLanguage(languages...) {
			return true
		}
	}

	return false
}

// HasGermanLanguage checks if release has german language.
func (rel *Info) HasGermanLanguage() bool {
	return rel.HasAnyLanguage("de", "german", "deutsch")
}

var (
	partRgx      = regexp.MustCompile(`(?i)\.part\d+\.rar`)
	firstPartRgx = regexp.MustCompile(`(?i)\.part0*1\.rar`)
	metaRarRgx   = regexp.MustCompile(`(?i)[._ -](subs?|proof)\.rar`)
)

// getRarForMediaInfo returns a fitting .rar file for media info generation.
func getRarForMediaInfo(startNode *dtree.Node) (*dtree.Node, error) {
	for _, node := range startNode.GetFiles(".rar") {
		if Regexes.MetaFolders.MatchString(node.Parent.Info.Name) {
			// ignore meta folders
			continue
		}

		if metaRarRgx.MatchString(node.Info.Name) {
			// ignore subtitle and proof files
			continue
		}

		if partRgx.MatchString(node.Info.Name) {
			if firstPartRgx.MatchString(node.Info.Name) {
				// use only the first part (e.g., part0*1.rar)
				return node, nil
			}
			continue // skip other parts
		}

		return node, nil
	}

	return nil, errors.New("no fitting rar file found")
}

// getEpisodes processes a list of media files and a node, extracting and sorting episodes by their numbers.
// If more than 1 episode has already been found in mediaFiles, the root node can be skipped, otherwise search in
// the subfolders (rootNode).
// Note: only call this function if the root node is a directory and not nil.
// Precondition: mediaFiles and rootNode must not be nil.
func getEpisodes(mediaFiles []*dtree.Node, rootNode *dtree.Node) []Episode {
	var episodes []Episode

	for _, nodes := range [][]*dtree.Node{mediaFiles, rootNode.Children} {
		for _, file := range nodes {
			if slices.Contains(PictureExtensions, file.Info.Extension) {
				continue
			}

			extractedEpisode := extractEpisodesFromFile(file)
			episodes = append(episodes, extractedEpisode...)
		}

		if len(episodes) > 1 {
			// we already found our episodes
			break
		}
	}

	// sort episodes by number
	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].Number < episodes[j].Number
	})

	return episodes
}

var (
	episodePattern      = regexp.MustCompile(`(?i)[ed](\d{1,3})`)
	episodeRangePattern = regexp.MustCompile(`(?i)[ed](\d{1,3})-[ed](\d{1,3})`)
)

// extractEpisodesFromFile parses a Node's file name to extract episode numbers and creates corresponding Episode objects.
func extractEpisodesFromFile(node *dtree.Node) []Episode {
	var (
		fileName   = node.Info.Name
		results    = make([]Episode, 0)
		episodeMap = make(map[int]struct{}) // To avoid duplicates
	)

	// Check for ranges first
	for _, match := range episodeRangePattern.FindAllStringSubmatch(fileName, -1) {
		start, err1 := strconv.Atoi(match[1])
		end, err2 := strconv.Atoi(match[2])

		if err1 == nil && err2 == nil && start <= end {
			for i := start; i <= end; i++ {
				episodeMap[i] = struct{}{}
			}
		}
	}

	// Check for individual episodes
	for _, match := range episodePattern.FindAllStringSubmatch(fileName, -1) {
		if episode, err := strconv.Atoi(match[1]); err == nil {
			episodeMap[episode] = struct{}{}
		}
	}

	for episode := range episodeMap {
		mediaFile := node.GetBiggest(nil)
		results = append(results, Episode{
			Number: episode,
			File:   mediaFile,
			Name:   mediaFile.Info.Name,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Number < results[j].Number
	})

	return results
}
