package main

import (
	"errors"
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/config"
	"git.300brand.com/coverage/doozer/idqueue"
	"git.300brand.com/coverage/skytypes"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"log"
)

type Service struct{}

const ServiceName = "Queue"

var (
	_             service.ServiceDelegate = &Service{}
	q             *idqueue.IdQueue
	Feed          *client.ServiceClient
	StorageReader *client.ServiceClient
)

// Funcs required for ServiceDelegate

func (s *Service) MethodCalled(m string) {}

func (s *Service) MethodCompleted(m string, d int64, err error) {}

func (s *Service) Registered(service *service.Service) {
	q = &idqueue.IdQueue{
		Addr: config.Doozer.Address,
		Max:  20,
		Name: service.Config.UUID,
	}
	log.Printf("Connecting to %s", q.Addr)
	if err := q.Connect(); err != nil {
		log.Fatal(err)
	}
}

func (s *Service) Started(service *service.Service) {}

func (s *Service) Stopped(service *service.Service) {
	q.Close()
}

func (s *Service) Unregistered(service *service.Service) {}

// Service funcs

func (s *Service) AddFeed(ri *skynet.RequestInfo, in *skytypes.NullType, out *skytypes.ObjectId) (err error) {
	// Determine if there's room to add new feeds
	ids, err := q.Get()
	if err != nil && err != idqueue.ErrEOQ {
		return
	}
	if len(ids) >= q.Max {
		return idqueue.ErrFull
	}

	// Find next oldest feed to insert
	f := &coverage.Feed{}
	if err = StorageReader.Send(nil, "OldestFeed", &skytypes.ObjectIds{Ids: ids}, f); err != nil {
		return
	}
	if f.ID.Hex() == "" {
		return errors.New("No feed found")
	}
	if err = q.Push(f.ID); err != nil {
		return
	}
	out.Id = f.ID
	return
}

func (s *Service) ProcessFeed(ri *skynet.RequestInfo, in *skytypes.NullType, out *skytypes.ObjectId) (err error) {
	if out.Id, err = q.Unshift(); err != nil {
		return
	}
	return Feed.SendOnce(ri, "Process", out, skytypes.Null)
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	cc, _ := skynet.GetClientConfig()
	c := client.NewClient(cc)

	StorageReader = c.GetService("StorageReader", "", "", "")
	Feed = c.GetService("Feed", "", "", "")

	sc, _ := skynet.GetServiceConfig()
	sc.Name = ServiceName
	sc.Region = "Management"
	sc.Version = "1"

	s := service.CreateService(&Service{}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
