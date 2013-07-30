package main

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/article/body"
	"git.300brand.com/coverage/article/lexer"
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
		if err := downloader.Article(a); err != nil {
			log.Printf("%s[%s] Error downloading: %s", a.ID.Hex(), a.URL, err)
			return
		}

		// If any step fails along the way, save the article's state
		defer func() {
			if err := StorageWriter.SendOnce(ri, "Article", a, a); err != nil {
				log.Printf("%s[%s] Error saving: %s", a.ID.Hex(), a.URL, err)
			}
		}()

		if err := body.SetBody(a); err != nil {
			log.Printf("%s[%s] Error setting body: %s", a.ID.Hex(), a.URL, err)
			return
		}

		a.Text.Words.All = lexer.Words(a.Text.Body.Text)
		a.Text.Words.Keywords = lexer.Keywords(a.Text.Body.Text)
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
