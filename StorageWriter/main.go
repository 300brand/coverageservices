package main

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/config"
	"git.300brand.com/coverage/storage/mongo"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/service"
	"log"
)

type Service struct{}

const ServiceName = "StorageWriter"

var (
	_ service.ServiceDelegate = &Service{}
	m *mongo.Mongo
)

// Funcs required for ServiceDelegate

func (s *Service) MethodCalled(m string) {}

func (s *Service) MethodCompleted(m string, d int64, err error) {}

func (s *Service) Registered(service *service.Service) {}

func (s *Service) Started(service *service.Service) {
	host, db := config.Mongo.Host, config.Mongo.Database
	log.Printf("Connecting to MongoDB %s %s", host, db)
	m = mongo.New(host, db)
	if err := m.Connect(); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %s", err)
	}
	log.Println("Connected to MongoDB")
}

func (s *Service) Stopped(service *service.Service) {
	log.Println("Closing MongoDB connection")
	m.Close()
}

func (s *Service) Unregistered(service *service.Service) {}

// Service funcs

func (s *Service) SaveArticle(ri *skynet.RequestInfo, in *coverage.Article, out *coverage.Article) (err error) {
	if err = m.UpdateArticle(in); err != nil {
		return
	}
	*out = *in
	return
}

func (s *Service) SaveFeed(ri *skynet.RequestInfo, in *coverage.Feed, out *coverage.Feed) (err error) {
	if err = m.UpdateFeed(in); err != nil {
		return
	}
	*out = *in
	return
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	sc, _ := skynet.GetServiceConfig()
	sc.Name = ServiceName
	sc.Region = "Storage"
	sc.Version = "1"

	s := service.CreateService(&Service{}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
