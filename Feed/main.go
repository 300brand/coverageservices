package main

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/downloader"
	"git.300brand.com/coverage/feed"
	"git.300brand.com/coverage/skytypes"
	"git.300brand.com/coverageservices/skynetstats"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"log"
	"math/rand"
	"time"
)

type Service struct {
	Config *skynet.ServiceConfig
}

const ServiceName = "Feed"

var (
	_             service.ServiceDelegate = &Service{}
	Article       *client.ServiceClient
	Stats         *client.ServiceClient
	StorageReader *client.ServiceClient
	StorageWriter *client.ServiceClient
)

// Funcs required for ServiceDelegate

func (s *Service) MethodCalled(m string) {}

func (s *Service) MethodCompleted(m string, d int64, err error) {
	skynetstats.Completed(m, d, err)
}

func (s *Service) Registered(service *service.Service) {
	skynetstats.Start(s.Config, Stats)
}

func (s *Service) Started(service *service.Service) {}

func (s *Service) Stopped(service *service.Service) {}

func (s *Service) Unregistered(service *service.Service) {
	skynetstats.Stop()
}

// Service funcs

func (s *Service) Process(ri *skynet.RequestInfo, in *skytypes.ObjectId, out *skytypes.NullType) (err error) {
	f := &coverage.Feed{}
	if err = StorageReader.Send(ri, "Feed", in, f); err != nil {
		return
	}
	go func(ri *skynet.RequestInfo, f *coverage.Feed) {
		defer StorageWriter.SendOnce(ri, "Feed", f, f)

		stat := skytypes.Stat{Config: s.Config}

		start := time.Now()
		if err := downloader.Feed(f); err != nil {
			stat.Name, stat.Duration = "Process.Download.Failure", time.Since(start)
			Stats.SendOnce(ri, "Timer", stat, skytypes.Null)
			log.Printf("%s[%s] Error downloading: %s", f.ID.Hex(), f.URL, err)
			return
		}
		stat.Name, stat.Duration = "Process.Download.Success", time.Since(start)
		Stats.SendOnce(ri, "Timer", stat, skytypes.Null)

		start = time.Now()
		if err := feed.Process(f); err != nil {
			stat.Name, stat.Duration = "Process.Process.Failure", time.Since(start)
			Stats.SendOnce(ri, "Timer", stat, skytypes.Null)
			log.Printf("%s[%s] Error parsing: %s", f.ID.Hex(), f.URL, err)
			return
		}
		stat.Name, stat.Duration = "Process.Process.Success", time.Since(start)
		Stats.SendOnce(ri, "Timer", stat, skytypes.Null)

		stat.Name, stat.Count = "Process.NewArticles", len(f.Articles)
		Stats.SendOnce(ri, "Increment", stat, skytypes.Null)

		for _, a := range f.Articles {
			// Add a 5-15 second delay between article downloads
			<-time.After(time.Duration(rand.Int63n(10)+5) * time.Second)
			Article.SendOnce(ri, "Process", a, skytypes.Null)
		}
	}(ri, f)
	return
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	cc, _ := skynet.GetClientConfig()
	c := client.NewClient(cc)

	Article = c.GetService("Article", "", "", "")
	Stats = c.GetService("Stats", "", "", "")
	StorageReader = c.GetService("StorageReader", "", "", "")
	StorageWriter = c.GetService("StorageWriter", "", "", "")

	sc, _ := skynet.GetServiceConfig()
	sc.Name = ServiceName
	sc.Region = "Processing"
	sc.Version = "1"

	s := service.CreateService(&Service{sc}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
