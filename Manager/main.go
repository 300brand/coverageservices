package Manager

import (
	"git.300brand.com/coverageservices/service"
	"git.300brand.com/coverageservices/types"
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/go-toml-config"
	"github.com/jbaikge/logger"
	"time"
)

type Service struct {
	client *disgo.Client
}

var (
	_ service.Service = &Service{}

	cfgStartup       = config.Bool("Manager.startup", false)
	cfgTick          = config.Duration("Manager.tick", 10*time.Second)
	cfgDownloadDelay = config.Duration("Manager.downloaddelay", -2*time.Hour)

	tProcessor *Ticker
)

func init() {
	service.Register("Manager", new(Service))
}

// Funcs required for Service

func (s *Service) Start(client *disgo.Client) (err error) {
	s.client = client

	tProcessor = NewTicker(s.processFeed, *cfgTick)
	go tProcessor.Run()

	if *cfgStartup {
		tProcessor.Start <- true
	}
	return
}

// Service funcs

func (s *Service) FeedProcessor(in *types.ClockCommand, out *disgo.NullType) (err error) {
	return tProcessor.ProcessCommand(in)
}

func (s *Service) processFeed() (err error) {
	id := new(types.ObjectId)
	thresh := types.DateThreshold{time.Now().Add(*cfgDownloadDelay)}
	logger.Debug.Printf("processFeed: Getting ID")
	if err = s.client.Call("StorageWriter.NextDownloadFeedId", thresh, id); err != nil {
		logger.Error.Printf("Manager.processFeed: %s", err)
		return
	}
	logger.Debug.Printf("processFeed: Got %s", id.Id.Hex())
	return s.client.Call("Feed.Process", id, disgo.Null)
}
