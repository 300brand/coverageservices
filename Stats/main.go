package main

import (
	"git.300brand.com/coverage/config"
	"git.300brand.com/coverage/skytypes"
	"github.com/cyberdelia/statsd"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/service"
	"log"
)

type Service struct{}

const (
	ServiceName = "Stats"
)

var (
	_      service.ServiceDelegate = &Service{}
	client *statsd.Client
)

// Funcs required for ServiceDelegate

func (s *Service) MethodCalled(m string) {}

func (s *Service) MethodCompleted(m string, d int64, err error) {}

func (s *Service) Registered(service *service.Service) {
	log.Printf("Connecting to %s", config.StatsdServer.Address)
	var err error
	if client, err = statsd.Dial(config.StatsdServer.Address); err != nil {
		log.Fatalf("Could not connect to Statsd Server: %s", err)
	}
}

func (s *Service) Started(service *service.Service) {}

func (s *Service) Stopped(service *service.Service) {
	client.Close()
}

func (s *Service) Unregistered(service *service.Service) {}

// Service funcs

func (s *Service) Decrement(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	return
}

func (s *Service) Gauge(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	return
}

func (s *Service) Increment(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	return
}

func (s *Service) MethodCall(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	return
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	sc, _ := skynet.GetServiceConfig()
	sc.Name = ServiceName
	sc.Region = "Stats"
	sc.Version = "1"

	s := service.CreateService(&Service{}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
