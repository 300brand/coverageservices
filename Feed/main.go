package main

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/downloader"
	"git.300brand.com/coverage/feed"
	"git.300brand.com/coverage/skytypes"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"log"
)

type Service struct{}

const ServiceName = "Feed"

var (
	_             service.ServiceDelegate = &Service{}
	StorageReader *client.ServiceClient
	StorageWriter *client.ServiceClient
)

// Funcs required for ServiceDelegate

func (s *Service) MethodCalled(m string)                        {}
func (s *Service) MethodCompleted(m string, d int64, err error) {}
func (s *Service) Registered(service *service.Service)          {}
func (s *Service) Started(service *service.Service)             {}
func (s *Service) Stopped(service *service.Service)             {}
func (s *Service) Unregistered(service *service.Service)        {}

// Service funcs

func (s *Service) Process(ri *skynet.RequestInfo, in *skytypes.ObjectId, out *skytypes.NullType) (err error) {
	f := &coverage.Feed{}
	if err = StorageReader.Send(ri, "GetFeed", in, f); err != nil {
		return
	}
	go func(f *coverage.Feed) {
		defer StorageWriter.SendOnce(ri, "SaveFeed", f, f)

		log.Printf("%s[%s] Downloading", f.ID, f.URL)
		if err := downloader.Feed(f); err != nil {
			log.Printf("%s[%s]Error downloading: %s", f.ID, f.URL, err)
			return
		}
		log.Printf("")

		if err := feed.Process(f); err != nil {
			log.Printf("%s[%s]Error parsing: %s", f.ID, f.URL, err)
			return
		}

		log.Printf("%s[%s] New Articles: %d", f.ID, f.URL, len(f.Articles))

		// TODO Send each article off for processing

		log.Printf("%s[%s] Completed", f.ID, f.URL)
	}(f)
	return
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	cc, _ := skynet.GetClientConfig()
	c := client.NewClient(cc)

	StorageReader = c.GetService("StorageReader", "", "", "")
	StorageWriter = c.GetService("StorageWriter", "", "", "")

	sc, _ := skynet.GetServiceConfig()
	sc.Name = ServiceName
	sc.Region = "Storage"
	sc.Version = "1"

	s := service.CreateService(&Service{}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
