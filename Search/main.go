package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/social"
	"git.300brand.com/coverageservices/skynetstats"
	"git.300brand.com/coverageservices/skytypes"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"labix.org/v2/mgo/bson"
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
	Social        *client.ServiceClient
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

	if _, err = http.Post(info.Notify.Done, "application/json", buf); err != nil {
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
		if err := Search.SendOnce(&ri, "Social", skytypes.ObjectId{cs.Id}, skytypes.Null); err != nil {
			log.Print(err)
		}
		skynetstats.Duration(time.Since(cs.Start), "Search.Completed")
	}(*ri, cs)

	// Prepare information for the caller
	out.Id = cs.Id
	out.Start = cs.Start

	return
}

func (s *Service) Social(ri *skynet.RequestInfo, in *skytypes.ObjectId, out *skytypes.NullType) (err error) {
	info := &coverage.Search{}
	if err = StorageReader.Send(ri, "Search", in, info); err != nil {
		return
	}

	go func(info coverage.Search) {
		for _, id := range info.Articles {
			go func(id bson.ObjectId) {
				// Get article from DB
				log.Printf("Getting %s from DB", id.Hex())
				a := &coverage.Article{}
				if err := StorageReader.Send(ri, "Article", skytypes.ObjectId{id}, a); err != nil {
					log.Print(err)
					return
				}
				// Get stats
				log.Printf("Calling Social.Article for %s", id.Hex())
				if err := Social.Send(ri, "Article", a, &a.Social); err != nil {
					log.Print(err)
					return
				}
				// Send stats to frontend
				stats := struct {
					ArticleId, SearchId bson.ObjectId
					Stats               social.Stats
				}{id, info.Id, a.Social}

				buf := new(bytes.Buffer)
				enc := json.NewEncoder(buf)
				if err = enc.Encode(stats); err != nil {
					return
				}
				log.Printf("Sending %+v to %s", stats, info.Notify.Social)
				if _, err = http.Post(info.Notify.Social, "application/json", buf); err != nil {
					return
				}
			}(id)
			<-time.After(1 * time.Second)
		}
	}(*info)

	return
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	cc, _ := skynet.GetClientConfig()
	c := client.NewClient(cc)

	Search = c.GetService("Search", "", "", "")
	Social = c.GetService("Social", "", "", "")
	Stats = c.GetService("Stats", "", "", "")
	StorageReader = c.GetService("StorageReader", "", "", "")
	StorageWriter = c.GetService("StorageWriter", "", "", "")

	Social.SetTimeout(30*time.Second, 60*time.Second)

	sc, _ := skynet.GetServiceConfig()
	sc.Name = ServiceName
	sc.Region = "Search"
	sc.Version = "1"

	s := service.CreateService(&Service{sc}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
