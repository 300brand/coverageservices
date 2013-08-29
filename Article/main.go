package main

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/article/body"
	"git.300brand.com/coverage/article/lexer"
	"git.300brand.com/coverage/downloader"
	"git.300brand.com/coverage/skytypes"
	"git.300brand.com/coverageservices/skynetstats"
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
		stat := skytypes.Stat{Config: s.Config}
		start := time.Now()
		host := a.URL.Host
		tld := host[strings.LastIndex(host[:len(host)-4], ".")+1:]
		domain := strings.Replace(tld, ".", "_", -1)

		// Download article
		if err := downloader.Article(a); err != nil {
			stat.Name = "Process.Download.Failure." + domain
			stat.Duration = time.Since(start)
			Stats.SendOnce(ri, "Timing", stat, skytypes.Null)
			log.Printf("%s[%s] Error downloading: %s", a.ID.Hex(), a.URL, err)
			return
		}

		stat.Name, stat.Duration = "Process.Download.Success."+domain, time.Since(start)
		Stats.SendOnce(ri, "Timing", stat, skytypes.Null)

		stat.Name, stat.Count = "Process.Download.Size."+domain, len(a.Text.HTML)
		Stats.SendOnce(ri, "Increment", stat, skytypes.Null)

		// If any step fails along the way, save the article's state
		defer func() {
			if err := StorageWriter.SendOnce(ri, "Article", a, a); err != nil {
				log.Printf("%s[%s] Error saving: %s", a.ID.Hex(), a.URL, err)
			}
		}()

		// Extract body
		start = time.Now()
		if err := body.SetBody(a); err != nil || a.Text.Body.Text == nil || len(a.Text.Body.Text) == 0 {
			stat.Name = "Process.Body.Failure." + domain
			stat.Duration = time.Since(start)
			Stats.SendOnce(ri, "Timing", stat, skytypes.Null)
			log.Printf("%s[%s] Error setting body: %s", a.ID.Hex(), a.URL, err)
			return
		}

		stat.Name, stat.Duration = "Process.Body.Success."+domain, time.Since(start)
		Stats.SendOnce(ri, "Timing", stat, skytypes.Null)

		stat.Name, stat.Count = "Process.Body.Size."+domain, len(a.Text.Body.Text)
		Stats.SendOnce(ri, "Increment", stat, skytypes.Null)

		// Filter out individual words
		a.Text.Words.All = lexer.Words(a.Text.Body.Text)
		stat.Name, stat.Count = "Process.Words", len(a.Text.Words.All)
		Stats.SendOnce(ri, "Increment", stat, skytypes.Null)

		// Filter out Keywords
		a.Text.Words.Keywords = lexer.Keywords(a.Text.Body.Text)
		stat.Name, stat.Count = "Process.Keywords", len(a.Text.Words.Keywords)
		Stats.SendOnce(ri, "Increment", stat, skytypes.Null)
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
