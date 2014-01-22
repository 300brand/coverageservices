package Search

import (
	"github.com/300brand/coverage"
	"github.com/300brand/coverageservices/types"
	"github.com/300brand/disgo"
	"github.com/300brand/logger"
	"labix.org/v2/mgo/bson"
	"sync"
	"time"
)

func (s *Service) GroupSearch(in *types.GroupQuery, out *types.SearchQueryResponse) (err error) {
	var gsLock sync.Mutex

	gs := &coverage.GroupSearch{
		Id: bson.NewObjectId(),
	}
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
			if err := s.client.Call("Search.NotifyComplete", types.ObjectId{gs.Id}, disgo.Null); err != nil {
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
