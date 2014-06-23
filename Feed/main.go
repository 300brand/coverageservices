package Feed

import (
	"fmt"
	"github.com/300brand/coverage"
	"github.com/300brand/coverage/downloader"
	"github.com/300brand/coverage/feed"
	"github.com/300brand/coverageservices/service"
	"github.com/300brand/coverageservices/types"
	"github.com/300brand/disgo"
	"github.com/300brand/logger"
	"net/url"
	"sync/atomic"
	"time"
)

type Service struct {
	Active int32
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

func (s *Service) Add(in *types.NewFeed, out *coverage.Feed) (err error) {
	*out = *coverage.NewFeed()
	out.PublicationId = in.PublicationId
	if _, err = url.Parse(in.URL); err != nil {
		return
	}
	out.URL = in.URL
	if err = s.client.Call("StorageWriter.Feed", out, disgo.Null); err != nil {
		return
	}
	return s.client.Call("StorageWriter.PubIncFeeds", &types.Inc{Id: in.PublicationId, Delta: 1}, disgo.Null)
}

func (s *Service) Process(in *types.ObjectId, out *disgo.NullType) (err error) {
	start := time.Now()

	atomic.AddInt32(&s.Active, 1)
	s.client.Call("Stats.Gauge", &types.Stat{Name: "Feed.Process.Active", Count: int(s.Active)}, disgo.Null)

	defer func() {
		atomic.AddInt32(&s.Active, -1)
		s.client.Call("Stats.Gauge", &types.Stat{Name: "Feed.Process.Active", Count: int(s.Active)}, disgo.Null)
	}()

	f := &coverage.Feed{}
	if err = s.client.Call("StorageReader.Feed", in, f); err != nil {
		s.client.Call("Stats.Increment", &types.Stat{Name: "Feed.Process.Errors.Database", Count: 1}, disgo.Null)
		logger.Error.Printf("[F:%s] Error fetching: %s", in.Id, err)
		return
	}
	defer s.client.Call("StorageWriter.Feed", f, f)

	prefix := fmt.Sprintf("Feed.Process: [P:%s] [F:%s] [U:%s]", f.PublicationId.Hex(), f.ID.Hex(), f.URL)

	if err = downloader.Feed(f); err != nil {
		s.client.Call("Stats.Increment", &types.Stat{Name: "Feed.Process.Errors.Download", Count: 1}, disgo.Null)
		logger.Error.Printf("%s Error downloading: %s", prefix, err)
		return
	}
	s.client.Call("Stats.Increment", &types.Stat{Name: "Feed.Process.FeedSize", Count: len(f.Content)}, disgo.Null)

	if err = feed.Process(f); err != nil {
		s.client.Call("Stats.Increment", &types.Stat{Name: "Feed.Process.Errors.Process", Count: 1}, disgo.Null)
		logger.Error.Printf("%s Error parsing: %s", prefix, err)
		return
	}

	s.client.Call("Stats.Increment", &types.Stat{Name: "Feed.Process.NewArticles", Count: len(f.Articles)}, disgo.Null)
	for i, a := range f.Articles {
		// Separate the dequeue times by 1-minute intervals to spread out
		// processing
		a.Dequeue = a.Added.Add(time.Duration(i) * time.Minute)
		if err := s.client.Call("StorageWriter.ArticleQueueAdd", a, disgo.Null); err != nil {
			logger.Error.Printf("%s ArticleQueueAdd: %s", prefix, err)
		}
	}
	s.client.Call("Stats.Duration", &types.Stat{Name: "Feed.Process", Duration: time.Since(start)}, disgo.Null)
	return
}

func (s *Service) Remove(in *types.ObjectId, out *disgo.NullType) (err error) {
	f := new(coverage.Feed)
	if err = s.client.Call("StorageReader.Feed", in, f); err != nil {
		return
	}
	// Bail early if already deleted; don't want to mess up the feed count or
	// trigger another write to the db
	if f.Deleted {
		return
	}
	f.Deleted = true
	if err = s.client.Call("StorageWriter.Feed", f, disgo.Null); err != nil {
		return
	}
	return s.client.Call("StorageWriter.PubIncFeeds", &types.Inc{Id: f.PublicationId, Delta: -1}, disgo.Null)
}
