package main

import (
	"fmt"
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/article/lexer"
	"git.300brand.com/coverage/search"
	"git.300brand.com/coverage/skytypes"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"labix.org/v2/mgo/bson"
	"log"
	"sync"
	"time"
)

type Service struct{}

type Payload struct {
	Id   bson.ObjectId
	Q    string
	Date time.Time
}

const ServiceName = "Search"

var (
	_             service.ServiceDelegate = &Service{}
	Search        *client.ServiceClient
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

func (s *Service) DateSearch(ri *skynet.RequestInfo, in *Payload, out *skytypes.NullType) (err error) {
	terms := lexer.Keywords([]byte(in.Q))
	boolSearch := search.NewBoolean(in.Q)
	idFilter := search.NewIdFilter(boolSearch.MinTerms())

	log.Printf("Terms: %s", terms)

	var wg sync.WaitGroup
	for _, term := range terms {
		wg.Add(1)
		go func(term string) {
			defer wg.Done()

			id := coverage.KeywordId{
				Date:    in.Date,
				Keyword: term,
			}
			kw := &coverage.Keyword{}
			log.Printf("Querying: %v", id)
			if err := StorageReader.Send(ri, "GetKeyword", id, kw); err != nil {
				log.Printf("Error in SearchReader.GetKeyword: %s", err)
				return
			}
			for _, id := range kw.Articles {
				idFilter.Chan <- id
			}
		}(term)
	}
	wg.Wait()

	log.Println("Finished GetKeywords")

	results := &skytypes.SearchResultSubset{
		Id:         in.Id,
		ArticleIds: idFilter.Ids(),
	}
	return StorageWriter.SendOnce(ri, "AddSearchResults", results, skytypes.Null)
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
		Q:        in.Q,
		Dates:    in.Dates,
		DaysLeft: len(dates),
	}
	if err = StorageWriter.SendOnce(ri, "NewSearch", cs, cs); err != nil {
		return
	}

	payload := Payload{
		Id: cs.Id,
		Q:  cs.Q,
	}
	for _, payload.Date = range dates {
		go Search.SendOnce(ri, "DateSearch", payload, skytypes.Null)
	}

	// Prepare information for the caller
	out.Id = cs.Id
	out.Start = time.Now()

	return
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	cc, _ := skynet.GetClientConfig()
	c := client.NewClient(cc)

	Search = c.GetService("Search", "", "", "")
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
