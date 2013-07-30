package main

import (
	"fmt"
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/skytypes"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"log"
	"time"
)

type Service struct{}

const ServiceName = "Search"

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

func (s *Service) Search(ri *skynet.RequestInfo, in *skytypes.SearchQuery, out *skytypes.SearchQueryResponse) (err error) {
	// Validation
	if in.Q == "" {
		return fmt.Errorf("Query cannot be empty")
	}
	if in.Dates.Start.After(in.Dates.End) {
		return fmt.Errorf("Invalid date range: %s > %s", in.Dates.Start, in.Dates.End)
	}

	// This is just silly, but most efficient way to calculate
	dates := []time.Time{}
	for st, t := in.Dates.Start.AddDate(0, 0, -1), in.Dates.End; t.After(st); t = t.AddDate(0, 0, -1) {
		dates = append(dates, t)
	}

	cs := &coverage.Search{
		Q:        in.Q,
		Dates:    in.Dates,
		DaysLeft: len(dates),
	}
	if err = StorageWriter.SendOnce(ri, "NewSearch", cs, cs); err != nil {
		return
	}

	ds := skytypes.DateSearch{
		Id:    cs.Id,
		Query: cs.Q,
	}
	for _, ds.Date = range dates {
		go StorageWriter.SendOnce(ri, "DateSearch", ds, skytypes.Null)
	}

	// Prepare information for the caller
	out.Id = cs.Id
	out.Start = cs.Start

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
	sc.Region = "Search"
	sc.Version = "1"

	s := service.CreateService(&Service{}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
