package srrdb

import (
	"bytes"
	"encoding/binary"
	"strings"
)

type RarHeaderType byte

type RarHeaderFlag uint16

type RarNameFmt byte

const (
	EmptyHead          RarHeaderType = 0x00
	SrrVolHead         RarHeaderType = 0x69
	SrrStoredFileHead  RarHeaderType = 0x6A
	OSOHashHead        RarHeaderType = 0x6B
	SrrRarPadHead      RarHeaderType = 0x6C
	SrrRarSubBlockHead RarHeaderType = 0x71
	MarkHead           RarHeaderType = 0x72
	MainHead           RarHeaderType = 0x73
	FileHead           RarHeaderType = 0x74
	CommHead           RarHeaderType = 0x75
	AvHead             RarHeaderType = 0x76
	SubHead            RarHeaderType = 0x77
	ProtectHead        RarHeaderType = 0x78
	SignHead           RarHeaderType = 0x79
	NewSubHead         RarHeaderType = 0x7A
	EndArcHead         RarHeaderType = 0x7B
	MHD_VOLUME         RarHeaderFlag = 0x0001
	MHD_COMMENT        RarHeaderFlag = 0x0002
	MHD_LOCK           RarHeaderFlag = 0x0004
	MHD_SOLID          RarHeaderFlag = 0x0008
	MHD_NEWNUMBERING   RarHeaderFlag = 0x0010
	MHD_AV             RarHeaderFlag = 0x0020
	MHD_PROTECT        RarHeaderFlag = 0x0040
	MHD_PASSWORD       RarHeaderFlag = 0x0080
	MHD_FIRSTVOLUME    RarHeaderFlag = 0x0100
	MHD_ENCRYPTVER     RarHeaderFlag = 0x0200
	LHD_SPLIT_BEFORE   RarHeaderFlag = 0x0001
	LHD_SPLIT_AFTER    RarHeaderFlag = 0x0002
	LHD_PASSWORD       RarHeaderFlag = 0x0004
	LHD_COMMENT        RarHeaderFlag = 0x0008
	LHD_SOLID          RarHeaderFlag = 0x0010
	LHD_LARGE          RarHeaderFlag = 0x0100
	LHD_UNICODE        RarHeaderFlag = 0x0200
	LHD_SALT           RarHeaderFlag = 0x0400
	LHD_VERSION        RarHeaderFlag = 0x0800
	LHD_EXTTIME        RarHeaderFlag = 0x1000
	LHD_EXTFLAGS       RarHeaderFlag = 0x2000
	HAS_DATA           RarHeaderFlag = 0x8000
	SRR_APP_NAME       RarHeaderFlag = 0x0001
)

type RarHeader struct {
	CRC   uint16
	Type  RarHeaderType
	Flags RarHeaderFlag
	Size  uint16
}

type SrrVolHeadBlock struct {
	RarHeader
	AppNameLength uint16
	AppName       []byte
}

type SrrStoredFileHeadBlock struct {
	RarHeader
	DataSize uint32
	NameSize uint16
	FileName []byte
	FileData []byte
}

type OSOHashHeadBlock struct {
	RarHeader
	FileSize uint64
	OSOHash  uint64
	NameSize uint16
	FileName []byte
}

type SrrRarPadHeadBlock struct {
	RarHeader
	PadSize uint32
	PadData []byte
}

type SrrRarSubBlockHeadBlock struct {
	RarHeader
	NameSize uint16
	FileName []byte
}

type MarkHeadBlock struct {
	RarHeader
}

type MainHeadBlock struct {
	RarHeader
}

type FileHeadBlock struct {
	RarHeader
	LowPackSize     uint32
	LowUnpackSize   uint32
	HostOS          uint8
	FileCRC         uint32
	FileTime        uint32
	UnpackVersion   uint8
	Method          uint8
	NameSize        uint16
	FileAttr        uint32
	HighPackSize    uint32
	HighUnpackSize  uint32
	FileName        []byte
	FileNameUnicode []byte
	Salt            uint64
}

type CommHeadBlock struct {
	RarHeader
}

type AvHeadBlock struct {
	RarHeader
}

type SubHeadBlock struct {
	RarHeader
}

type ProtectHeadBlock struct {
	RarHeader
	PackedSize      uint32
	Version         uint8
	RecSectorCount  uint16
	DataSectorCount uint32
}

type SignHeadBlock struct {
	RarHeader
}

type NewSubHeadBlock struct {
	RarHeader
	LowPackSize    uint32
	LowUnpackSize  uint32
	HostOS         uint8
	FileCRC        uint32
	FileTime       uint32
	UnpackVersion  uint8
	Method         uint8
	NameSize       uint16
	FileAttr       uint32
	HighPackSize   uint32
	HighUnpackSize uint32
	FileName       []byte
	Salt           uint64
}

type EndArcHeadBlock struct {
	RarHeader
}

func (h *RarHeader) Flag(f RarHeaderFlag) bool {
	return (h.Flags & f) == f
}

func (h *RarHeader) Parse(b []byte) error {
	buffer := bytes.NewBuffer(b[0:7])
	return binary.Read(buffer, binary.LittleEndian, h)
}

func (h *RarHeader) GetSize() int {
	return int(h.Size)
}

func (b *SrrVolHeadBlock) Parse(data []byte) (err error) {
	if b.CRC != 0x6969 {
		return ErrCRCMismatch
	}
	buffer := bytes.NewBuffer(data[7:])
	if !b.RarHeader.Flag(SRR_APP_NAME) {
		b.AppNameLength = 0
		return nil
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.AppNameLength)
	if err != nil {
		return err
	}
	b.AppName = make([]byte, b.AppNameLength)
	err = binary.Read(buffer, binary.LittleEndian, &b.AppName)
	if err != nil {
		return err
	}
	return nil
}

func (b *SrrVolHeadBlock) GetAppName() string {
	if b.RarHeader.Flag(SRR_APP_NAME) {
		return string(b.AppName)
	} else {
		return ""
	}
}

func (b *SrrStoredFileHeadBlock) Parse(data []byte) error {
	if b.CRC != 0x6A6A {
		return ErrCRCMismatch
	}
	if !b.Flag(HAS_DATA) {
		return ErrBadBlock
	}
	buffer := bytes.NewBuffer(data[7:])
	err := binary.Read(buffer, binary.LittleEndian, &b.DataSize)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.NameSize)
	if err != nil {
		return err
	}
	b.FileName = make([]byte, b.NameSize)
	b.FileData = make([]byte, b.DataSize)
	err = binary.Read(buffer, binary.LittleEndian, &b.FileName)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.FileData)
	if err != nil {
		return err
	}
	return nil
}

func (b *SrrStoredFileHeadBlock) GetSize() int {
	return int(b.RarHeader.Size) + int(b.DataSize)
}

func (b *SrrStoredFileHeadBlock) GetStoredFile() (f *StoredFile, err error) {
	if b.NameSize == 0 {
		err = ErrNoData
	} else {
		f = &StoredFile{}
		f.Data = b.FileData
		f.Path = string(b.FileName)
	}
	return
}

func (b *OSOHashHeadBlock) Parse(data []byte) error {
	if b.CRC != 0x6B6B {
		return ErrCRCMismatch
	}
	buffer := bytes.NewBuffer(data[7:])
	err := binary.Read(buffer, binary.LittleEndian, &b.FileSize)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.OSOHash)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.NameSize)
	if err != nil {
		return err
	}
	b.FileName = make([]byte, b.NameSize)
	err = binary.Read(buffer, binary.LittleEndian, &b.FileName)
	if err != nil {
		return err
	}
	return nil
}

func (b *OSOHashHeadBlock) GetOSOHash() (h *OSOHash, err error) {
	if b.FileSize == 0 || b.OSOHash == 0 || b.NameSize == 0 {
		err = ErrNoData
	} else {
		h = &OSOHash{}
		h.Hash = b.OSOHash
		h.Path = string(b.FileName)
		h.Size = b.FileSize
	}
	return
}

func (b *SrrRarPadHeadBlock) Parse(data []byte) error {
	if b.CRC != 0x6C6C {
		return ErrCRCMismatch
	}
	if !b.Flag(HAS_DATA) {
		return ErrBadBlock
	}
	buffer := bytes.NewBuffer(data[7:])
	err := binary.Read(buffer, binary.LittleEndian, &b.PadSize)
	if err != nil {
		return err
	}
	b.PadData = make([]byte, b.PadSize)
	err = binary.Read(buffer, binary.LittleEndian, &b.PadData)
	if err != nil {
		return err
	}
	return nil
}

func (b *SrrRarPadHeadBlock) GetSize() int {
	return int(b.RarHeader.Size) + int(b.PadSize)
}

func (b *SrrRarPadHeadBlock) GetPadSize() int {
	return int(b.PadSize)
}

func (b *SrrRarSubBlockHeadBlock) Parse(data []byte) error {
	if b.CRC != 0x7171 {
		return ErrCRCMismatch
	}
	buffer := bytes.NewBuffer(data[7:])
	err := binary.Read(buffer, binary.LittleEndian, &b.NameSize)
	if err != nil {
		return err
	}
	b.FileName = make([]byte, b.NameSize)
	err = binary.Read(buffer, binary.LittleEndian, &b.FileName)
	if err != nil {
		return err
	}
	return nil
}

func (b *SrrRarSubBlockHeadBlock) GetRarFileName() string {
	if b.NameSize == 0 {
		return ""
	}
	return string(b.FileName)
}

func (b *FileHeadBlock) Parse(data []byte) error {
	buffer := bytes.NewBuffer(data[7:])
	err := binary.Read(buffer, binary.LittleEndian, &b.LowPackSize)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.LowUnpackSize)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.HostOS)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.FileCRC)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.FileTime)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.UnpackVersion)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.Method)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.NameSize)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.FileAttr)
	if err != nil {
		return err
	}
	if b.Flag(LHD_LARGE) {
		err = binary.Read(buffer, binary.LittleEndian, &b.HighPackSize)
		if err != nil {
			return err
		}
		err = binary.Read(buffer, binary.LittleEndian, &b.HighUnpackSize)
		if err != nil {
			return err
		}
	}
	b.FileName = make([]byte, int(b.NameSize))
	err = binary.Read(buffer, binary.LittleEndian, &b.FileName)
	if err != nil {
		return err
	}
	if b.Flag(LHD_SALT) {
		err = binary.Read(buffer, binary.LittleEndian, &b.Salt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *FileHeadBlock) GetMethod() uint8 {
	return b.Method
}

func (b *FileHeadBlock) GetPackSize() int {
	if b.Flag(LHD_LARGE) {
		return int(b.HighPackSize)<<32 + int(b.LowPackSize)
	} else {
		return int(b.LowPackSize)
	}
}

func (b *FileHeadBlock) GetUnpackSize() int {
	if b.Flag(LHD_LARGE) {
		return int(b.HighUnpackSize)<<32 + int(b.LowUnpackSize)
	} else {
		return int(b.LowUnpackSize)
	}
}

func (b *FileHeadBlock) GetFileName() string {
	if b.NameSize > 0 {
		if b.Flag(LHD_UNICODE) {
			s := strings.Split(string(b.FileName), "\x00")
			return s[0]
		}
		return string(b.FileName)
	} else {
		return ""
	}
}

func (b *FileHeadBlock) GetCRC() uint32 {
	if b.NameSize > 0 {
		return b.FileCRC
	} else {
		return 0
	}
}

func (b *FileHeadBlock) UpdatePackedFile(p *PackedFile) error {
	if p.Path == "" {
		p.Path = b.GetFileName()
	}
	if p.Path != b.GetFileName() {
		return ErrBadData
	}
	p.CRC = b.GetCRC()

	if b.GetMethod() == 0x30 {
		p.Size += uint64(b.GetPackSize())
	} else {
		p.Size = uint64(b.GetUnpackSize())
	}

	return nil
}

func (b *FileHeadBlock) GetSize() int {
	return int(b.RarHeader.Size) + b.GetPackSize()
}

func (b *FileHeadBlock) IsCompressed() bool {
	if b.GetMethod() == 0x30 {
		return false
	} else {
		return true
	}
}

func (b *NewSubHeadBlock) Parse(data []byte) error {
	buffer := bytes.NewBuffer(data[7:])
	err := binary.Read(buffer, binary.LittleEndian, &b.LowPackSize)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.LowUnpackSize)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.HostOS)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.FileCRC)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.FileTime)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.UnpackVersion)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.Method)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.NameSize)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.FileAttr)
	if err != nil {
		return err
	}
	if b.Flag(LHD_LARGE) {
		err = binary.Read(buffer, binary.LittleEndian, &b.HighPackSize)
		if err != nil {
			return err
		}
		err = binary.Read(buffer, binary.LittleEndian, &b.HighUnpackSize)
		if err != nil {
			return err
		}
	}
	b.FileName = make([]byte, int(b.NameSize))
	err = binary.Read(buffer, binary.LittleEndian, &b.FileName)
	if err != nil {
		return err
	}

	if b.Flag(LHD_SALT) {
		err = binary.Read(buffer, binary.LittleEndian, &b.Salt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *NewSubHeadBlock) GetFileName() string {
	if b.NameSize > 0 {
		return string(b.FileName)
	} else {
		return ""
	}
}

func (b *NewSubHeadBlock) GetPackSize() int {
	if b.Flag(LHD_LARGE) {
		return int(b.HighPackSize)<<32 + int(b.LowPackSize)
	} else {
		return int(b.LowPackSize)
	}
}

func (b *NewSubHeadBlock) GetSize() int {
	return int(b.RarHeader.Size) + b.GetPackSize()
}

func (b *ProtectHeadBlock) Parse(data []byte) error {
	if !b.Flag(HAS_DATA) {
		return ErrBadBlock
	}
	buffer := bytes.NewBuffer(data[7:])
	err := binary.Read(buffer, binary.LittleEndian, &b.PackedSize)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.Version)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.RecSectorCount)
	if err != nil {
		return err
	}
	err = binary.Read(buffer, binary.LittleEndian, &b.DataSectorCount)
	if err != nil {
		return err
	}
	return nil
}

func (b *ProtectHeadBlock) GetSize() int {
	return int(b.RarHeader.Size) + int(b.PackedSize)
}
