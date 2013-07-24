package main

import (
	"git.300brand.com/coverage/skytypes"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"log"
	"time"
)

type Service struct{}

const ServiceName = "Manager"

var (
	_     service.ServiceDelegate = &Service{}
	t     *Ticker
	Queue *client.ServiceClient
)

// Funcs required for ServiceDelegate

func (s *Service) MethodCalled(m string) {}

func (s *Service) MethodCompleted(m string, d int64, err error) {}

func (s *Service) Registered(service *service.Service) {}

func (s *Service) Started(service *service.Service) {
	t = NewTicker(s.addFeed, time.Second*10)
	t.Start <- true
}

func (s *Service) Stopped(service *service.Service) {
	t.ProcessCommand(&skytypes.ClockCommand{Command: "stop"})
}

func (s *Service) Unregistered(service *service.Service) {}

// Service funcs

func (s *Service) FeedAdder(ri *skynet.RequestInfo, in *skytypes.ClockCommand, out *skytypes.ClockResult) (err error) {
	return t.ProcessCommand(in)
}

func (s *Service) addFeed() (err error) {
	id := &skytypes.ObjectId{}
	return Queue.SendOnce(nil, "AddFeed", skytypes.Null, id)
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	cc, _ := skynet.GetClientConfig()
	c := client.NewClient(cc)

	Queue = c.GetService("Queue", "", "", "")

	sc, _ := skynet.GetServiceConfig()
	sc.Name = ServiceName
	sc.Region = "Management"
	sc.Version = "1"

	s := service.CreateService(&Service{}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
