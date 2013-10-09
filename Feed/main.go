package Feed

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/downloader"
	"git.300brand.com/coverage/feed"
	"git.300brand.com/coverageservices/service"
	"git.300brand.com/coverageservices/types"
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/logger"
	"math/rand"
	"time"
)

type Service struct {
	client *disgo.Client
}

var _ service.Service = new(Service)

func init() {
	service.Register("Feed", new(Service))
}

// Funcs required for Service

func (s *Service) Start(client *disgo.Client) (err error) {
	s.client = client
	return
}

// Service funcs

func (s *Service) Process(in *types.ObjectId, out *disgo.NullType) (err error) {
	f := &coverage.Feed{}
	if err = s.client.Call("StorageReader.Feed", in, f); err != nil {
		logger.Error.Printf("Feed.Process: StorageReader.Feed error: %s", err)
		return
	}
	defer s.client.Call("StorageWriter.Feed", f, f)

	// start := time.Now()
	if err = downloader.Feed(f); err != nil {
		// skynetstats.Duration(time.Since(start), "Process.Download.Failure")
		logger.Error.Printf("%s[%s] Error downloading: %s", f.ID.Hex(), f.URL, err)
		return
	}
	// skynetstats.Duration(time.Since(start), "Process.Download.Success")
	// skynetstats.Count(len(f.Content), "Process.Bandwidth")

	// start = time.Now()
	if err = feed.Process(f); err != nil {
		// skynetstats.Duration(time.Since(start), "Process.Process.Failure")
		logger.Error.Printf("%s[%s] Error parsing: %s", f.ID.Hex(), f.URL, err)
		return
	}
	// skynetstats.Duration(time.Since(start), "Process.Process.Success")
	// skynetstats.Count(len(f.Articles), "Process.NewArticles")

	for _, a := range f.Articles {
		// Add a 5-15 second delay between article downloads
		<-time.After(time.Duration(rand.Int63n(10)+5) * time.Second)
		s.client.Call("Article.Process", a, disgo.Null)
	}
	return
}
