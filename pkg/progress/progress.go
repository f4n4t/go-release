package progress

import (
	"fmt"

	"github.com/f4n4t/progressbar/v3"
)

// Progress is an interface to represent a progress tracking mechanism with methods to set, update, and finalize progress.
type Progress interface {
	Set(value int) error
	Set64(value int64) error
	Finish() error
	Cancel()
	ChangeMax(value int)
	ChangeMax64(value int64)
	Read(buf []byte) (int, error)
	Close() error
	Write(buf []byte) (int, error)
}

// RealProgressBar is a progress bar that uses schollz/progressbar
type RealProgressBar struct {
	bar *progressbar.ProgressBar
}

func (p *RealProgressBar) Read(buf []byte) (int, error) {
	return p.bar.Read(buf)
}

func (p *RealProgressBar) Write(buf []byte) (int, error) {
	return p.bar.Write(buf)
}

func (p *RealProgressBar) Close() error {
	return p.bar.Close()
}

func (p *RealProgressBar) ChangeMax(value int) {
	p.bar.ChangeMax(value)
}

func (p *RealProgressBar) ChangeMax64(value int64) {
	p.bar.ChangeMax64(value)
}

func (p *RealProgressBar) Set(value int) error {
	return p.bar.Set(value)
}

func (p *RealProgressBar) Set64(value int64) error {
	return p.bar.Set64(value)
}

func (p *RealProgressBar) Finish() error {
	return p.bar.Finish()
}

func (p *RealProgressBar) Cancel() {
	fmt.Printf("\n")
}

// NoOpProgressBar is a progress bar that does nothing
type NoOpProgressBar struct{}

func (p *NoOpProgressBar) Read(buf []byte) (int, error) {
	return 0, nil
}

func (p *NoOpProgressBar) Write(buf []byte) (int, error) {
	return len(buf), nil
}

func (p *NoOpProgressBar) Close() error {
	return nil
}

func (p *NoOpProgressBar) ChangeMax(value int) {}

func (p *NoOpProgressBar) ChangeMax64(value int64) {}

func (p *NoOpProgressBar) Set(value int) error {
	return nil
}

func (p *NoOpProgressBar) Set64(value int64) error {
	return nil
}

func (p *NoOpProgressBar) Finish() error {
	return nil
}

func (p *NoOpProgressBar) Cancel() {}

// NewProgressBar returns a new Progress interface
func NewProgressBar[T int64 | uint64 | int](showProgress bool, totalLength T, showBytes bool, desc ...string) Progress {
	if !showProgress {
		return &NoOpProgressBar{}
	}
	bar := DefaultProgressBar(totalLength, showBytes, desc...)
	return &RealProgressBar{bar: bar}
}

// DefaultProgressBar returns a *progressbar.ProgressBar with customized settings
func DefaultProgressBar[T int64 | uint64 | int](totalLength T, showBytes bool, desc ...string) *progressbar.ProgressBar {
	options := []progressbar.Option{
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Printf("\n")
		}),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionUseANSICodes(true),
		progressbar.OptionShowBytes(showBytes),
	}

	if len(desc) > 0 && desc[0] != "" {
		options = append(options, progressbar.OptionSetDescription(desc[0]))
	}

	if !showBytes {
		// showCount is only useful if showBytes is not set, could slow down the progressbar
		options = append(options, progressbar.OptionShowCount())
	}

	return progressbar.NewOptions64(int64(totalLength), options...)
}
