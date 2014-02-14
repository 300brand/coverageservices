package Search

import (
	"bytes"
	"encoding/json"
	"github.com/300brand/coverage"
	"github.com/300brand/coverageservices/types"
	"github.com/300brand/disgo"
	"github.com/300brand/logger"
	"net/http"
	"sync"
	"time"
)

func (s *Service) GroupSearch(in *types.GroupQuery, out *types.SearchQueryResponse) (err error) {
	var gsLock sync.Mutex

	gs := coverage.NewGroupSearch()
	gs.Notify.Done = in.Notify.Done
	s.client.Call("StorageWriter.NewGroupSearch", gs, gs)

	searchQuery := types.SearchQuery{
		Dates:          in.Dates,
		PublicationIds: in.PublicationIds,
		Foreground:     true,
	}
	// Do not want the complete notification to send out after each sub-search
	searchQuery.Notify.Social = in.Notify.Social

	var wg sync.WaitGroup
	for _, q := range in.Queries {
		wg.Add(1)
		searchQuery.Q = q.Q
		searchQuery.Label = q.Label
		go func(sq types.SearchQuery) {
			resp := new(types.SearchQueryResponse)
			s.client.Call("Search.Search", sq, resp)
			gsLock.Lock()
			gs.SearchIds = append(gs.SearchIds, resp.Id)
			gsLock.Unlock()
			wg.Done()
		}(searchQuery)
	}

	// Wait for all of the DateSearch calls to finish, then send the
	// notification of completeness
	go func(gs *coverage.GroupSearch) {
		wg.Wait()
		// This is a little manual, but it's explicit
		gs.Complete = new(time.Time)
		*gs.Complete = time.Now()
		if err := s.client.Call("StorageWriter.UpdateGroupSearch", gs, disgo.Null); err != nil {
			logger.Error.Printf("StorageWriter.UpdateGroupSearch: %s", err)
			return
		}
		if gs.Notify.Done != "" {
			if err := s.client.Call("Search.GroupSearchNotifyComplete", types.ObjectId{gs.Id}, disgo.Null); err != nil {
				logger.Error.Print(err)
			}
		}
		logger.Info.Printf("Group Search completed in %s", time.Since(gs.Start))
	}(gs)

	// Prepare information for the caller
	out.Id = gs.Id
	out.Start = gs.Start
	return
}

func (s *Service) GroupSearchNotifyComplete(in *types.ObjectId, out *disgo.NullType) (err error) {
	info := new(coverage.GroupSearch)
	if err = s.client.Call("StorageReader.GroupSearch", in, info); err != nil {
		return
	}

	info.Searches = make([]coverage.Search, len(info.SearchIds))
	for i, id := range info.SearchIds {
		if err := s.client.Call("StorageReader.Search", types.ObjectId{id}, &info.Searches[i]); err != nil {
			logger.Error.Printf("Error fetching Group Sub-Search: %s", err)
		}
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
