package Manager

import (
	"github.com/300brand/coverage"
	"github.com/300brand/coverageservices/service"
	"github.com/300brand/coverageservices/types"
	"github.com/300brand/disgo"
	"github.com/300brand/go-toml-config"
	"github.com/300brand/logger"
	"time"
)

type Service struct {
	client *disgo.Client
}

var (
	_ service.Service = &Service{}

	cfgStartup           = config.Bool("Manager.startup", false)
	cfgArticleTick       = config.Duration("Manager.article.tick", 10*time.Second)
	cfgFeedTick          = config.Duration("Manager.feed.tick", 10*time.Second)
	cfgFeedDownloadDelay = config.Duration("Manager.feed.downloaddelay", -2*time.Hour)

	aProcessor *Ticker
	fProcessor *Ticker
)

func init() {
	service.Register("Manager", new(Service))
}

// Funcs required for Service

func (s *Service) Start(client *disgo.Client) (err error) {
	s.client = client

	aProcessor = NewTicker(s.processArticle, *cfgArticleTick)
	go aProcessor.Run()

	fProcessor = NewTicker(s.processFeed, *cfgFeedTick)
	go fProcessor.Run()

	if *cfgStartup {
		aProcessor.Start <- true
		fProcessor.Start <- true
	}
	return
}

// Service funcs

func (s *Service) ArticleProcessor(in *types.ClockCommand, out *disgo.NullType) (err error) {
	return aProcessor.ProcessCommand(in)
}

func (s *Service) FeedProcessor(in *types.ClockCommand, out *disgo.NullType) (err error) {
	return fProcessor.ProcessCommand(in)
}

func (s *Service) processArticle() (err error) {
	logger.Debug.Printf("processArticle: Getting Article")
	a := new(coverage.Article)
	if err = s.client.Call("StorageWriter.ArticleQueueNext", disgo.Null, a); err != nil {
		logger.Warn.Printf("No article to process right now: %s", err)
		return
	}
	logger.Debug.Printf("processArticle: Got %s", a.ID.Hex())
	return s.client.Call("Article.Process", a, disgo.Null)
}

func (s *Service) processFeed() (err error) {
	id := new(types.ObjectId)
	thresh := types.DateThreshold{time.Now().Add(*cfgFeedDownloadDelay)}
	logger.Debug.Printf("processFeed: Getting ID")
	if err = s.client.Call("StorageWriter.NextDownloadFeedId", thresh, id); err != nil {
		logger.Error.Printf("Manager.processFeed: %s", err)
		return
	}
	logger.Debug.Printf("processFeed: Got %s", id.Id.Hex())
	return s.client.Call("Feed.Process", id, disgo.Null)
}
