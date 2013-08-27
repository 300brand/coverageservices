package main

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/config"
	"git.300brand.com/coverage/skytypes"
	"git.300brand.com/coverage/storage/mongo"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"log"
	"runtime"
)

type Service struct {
	Config *skynet.ServiceConfig
}

const ServiceName = "StorageReader"

var (
	_     service.ServiceDelegate = &Service{}
	m     *mongo.Mongo
	Stats *client.ServiceClient
)

// Funcs required for ServiceDelegate

func (s *Service) MethodCalled(m string) {}

func (s *Service) MethodCompleted(m string, d int64, err error) {
	stat := skytypes.Stat{
		Config:     s.Config,
		Name:       m,
		Nanos:      d,
		Error:      err,
		Goroutines: runtime.NumGoroutine(),
	}
	runtime.ReadMemStats(&stat.Mem)
	Stats.SendOnce(nil, "Completed", stat, skytypes.Null)
}

func (s *Service) Registered(service *service.Service) {}

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

func (s *Service) Unregistered(service *service.Service) {}

// Service funcs

func (s *Service) Article(ri *skynet.RequestInfo, in *skytypes.ObjectId, out *coverage.Article) (err error) {
	return m.GetArticle(in.Id, out)
}

func (s *Service) Feed(ri *skynet.RequestInfo, in *skytypes.ObjectId, out *coverage.Feed) (err error) {
	return m.GetFeed(in.Id, out)
}

func (s *Service) OldestFeed(ri *skynet.RequestInfo, in *skytypes.ObjectIds, out *coverage.Feed) (err error) {
	return m.GetOldestFeed(in.Ids, out)
}

func (s *Service) Search(ri *skynet.RequestInfo, in *skytypes.ObjectId, out *coverage.Search) (err error) {
	return m.GetSearch(in.Id, out)
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
