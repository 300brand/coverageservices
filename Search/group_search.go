package Search

import (
	"github.com/300brand/coverage"
	"github.com/300brand/coverageservices/types"
	"labix.org/v2/mgo/bson"
)

func (s *Service) GroupSearch(in *types.GroupQuery, out *types.SearchQueryResponse) (err error) {
	gs := coverage.GroupSearch{
		Id: bson.NewObjectId(),
	}
	s.client.Call("StorageWriter.NewGroupSearch", gs, gs)

	// Prepare information for the caller
	out.Id = gs.Id
	out.Start = gs.Start
	return
}
