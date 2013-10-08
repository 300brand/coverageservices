package Manager

import (
	"git.300brand.com/coverageservices/service"
	"git.300brand.com/coverageservices/types"
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/go-toml-config"
	"time"
)

type Service struct {
	client *disgo.Client
}

var (
	_ service.Service = &Service{}

	cfgStartup = config.Bool("Manager.startup", false)
	cfgTick    = config.Duration("Manager.tick", 10*time.Second)

	tAdder     *Ticker
	tProcessor *Ticker
)

func init() {
	service.Register("Manager", new(Service))
}

// Funcs required for Service

func (s *Service) Start(client *disgo.Client) (err error) {
	s.client = client

	tAdder = NewTicker(s.addFeed, *cfgTick)
	go tAdder.Run()

	tProcessor = NewTicker(s.processFeed, *cfgTick)
	go tProcessor.Run()

	if *cfgStartup {
		tAdder.Start <- true
		tProcessor.Start <- true
	}
	return
}

// Service funcs

func (s *Service) FeedAdder(in *types.ClockCommand, out *types.ClockResult) (err error) {
	return tAdder.ProcessCommand(in)
}

func (s *Service) addFeed() (err error) {
	id := new(types.ObjectId)
	return s.client.Call("Queue.AddFeed", disgo.Null, id)
}

func (s *Service) FeedProcessor(in *types.ClockCommand, out *disgo.NullType) (err error) {
	return tProcessor.ProcessCommand(in)
}

func (s *Service) processFeed() (err error) {
	id := new(types.ObjectId)
	return s.client.Call("Queue.ProcessFeed", disgo.Null, id)
}
