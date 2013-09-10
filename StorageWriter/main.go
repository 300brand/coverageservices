package main

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/config"
	"git.300brand.com/coverage/storage/mongo"
	"git.300brand.com/coverageservices/skynetstats"
	"git.300brand.com/coverageservices/skytypes"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"labix.org/v2/mgo/bson"
	"log"
	"time"
)

type Service struct {
	Config *skynet.ServiceConfig
}

const ServiceName = "StorageWriter"

var (
	_     service.ServiceDelegate = &Service{}
	m     *mongo.Mongo
	Stats *client.ServiceClient
)

// Funcs required for ServiceDelegate

func (s *Service) MethodCalled(m string) {}

func (s *Service) MethodCompleted(m string, d int64, err error) {
	skynetstats.Completed(m, d, err)
}

func (s *Service) Registered(service *service.Service) {
	skynetstats.Start(s.Config, Stats)
}

func (s *Service) Started(service *service.Service) {
	log.Printf("Connecting to MongoDB %s", config.Mongo.Host)
	m = mongo.New(config.Mongo.Host)
	m.Prefix = "A_"
	if err := m.Connect(); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %s", err)
	}
	log.Println("Connected to MongoDB")
}

func (s *Service) Stopped(service *service.Service) {
	log.Println("Closing MongoDB connection")
	m.Close()
}

func (s *Service) Unregistered(service *service.Service) {
	skynetstats.Stop()
}

// Service funcs

func (s *Service) DateSearch(ri *skynet.RequestInfo, in *skytypes.DateSearch, out *skytypes.NullType) (err error) {
	return m.DateSearch(in.Id, in.Query, in.Date)
}

func (s *Service) NewSearch(ri *skynet.RequestInfo, in *coverage.Search, out *coverage.Search) (err error) {
	*out = *in
	out.Start = time.Now()
	out.Id = bson.NewObjectId()
	return m.UpdateSearch(out)
}

func (s *Service) Article(ri *skynet.RequestInfo, in *coverage.Article, out *coverage.Article) (err error) {
	defer func() {
		*out = *in
	}()

	if err = m.AddURL(in.URL, in.ID); err != nil {
		log.Printf("Duplicate URL: %s", in.URL)
		return
	}
	if err = m.UpdateArticle(in); err != nil {
		return
	}
	go func(a *coverage.Article) {
		start := time.Now()
		if err := m.AddKeywords(in); err != nil {
			log.Printf("Error saving keywords: %s", err)
		}
		skynetstats.Duration(time.Since(start), "ArticleKeywords")
	}(in)
	return
}

func (s *Service) Feed(ri *skynet.RequestInfo, in *coverage.Feed, out *coverage.Feed) (err error) {
	defer func() {
		*out = *in
	}()

	if err = m.UpdateFeed(in); err != nil {
		return
	}
	return
}

func (s *Service) Publication(ri *skynet.RequestInfo, in *coverage.Publication, out *coverage.Publication) (err error) {
	defer func() {
		*out = *in
	}()

	if err = m.UpdatePublication(in); err != nil {
		return
	}
	return
}

func (s *Service) PubIncArticles(ri *skynet.RequestInfo, in *skytypes.Inc, out *skytypes.NullType) (err error) {
	return m.PublicationIncArticles(in.Id, in.Delta)
}

func (s *Service) PubIncFeeds(ri *skynet.RequestInfo, in *skytypes.Inc, out *skytypes.NullType) (err error) {
	return m.PublicationIncFeeds(in.Id, in.Delta)
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	cc, _ := skynet.GetClientConfig()
	c := client.NewClient(cc)

	Stats = c.GetService("Stats", "", "", "")

	sc, _ := skynet.GetServiceConfig()
	sc.Name = ServiceName
	sc.Region = "Storage"
	sc.Version = "1"

	s := service.CreateService(&Service{sc}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
