package srrdb

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type StoredFile struct {
	Path string
	Data []byte
}

type OSOHash struct {
	Path string
	Size uint64
	Hash uint64
}

type RarFile struct {
	Path     string
	Size     int
	CRC      uint32
	IsFirst  bool
	IsNewFmt bool
}

type PackedFile struct {
	Path string
	Size uint64
	CRC  uint32
}

type SrrFile struct {
	ApplicationName string
	StoredFiles     []*StoredFile
	OSOHashes       []*OSOHash
	RarFiles        []*RarFile
	RarCompressed   bool
	PackedFiles     []*PackedFile
	SFVComments     []string
}

func (f *SrrFile) Unmarshal(b []byte) (err error) {
	f.StoredFiles = make([]*StoredFile, 0)
	f.OSOHashes = make([]*OSOHash, 0)
	f.RarFiles = make([]*RarFile, 0)
	f.RarCompressed = false
	f.PackedFiles = make([]*PackedFile, 0)
	f.SFVComments = make([]string, 0)
	prevHeader := &RarHeader{}
	currentRarFile := &RarFile{}
	currentPackedFile := &PackedFile{}
	offset := 0
	for offset < len(b) {
		m := make(map[string]any)
		m["srr_header_offset"] = offset

		header := &RarHeader{}
		err := header.Parse(b[offset : offset+7])
		if err != nil {
			return err
		}
		m["srr_header_type"] = fmt.Sprintf("%x", header.Type)
		m["srr_header_flags"] = fmt.Sprintf("%x", header.Flags)
		m["srr_header_crc"] = fmt.Sprintf("%x", header.CRC)
		m["srr_header_size"] = header.Size

		switch header.Type {
		case SrrVolHead: // 0x69
			block := &SrrVolHeadBlock{
				RarHeader: *header,
			}
			if err = block.Parse(b[offset:]); err != nil {
				return err
			}
			f.ApplicationName = block.GetAppName()
			offset += int(block.GetSize())
			prevHeader = header
		case SrrStoredFileHead: // 0x6A
			block := &SrrStoredFileHeadBlock{
				RarHeader: *header,
			}
			if err = block.Parse(b[offset:]); err != nil {
				return err
			}
			s := &StoredFile{}
			if s, err = block.GetStoredFile(); err != nil {
				return err
			} else {
				f.StoredFiles = append(f.StoredFiles, s)
			}
			offset += int(block.GetSize())
			prevHeader = header
		case OSOHashHead: // 0x6B
			block := &OSOHashHeadBlock{
				RarHeader: *header,
			}
			if err = block.Parse(b[offset:]); err != nil {
				return err
			}
			h := &OSOHash{}
			if h, err = block.GetOSOHash(); err != nil {
				return err
			}
			f.OSOHashes = append(f.OSOHashes, h)
			offset += int(block.GetSize())
			prevHeader = header
		case SrrRarPadHead: // 0x6C
			block := &SrrRarPadHeadBlock{
				RarHeader: *header,
			}
			if err = block.Parse(b[offset:]); err != nil {
				return err
			}
			offset += int(block.GetSize())
			currentRarFile.Size += int(block.PadSize)
		case SrrRarSubBlockHead: // 0x71
			block := &SrrRarSubBlockHeadBlock{
				RarHeader: *header,
			}
			if err = block.Parse(b[offset:]); err != nil {
				return err
			}
			currentRarFile = &RarFile{
				Size: 0,
				Path: block.GetRarFileName(),
			}
			f.RarFiles = append(f.RarFiles, currentRarFile)
			offset += int(block.GetSize())
			prevHeader = header
		case MarkHead: // 0x72
			block := &MarkHeadBlock{
				RarHeader: *header,
			}
			if err = block.Parse(b[offset:]); err != nil {
				return err
			}
			if prevHeader.Type != SrrRarSubBlockHead {
				return ErrBadFile
			}
			currentRarFile.Size += block.GetSize()
			offset += int(header.Size)
			prevHeader = header
		case MainHead: // 0x73
			//nothing to do
			currentRarFile.IsFirst = header.Flag(MHD_FIRSTVOLUME)
			currentRarFile.IsNewFmt = header.Flag(MHD_NEWNUMBERING)
			currentRarFile.Size += int(header.Size)
			offset += int(header.Size)
			prevHeader = header
		case FileHead: // 0x74
			block := &FileHeadBlock{
				RarHeader: *header,
			}
			if err = block.Parse(b[offset:]); err != nil {
				return err
			}
			if !block.Flag(LHD_SPLIT_BEFORE) || (currentPackedFile.Path == "" && block.GetFileName() != "") {
				currentPackedFile = &PackedFile{
					Path: block.GetFileName(),
					CRC:  block.GetCRC(),
				}
			}
			if err = block.UpdatePackedFile(currentPackedFile); err != nil {
				return err
			}
			newFile := true
			for i := range f.PackedFiles {
				if f.PackedFiles[i].Path == currentPackedFile.Path && f.PackedFiles[i].CRC == currentPackedFile.CRC {
					newFile = false
					break
				}
			}
			if newFile {
				f.PackedFiles = append(f.PackedFiles, currentPackedFile)
			}

			if !f.RarCompressed && block.IsCompressed() {
				f.RarCompressed = true
			}
			currentRarFile.Size += block.GetSize()
			offset += int(header.Size)
			prevHeader = header
		case CommHead: // 0x75
			// nothing to do
			currentRarFile.Size += int(header.Size)
			offset += int(header.Size)
			prevHeader = header
		case AvHead: // 0x76
			// nothing to do
			currentRarFile.Size += int(header.Size)
			offset += int(header.Size)
			prevHeader = header
		case SubHead: // 0x77
			// nothing to do
			currentRarFile.Size += int(header.Size)
			offset += int(header.Size)
			prevHeader = header
		case ProtectHead: // 0x78
			block := &ProtectHeadBlock{
				RarHeader: *header,
			}
			if err = block.Parse(b[offset:]); err != nil {
				return err
			}
			currentRarFile.Size += block.GetSize()
			offset += int(header.Size)
			prevHeader = header
		case SignHead: // 0x79
			// nothing to do
			currentRarFile.Size += int(header.Size)
			offset += int(header.Size)
			prevHeader = header
		case NewSubHead: // 0x7A
			block := &NewSubHeadBlock{
				RarHeader: *header,
			}
			if err = block.Parse(b[offset:]); err != nil {
				return err
			}
			currentRarFile.Size += block.GetSize()
			if block.GetFileName() == "RR" {
				// stripped data
				offset += int(header.Size)
			} else {
				offset += int(header.Size) + block.GetPackSize()
			}
			prevHeader = header
		case EndArcHead: // 0x7B
			// Terminator, nothing to do
			currentRarFile.Size += int(header.Size)
			offset += int(header.Size)
			prevHeader = header
		case EmptyHead: // 0x00
			// for P0W4 releases, cleared header, add to rarfile size
			currentRarFile.Size += int(header.Size)
			offset += int(header.Size)
			prevHeader = header
		default:
			return nil
		}
	}

	sfv := make(map[string]uint32)
	reSFVComment := regexp.MustCompile(`^\s*(?P<Comment>;.*)$`)
	reSFV := regexp.MustCompile(`\s*(?P<Filename>[!-~ ]*)[ ]+(?P<CRC>[a-fA-F0-9]{1,8})\s*$`)

	for _, v := range f.StoredFiles {
		if strings.ToLower(filepath.Ext(v.Path)) == ".sfv" {
			sfvPath := strings.ToLower(filepath.Dir(v.Path))
			if sfvPath == "." {
				sfvPath = ""
			} else {
				sfvPath = sfvPath + "/"
			}
			reLine := regexp.MustCompile(`[^\r\n]+([\r\n]+|$)`)
			for _, line := range reLine.FindAllString(string(v.Data), -1) {
				l := strings.Trim(line, "\r\n")
				if len(line) < 10 {
					if l != "" {
						f.SFVComments = append(f.SFVComments, l)
						// this is a malformed line, considered as a comment
					}
					if err != nil {
						break
					}
					continue
				}
				if reSFVComment.MatchString(l) {
					f.SFVComments = append(f.SFVComments, l)
					// this is a comment
					if err != nil {
						break
					}
					continue
				}
				if !reSFV.MatchString(l) {
					// this is not a proper line
					if err != nil {
						break
					}
					continue
				}

				file := sfvPath + strings.ToLower(reSFV.ReplaceAllString(l, "${Filename}"))

				crc := reSFV.ReplaceAllString(l, "${CRC}")
				crc64, err := strconv.ParseUint(crc, 16, 64)

				if err != nil {
					return err
				}

				if crc32, ok := sfv[file]; ok {
					if crc32 != uint32(crc64) {
						return ErrDuplicateSFV
					}
				} else {
					sfv[file] = uint32(crc64)
				}
				if err != nil {
					break
				}
			}
		}
	}

	for filename, crc := range sfv {
		for id, rar := range f.RarFiles {
			rarPath := strings.ToLower(rar.Path)
			if rarPath == filename {
				f.RarFiles[id].CRC = crc
				break
			} else if (filepath.Base(rarPath) == filepath.Base(filename)) && (f.RarFiles[id].CRC == 0) {
				f.RarFiles[id].CRC = crc
			}
		}
	}

	sort.Slice(f.RarFiles, func(i, j int) bool {
		a := strings.ToLower(f.RarFiles[i].Path)
		b := strings.ToLower(f.RarFiles[j].Path)
		if a == b {
			return f.RarFiles[i].Path < f.RarFiles[j].Path
		} else {
			return a < b
		}
	})

	return nil
}

func RarRootName(path string) string {
	var re *regexp.Regexp
	reOld := regexp.MustCompile(`^(?P<Base>.*)(\\.[rstu]{1}[0-9]{2})$`)
	re = reOld
	if re.MatchString(path) {
		return strings.ToLower(re.ReplaceAllString(path, "${Base}"))
	}
	reOld2 := regexp.MustCompile(`^(?P<Base>.*)(\\.[0-9]{3})$`)
	re = reOld2
	if re.MatchString(path) {
		return strings.ToLower(re.ReplaceAllString(path, "${Base}"))
	}
	reNew := regexp.MustCompile(`^(?P<Base>.*)(\\.part[0-9]+\\.rar)$`)
	re = reNew
	if re.MatchString(path) {
		return strings.ToLower(re.ReplaceAllString(path, "${Base}"))
	}
	reOld3 := regexp.MustCompile(`^(?P<Base>.*)([0-9]{2,3}.rar)$`)
	re = reOld3
	if re.MatchString(path) {
		return strings.ToLower(re.ReplaceAllString(path, "${Base}"))
	}
	reOldInit := regexp.MustCompile(`^(?P<Base>.*)(\\.rar)$`)
	re = reOldInit
	if re.MatchString(path) {
		return strings.ToLower(re.ReplaceAllString(path, "${Base}"))
	}
	return ""
}
