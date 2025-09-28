package release

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/remko/go-mkvparse"
)

type NFOHandler struct {
	mkvparse.DefaultHandler

	currentAttachmentData        []byte
	currentAttachmentFileName    string
	currentAttachmentMIMEType    string
	currentAttachmentDescription string

	Data        []byte
	FileName    string
	MIMEType    string
	Description string
}

func (p *NFOHandler) HandleMasterEnd(id mkvparse.ElementID, info mkvparse.ElementInfo) error {
	if id == mkvparse.AttachedFileElement {
		if strings.EqualFold(".nfo", filepath.Ext(p.currentAttachmentFileName)) {
			p.Data = p.currentAttachmentData
			p.FileName = p.currentAttachmentFileName
			p.MIMEType = p.currentAttachmentMIMEType
			p.Description = p.currentAttachmentDescription
		}
	}
	return nil
}

func (p *NFOHandler) HandleString(id mkvparse.ElementID, value string, info mkvparse.ElementInfo) error {
	switch id {
	case mkvparse.FileNameElement:
		p.currentAttachmentFileName = value
	case mkvparse.FileMimeTypeElement:
		p.currentAttachmentMIMEType = value
	case mkvparse.FileDescriptionElement:
		p.currentAttachmentDescription = value
	}
	return nil
}

func (p *NFOHandler) HandleBinary(id mkvparse.ElementID, value []byte, info mkvparse.ElementInfo) error {
	if id == mkvparse.FileDataElement {
		p.currentAttachmentData = value
	}
	return nil
}

func ParseNfoAttachment(path string) (NFOFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return NFOFile{}, err
	}
	defer f.Close()

	handler := NFOHandler{}
	if err := mkvparse.ParseSections(f, &handler, mkvparse.AttachmentsElement); err != nil {
		return NFOFile{}, err
	}

	return NFOFile{Name: handler.FileName, Content: handler.Data}, nil
}
