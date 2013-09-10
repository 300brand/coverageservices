package main

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/article/body"
	"git.300brand.com/coverage/article/lexer"
	"git.300brand.com/coverage/downloader"
	"git.300brand.com/coverageservices/skynetstats"
	"git.300brand.com/coverageservices/skytypes"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"log"
	"strings"
	"time"
)

type Service struct {
	Config *skynet.ServiceConfig
}

const ServiceName = "Article"

var (
	_             service.ServiceDelegate = &Service{}
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

func (s *Service) Process(ri *skynet.RequestInfo, in *coverage.Article, out *skytypes.NullType) (err error) {
	go func(ri *skynet.RequestInfo, a *coverage.Article) {
		start := time.Now()
		host := a.URL.Host
		tld := host[strings.LastIndex(host[:strings.LastIndex(host, ".")], ".")+1:]
		domain := strings.Replace(tld, ".", "_", -1)

		// Download article
		if err := downloader.Article(a); err != nil {
			skynetstats.Duration(time.Since(start), "Process.Download.Failure", domain)
			log.Printf("%s[%s] Error downloading: %s", a.ID.Hex(), a.URL, err)
			return
		}
		skynetstats.Duration(time.Since(start), "Process.Download.Success", domain)
		skynetstats.Count(len(a.Text.HTML), "Process.Bandwidth", domain)

		// If any step fails along the way, save the article's state
		defer func() {
			if err := StorageWriter.SendOnce(ri, "Article", a, a); err != nil {
				log.Printf("%s[%s] Error saving: %s", a.ID.Hex(), a.URL, err)
				return
			}
			inc := skytypes.Inc{
				Id:    a.PublicationId,
				Delta: 1,
			}
			if err := StorageWriter.SendOnce(ri, "PubIncArticles", inc, skytypes.Null); err != nil {
				log.Printf("%s[%s] Error incrementing article count for pub [%s]: %s", a.ID.Hex(), a.URL, a.PublicationId.Hex(), err)
			}
		}()

		// Extract body
		start = time.Now()
		if err := body.SetBody(a); err != nil || a.Text.Body.Text == nil || len(a.Text.Body.Text) == 0 {
			skynetstats.Duration(time.Since(start), "Process.Body.Failure", domain)
			log.Printf("%s[%s] Error setting body: %s", a.ID.Hex(), a.URL, err)
			return
		}
		skynetstats.Duration(time.Since(start), "Process.Body.Success", domain)
		skynetstats.Count(len(a.Text.Body.Text), "Process.BodyLength", domain)

		// Filter out individual words
		a.Text.Words.All = lexer.Words(a.Text.Body.Text)
		skynetstats.Count(len(a.Text.Words.All), "Process.Words")

		// Filter out Keywords
		a.Text.Words.Keywords = lexer.Keywords(a.Text.Body.Text)
		skynetstats.Count(len(a.Text.Words.Keywords), "Process.Keywords")
	}(ri, in)
	return
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	cc, _ := skynet.GetClientConfig()
	c := client.NewClient(cc)

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
