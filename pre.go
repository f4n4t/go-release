package release

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/f4n4t/go-release/pkg/predbnet"
	"github.com/f4n4t/go-release/pkg/xrel"
)

// Pre is the struct that holds the pre-information.
type Pre struct {
	Name    string    `json:"name"`
	Group   string    `json:"group"`
	Section string    `json:"section"`
	Genre   string    `json:"genre"`
	Size    int64     `json:"size"`
	Files   int       `json:"files"`
	Nuke    string    `json:"nuke"`
	Time    time.Time `json:"pre_time"`
	Site    string    `json:"site"`
}

// GetPre searches for a pre on all available sources
// It ignores errors and returns nil if no pre was found.
func (s *Service) GetPre(name string) *Pre {
	const searchTimeout = 3 * time.Second

	preServices := []func(ctx context.Context, name string) (*Pre, error){
		s.searchPreNet,
		s.searchXREL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()

	resultChan := make(chan *Pre)
	wg := sync.WaitGroup{}

	for _, searchFunc := range preServices {
		wg.Go(func() {
			func(searchFunc func(context.Context, string) (*Pre, error)) {
				pre, err := searchFunc(ctx, name)
				if ctx.Err() != nil || err != nil || pre == nil {
					return
				}

				select {
				case resultChan <- pre:
				case <-ctx.Done():
					return
				}
			}(searchFunc)
		})
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	select {
	case pre, ok := <-resultChan:
		if ok {
			s.log.Debug().Str("site", pre.Site).Msg("found pre information")
			return pre
		}
		return nil
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			s.log.Debug().Msg("timeout while searching for pre information")
		}
		return nil
	}
}

// searchPreNet retrieves pre-information details from predb.net given a release name.
func (s *Service) searchPreNet(ctx context.Context, name string) (*Pre, error) {
	preRes, err := predbnet.GetWithContext(ctx, name)
	if err != nil {
		s.log.Debug().Err(err).Str("site", "predb.net").Msg("")
		return nil, err
	}

	pre := &Pre{
		Name:    preRes.Release,
		Group:   preRes.Group,
		Section: preRes.Section,
		Genre:   preRes.Genre,
		//Size: preRes.Size,
		Files: preRes.Files,
		Nuke:  preRes.Reason,
		Time:  time.Unix(preRes.PreTime, 0),
		Site:  "predb.net",
	}

	return pre, nil
}

// searchXREL retrieves release information from xrel.to based on the provided name and maps it to a Pre struct.
func (s *Service) searchXREL(ctx context.Context, name string) (*Pre, error) {
	xrelRes, err := xrel.GetWithContext(ctx, name)
	if err != nil {
		s.log.Debug().Err(err).Str("site", "xrel.to").Msg("")
		return nil, err
	}

	pre := &Pre{
		Name:    xrelRes.Dirname,
		Time:    time.Unix(int64(xrelRes.Time), 0),
		Group:   xrelRes.GroupName,
		Section: xrelRes.ExtInfo.Type,
		Site:    "xrel.to",
	}

	return pre, nil
}
