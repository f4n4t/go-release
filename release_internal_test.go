package release

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/f4n4t/go-dtree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		rlsName  string
		expected string
	}{
		{
			rlsName:  "Entity.1982.German.DL.AC3.1080p.BluRay.x265-RobertDeNiro",
			expected: "Entity",
		},
		{
			rlsName:  "Plattfuss.Ein.Cop.in.Neapel.S01.GERMAN.DL.1080p.WEB.H264-FWB",
			expected: "Plattfuss Ein Cop in Neapel",
		},
		{
			rlsName:  "Shining-Live_in_Transylvania-SE-16BIT-WEB-FLAC-2025-MOONBLOOD",
			expected: "Shining Live in Transylvania SE 16BIT",
		},
		{
			rlsName:  "PJ_Parker--Deep_House_Cat_Show_(Planet_Radio)-STREAM-04-20-2025-OMA_INT",
			expected: "PJ Parker Deep House Cat Show (Planet Radio)",
		},
	}

	for _, tt := range tests {
		got := cleanTitle(tt.rlsName)
		assert.Equal(t, tt.expected, got)
	}
}

func TestCheckIgnoreList(t *testing.T) {
	tests := []struct {
		desc         string
		info         *Info
		path         string
		fileInfo     *dtree.FileInfo
		ignore       []string
		expectedSkip skipType
		shouldError  bool
	}{
		{
			desc: "ignore nothing on non-matching file",
			info: &Info{Root: &dtree.Node{Info: &dtree.FileInfo{IsDir: false}}, BaseDir: "/base"},
			path: "/base/file.txt",
			fileInfo: &dtree.FileInfo{
				Name: "file.txt",
			},
			ignore:       []string{"*.jpg"},
			expectedSkip: skipNothing,
		},
		{
			desc: "ignore file matching pattern",
			info: &Info{Root: &dtree.Node{Info: &dtree.FileInfo{IsDir: false}}, BaseDir: "/base"},
			path: "/base/file.jpg",
			fileInfo: &dtree.FileInfo{
				Name: "file.jpg",
			},
			ignore:       []string{"*.jpg"},
			expectedSkip: skipFile,
		},
		{
			desc: "ignore directory matching pattern",
			info: &Info{Root: &dtree.Node{Info: &dtree.FileInfo{IsDir: false}}, BaseDir: "/base"},
			path: "/base/images",
			fileInfo: &dtree.FileInfo{
				Name:  "images",
				IsDir: true,
			},
			ignore:       []string{"images"},
			expectedSkip: skipDir,
		},
		{
			desc: "non-matching directory",
			info: &Info{Root: &dtree.Node{Info: &dtree.FileInfo{IsDir: false}}, BaseDir: "/base"},
			path: "/base/other",
			fileInfo: &dtree.FileInfo{
				Name:  "other",
				IsDir: true,
			},
			ignore:       []string{"images"},
			expectedSkip: skipNothing,
		},
		{
			desc: "error on invalid relative path",
			info: &Info{Root: &dtree.Node{Info: &dtree.FileInfo{IsDir: true}}, BaseDir: "asd/invalidBase"},
			path: "/base/file.txt",
			fileInfo: &dtree.FileInfo{
				Name: "file.txt",
			},
			ignore:       []string{"*.txt"},
			expectedSkip: skipNothing,
			shouldError:  true,
		},
		{
			desc: "ignore file with multiple patterns",
			info: &Info{Root: &dtree.Node{Info: &dtree.FileInfo{IsDir: false}}, BaseDir: "/base"},
			path: "/base/image.jpg",
			fileInfo: &dtree.FileInfo{
				Name: "image.jpg",
			},
			ignore:       []string{"*.png", "*.jpg"},
			expectedSkip: skipFile,
		},
		{
			desc: "ignore directory relative pattern",
			info: &Info{Root: &dtree.Node{Info: &dtree.FileInfo{IsDir: true}}, BaseDir: "/base"},
			path: "/base/images",
			fileInfo: &dtree.FileInfo{
				Name:  "images",
				IsDir: true,
			},
			ignore:       []string{"images"},
			expectedSkip: skipDir,
			shouldError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			service := &Service{}
			result, err := service.checkIgnoreList(tt.info, tt.path, tt.fileInfo, tt.ignore)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSkip, result)
			}
		})
	}
}

func TestService_checkFileExtension(t *testing.T) {
	type testFile struct {
		name    string
		content []byte
	}

	tests := []struct {
		desc         string
		testFile     testFile
		expectedInfo *Info
		expectedErr  error
	}{
		{
			desc:     "forbidden extension",
			testFile: testFile{"release/test.srr", []byte("blub")},
			expectedInfo: &Info{
				ForbiddenFiles: []ForbiddenFile{
					{
						FullPath: "release/test.srr",
						Error:    ErrForbiddenExtension,
					},
				},
			},
		},
		{
			desc:     "sfv file",
			testFile: testFile{"release/test.sfv", []byte("blub")},
			expectedInfo: &Info{
				SfvCount: 1,
			},
		},
		{
			desc:     "archive",
			testFile: testFile{"release/test.rar", []byte("blub")},
			expectedInfo: &Info{
				ArchiveCount: 1,
			},
		},
		{
			desc:     "media file",
			testFile: testFile{"release/test.mkv", []byte("blub")},
			expectedInfo: &Info{
				MediaFiles: []*dtree.Node{
					{
						FullPath: "release/test.mkv",
					},
				},
			},
		},
		{
			desc:     "nfo file without imdbid",
			testFile: testFile{"release/test.nfo", []byte("blub")},
			expectedInfo: &Info{
				NFO: &NFOFile{
					Name:    "test.nfo",
					Content: []byte("blub"),
				},
			},
		},
		{
			desc:     "nfo file with imdbid",
			testFile: testFile{"release/test.nfo", []byte("asd\nhttps://www.imdb.com/title/tt1702924/\nasd")},
			expectedInfo: &Info{
				ImdbID: 1702924,
				NFO: &NFOFile{
					Name:    "test.nfo",
					Content: []byte("asd\nhttps://www.imdb.com/title/tt1702924/\nasd"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			testService := &Service{}

			// setup test file
			tempDir := t.TempDir()
			testFiles := map[string][]byte{tt.testFile.name: tt.testFile.content}

			setupTestDir(t, tempDir, testFiles)

			path := filepath.Join(tempDir, tt.testFile.name)

			node := &dtree.Node{
				FullPath: path,
				Info: &dtree.FileInfo{
					Name:      filepath.Base(tt.testFile.name),
					Size:      int64(len(tt.testFile.content)),
					IsDir:     false,
					ModTime:   time.Now(),
					Extension: filepath.Ext(tt.testFile.name),
				},
			}

			info := &Info{}

			if len(tt.expectedInfo.ForbiddenFiles) > 0 {
				tt.expectedInfo.ForbiddenFiles[0].FullPath = path
				tt.expectedInfo.ForbiddenFiles[0].Info = node.Info
			}

			if len(tt.expectedInfo.MediaFiles) > 0 {
				tt.expectedInfo.MediaFiles[0].FullPath = path
				tt.expectedInfo.MediaFiles[0].Info = node.Info
			}

			gotErr := testService.checkFileExtension(info, node)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, gotErr, tt.expectedErr)
				return
			}
			assert.Equal(t, tt.expectedInfo, info)
		})
	}
}

func TestCanSkip(t *testing.T) {
	tests := []struct {
		desc        string
		path        string
		pattern     []string
		ignoreCase  bool
		expected    bool
		expectedErr error
	}{
		{
			desc:       "skip sample (ignore case)",
			path:       "/release-test/Sample",
			pattern:    []string{"sample"},
			ignoreCase: true,
			expected:   true,
		},
		{
			desc:       "skip sample, case sensitive",
			path:       "/release-test/sample",
			pattern:    []string{"Sample"},
			ignoreCase: false,
			expected:   false,
		},
		{
			desc:       "skip sample (pattern)",
			path:       "/release-test/Sample",
			pattern:    []string{"[sS]ample"},
			ignoreCase: false,
			expected:   true,
		},
		{
			desc:       "skip test.par2",
			path:       "/release-test/test.PAR2",
			pattern:    []string{"test.par2"},
			ignoreCase: true,
			expected:   true,
		},
		{
			desc:        "bad pattern",
			path:        "/release-test/Sample",
			pattern:     []string{"[sSample"},
			expectedErr: filepath.ErrBadPattern,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := canSkip(tt.path, tt.pattern, tt.ignoreCase)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, gotErr, tt.expectedErr)
				return
			}
			require.NoError(t, gotErr)

			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestExtractEpisodesFromFile(t *testing.T) {
	tests := []struct {
		desc         string
		inputFile    *dtree.Node
		wantEpisodes []Episode
	}{
		{
			desc:         "no regex match, without children",
			inputFile:    createFileNode("/Release.S01.German/Release.S01B01.German", true, 0),
			wantEpisodes: []Episode{},
		},
		{
			desc: "no regex match, with children",
			inputFile: func() *dtree.Node {
				parent := createFileNode("/Release.S01.German/Release.S01B01.German", true, 0)
				child := createFileNode("/Release.S01.German/Release.S01B01.German/test.mkv", false, 4)
				parent.Children = []*dtree.Node{child}
				return parent
			}(),
			wantEpisodes: []Episode{},
		},
		{
			desc:      "one episode in main folder",
			inputFile: createFileNode("/Release.S01E01.German.mkv", false, 4),
			wantEpisodes: []Episode{
				createEpisode(1, createFileNode("/Release.S01E01.German.mkv", false, 4)),
			},
		},
		{
			desc:      "multiple episodes in single file",
			inputFile: createFileNode("/Release.S01E01E02E03.German.mkv", false, 4),
			wantEpisodes: func() []Episode {
				file := createFileNode("/Release.S01E01E02E03.German.mkv", false, 4)
				return []Episode{
					createEpisode(1, file),
					createEpisode(2, file),
					createEpisode(3, file),
				}
			}(),
		},
		{
			desc: "multiple episodes in sub directory",
			inputFile: func() *dtree.Node {
				parent := createFileNode("/Release.S01.German/Release.S01E01E02E03.German", true, 0)
				child := createFileNode("/Release.S01.German/Release.S01E01E02E03.German/test.mkv", false, 4)
				parent.Children = []*dtree.Node{child}
				return parent
			}(),
			wantEpisodes: func() []Episode {
				file := createFileNode("/Release.S01.German/Release.S01E01E02E03.German/test.mkv", false, 4)
				return []Episode{
					createEpisode(1, file),
					createEpisode(2, file),
					createEpisode(3, file),
				}
			}(),
		},
		{
			desc: "single episode, get biggest file",
			inputFile: func() *dtree.Node {
				parent := createFileNode("/Release.S01E01.German", true, 0)
				child1 := createFileNode("/Release.S01E01.German/test.mkv", false, 4)
				child2 := createFileNode("/Release.S01E01.German/sample.mkv", false, 2)
				child3 := createFileNode("/Release.S01E01.German/test.nfo", false, 1)
				parent.Children = []*dtree.Node{child1, child2, child3}
				return parent
			}(),
			wantEpisodes: func() []Episode {
				file := createFileNode("/Release.S01E01.German/test.mkv", false, 4)
				return []Episode{
					createEpisode(1, file),
				}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := extractEpisodesFromFile(tt.inputFile)
			assert.Equal(t, tt.wantEpisodes, got)
		})
	}
}

func createFileNode(path string, isDir bool, size int64) *dtree.Node {
	name := filepath.Base(path)
	ext := ""
	if !isDir {
		ext = filepath.Ext(path)
	}
	return &dtree.Node{
		FullPath: path,
		Info: &dtree.FileInfo{
			Name:      name,
			IsDir:     isDir,
			Size:      size,
			Extension: ext,
		},
	}
}

func createEpisode(number int, file *dtree.Node) Episode {
	return Episode{
		Number: number,
		Name:   file.Info.Name,
		File:   file,
	}
}

func TestGetEpisodes(t *testing.T) {
	type input struct {
		mediaFiles []*dtree.Node
		rootNode   *dtree.Node
	}

	tests := []struct {
		desc         string
		input        input
		rootNode     *dtree.Node
		wantEpisodes []Episode
	}{
		{
			desc: "multiple episodes in single file",
			input: func() input {
				rootNode := createFileNode("/Release.S01.German", true, 0)
				file1 := createFileNode("/Release.S01.German/s01e01e02e03.mkv", false, 4)
				rootNode.Children = []*dtree.Node{file1}
				return input{rootNode: rootNode, mediaFiles: []*dtree.Node{file1}}
			}(),
			wantEpisodes: func() []Episode {
				file1 := createFileNode("/Release.S01.German/s01e01e02e03.mkv", false, 4)
				return []Episode{
					createEpisode(1, file1),
					createEpisode(2, file1),
					createEpisode(3, file1),
				}
			}(),
		},
		{
			desc: "multiple episodes, without sub directory",
			input: func() input {
				parent := createFileNode("/Release.S01.German", true, 0)
				child1 := createFileNode("/Release.S01.German/s01e01rp.mkv", false, 4)
				child2 := createFileNode("/Release.S01.German/s01e02.mkv", false, 3)
				child3 := createFileNode("/Release.S01.German/s01e03rp.mkv", false, 2)
				parent.Children = []*dtree.Node{child1, child2, child3}
				return input{rootNode: parent, mediaFiles: []*dtree.Node{child1, child2, child3}}
			}(),
			wantEpisodes: func() []Episode {
				file1 := createFileNode("/Release.S01.German/s01e01rp.mkv", false, 4)
				file2 := createFileNode("/Release.S01.German/s01e02.mkv", false, 3)
				file3 := createFileNode("/Release.S01.German/s01e03rp.mkv", false, 2)
				return []Episode{
					createEpisode(1, file1),
					createEpisode(2, file2),
					createEpisode(3, file3),
				}
			}(),
		},
		{
			desc: "multiple episodes, with sub directories",
			input: func() input {
				parent := createFileNode("/Release.S01.German", true, 0)
				childDir1 := createFileNode("/Release.S01.German/Release.S01E01E02.German", true, 0)
				childDir2 := createFileNode("/Release.S01.German/Release.S01E03.German", true, 0)
				child1 := createFileNode("/Release.S01.German/Release.S01E01E02.German/episode.mkv", false, 4)
				child2 := createFileNode("/Release.S01.German/Release.S01E03.German/episode.mkv", false, 3)
				childDir1.Children = []*dtree.Node{child1}
				childDir2.Children = []*dtree.Node{child2}
				parent.Children = []*dtree.Node{childDir1, childDir2}
				return input{rootNode: parent, mediaFiles: []*dtree.Node{child1, child2}}
			}(),
			wantEpisodes: func() []Episode {
				file1 := createFileNode("/Release.S01.German/Release.S01E01E02.German/episode.mkv", false, 4)
				file2 := createFileNode("/Release.S01.German/Release.S01E03.German/episode.mkv", false, 3)
				return []Episode{
					createEpisode(1, file1),
					createEpisode(2, file1),
					createEpisode(3, file2),
				}
			}(),
		},
		{
			desc: "multiple episodes, mixed with multi episode files",
			input: func() input {
				parent := createFileNode("/Release.S01.German", true, 0)
				child1 := createFileNode("/Release.S01.German/s01e01.mkv", false, 4)
				child2 := createFileNode("/Release.S01.German/s01e02.mkv", false, 3)
				child3 := createFileNode("/Release.S01.German/s01e03.e04.e05.mkv", false, 2)
				child4 := createFileNode("/Release.S01.German/s01e06.mkv", false, 1)
				parent.Children = []*dtree.Node{child1, child2, child3, child4}
				return input{rootNode: parent, mediaFiles: []*dtree.Node{child1, child2, child3, child4}}
			}(),
			wantEpisodes: func() []Episode {
				file1 := createFileNode("/Release.S01.German/s01e01.mkv", false, 4)
				file2 := createFileNode("/Release.S01.German/s01e02.mkv", false, 3)
				file3 := createFileNode("/Release.S01.German/s01e03.e04.e05.mkv", false, 2)
				file4 := createFileNode("/Release.S01.German/s01e06.mkv", false, 1)
				return []Episode{
					createEpisode(1, file1),
					createEpisode(2, file2),
					createEpisode(3, file3),
					createEpisode(4, file3),
					createEpisode(5, file3),
					createEpisode(6, file4),
				}
			}(),
		},
		{
			desc: "anime style without season",
			input: func() input {
				parent := createFileNode("/Release.German", true, 0)
				child1 := createFileNode("/Release.German/e001.mkv", false, 4)
				child2 := createFileNode("/Release.German/e002.mkv", false, 4)
				child3 := createFileNode("/Release.German/e003.mkv", false, 4)
				child4 := createFileNode("/Release.German/e004.mkv", false, 4)
				parent.Children = []*dtree.Node{child1, child2, child3, child4}
				return input{rootNode: parent, mediaFiles: []*dtree.Node{child1, child2, child3, child4}}
			}(),
			wantEpisodes: func() []Episode {
				file1 := createFileNode("/Release.German/e001.mkv", false, 4)
				file2 := createFileNode("/Release.German/e002.mkv", false, 4)
				file3 := createFileNode("/Release.German/e003.mkv", false, 4)
				file4 := createFileNode("/Release.German/e004.mkv", false, 4)
				return []Episode{
					createEpisode(1, file1),
					createEpisode(2, file2),
					createEpisode(3, file3),
					createEpisode(4, file4),
				}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := getEpisodes(tt.input.mediaFiles, tt.input.rootNode)
			assert.Equal(t, tt.wantEpisodes, got)
		})
	}
}

func TestDetectSectionByExtensions(t *testing.T) {
	t.Run("CheckIfIsIgnored", func(t *testing.T) {
		extensions := map[string]int{
			".mp3": 1,
		}
		testInfo := Info{
			Section:    TV,
			Extensions: extensions,
		}
		testInfo.checkForSectionByExtensions()
		assert.Equal(t, TV, testInfo.Section)
	})

	t.Run("CheckIfIsIgnored", func(t *testing.T) {
		extensions := map[string]int{
			".mp3": 1,
			".mkv": 1,
		}
		testInfo := Info{
			Section:    Unknown,
			Extensions: extensions,
		}
		testInfo.checkForSectionByExtensions()
		assert.Equal(t, Unknown, testInfo.Section)
	})

	t.Run("TestForMP3", func(t *testing.T) {
		extensions := map[string]int{
			".mp3": 1,
		}
		testInfo := Info{
			Section:    Unknown,
			Extensions: extensions,
		}
		testInfo.checkForSectionByExtensions()
		assert.Equal(t, AudioMP3, testInfo.Section)
	})

	t.Run("TestForFLAC", func(t *testing.T) {
		extensions := map[string]int{
			".flac": 1,
		}
		testInfo := Info{
			Section:    Unknown,
			Extensions: extensions,
		}
		testInfo.checkForSectionByExtensions()
		assert.Equal(t, AudioFLAC, testInfo.Section)
	})

	t.Run("TestForEbook", func(t *testing.T) {
		extensions := map[string]int{
			".pdf": 1,
		}
		testInfo := Info{
			Section:    Unknown,
			Extensions: extensions,
		}
		testInfo.checkForSectionByExtensions()
		assert.Equal(t, Ebooks, testInfo.Section)
	})
}
