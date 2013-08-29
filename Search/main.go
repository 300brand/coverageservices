package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/skytypes"
	"git.300brand.com/coverageservices/skynetstats"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"log"
	"net/http"
	"sync"
	"time"
)

type Service struct {
	Config *skynet.ServiceConfig
}

const ServiceName = "Search"

var (
	_             service.ServiceDelegate = &Service{}
	Search        *client.ServiceClient
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

func (s *Service) NotifyComplete(ri *skynet.RequestInfo, in *skytypes.ObjectId, out *skytypes.NullType) (err error) {
	info := &coverage.Search{}
	if err = StorageReader.Send(ri, "Search", in, info); err != nil {
		return
	}

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err = enc.Encode(info); err != nil {
		return
	}

	if _, err = http.Post(info.Notify, "application/json", buf); err != nil {
		return
	}

	return
}

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
		Notify:   in.Notify,
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
	var wg sync.WaitGroup
	for _, ds.Date = range dates {
		wg.Add(1)
		go func(ds skytypes.DateSearch) {
			StorageWriter.SendOnce(ri, "DateSearch", ds, skytypes.Null)
			wg.Done()
		}(ds)
	}

	// Wait for all of the DateSearch calls to finish, then send the
	// notification of completeness
	go func(ri skynet.RequestInfo, cs *coverage.Search) {
		wg.Wait()
		if err := Search.SendOnce(&ri, "NotifyComplete", skytypes.ObjectId{cs.Id}, skytypes.Null); err != nil {
			log.Print(err)
		}

		// Track how long it took to do the search
		stat := skytypes.Stat{
			Config:   s.Config,
			Name:     "Search.Completed",
			Duration: time.Since(cs.Start),
		}
		Stats.SendOnce(&ri, "Duration", stat, skytypes.Null)
	}(*ri, cs)

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

	Search = c.GetService("Search", "", "", "")
	Stats = c.GetService("Stats", "", "", "")
	StorageReader = c.GetService("StorageReader", "", "", "")
	StorageWriter = c.GetService("StorageWriter", "", "", "")

	sc, _ := skynet.GetServiceConfig()
	sc.Name = ServiceName
	sc.Region = "Search"
	sc.Version = "1"

	s := service.CreateService(&Service{sc}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
