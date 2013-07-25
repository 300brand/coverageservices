package main

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/downloader"
	"git.300brand.com/coverage/skytypes"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"log"
)

type Service struct{}

const ServiceName = "Article"

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

func (s *Service) Process(ri *skynet.RequestInfo, in *coverage.Article, out *skytypes.NullType) (err error) {
	go func(ri *skynet.RequestInfo, a *coverage.Article) {
		log.Printf("%s[%s] Downloading", a.ID.Hex(), a.URL)
		if err := downloader.Article(a); err != nil {
			log.Printf("%s[%s] Error downloading: %s", a.ID.Hex(), a.URL, err)
			return
		}
		log.Printf("%s[%s] Saving", a.ID.Hex(), a.URL)
		if err := StorageWriter.SendOnce(ri, "SaveArticle", a, a); err != nil {
			log.Printf("%s[%s] Error saving: %s", a.ID.Hex(), a.URL, err)
		}
	}(ri, in)
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
	sc.Region = "Processing"
	sc.Version = "1"

	s := service.CreateService(&Service{}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
