package Publication

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverageservices/service"
	"git.300brand.com/coverageservices/types"
	"github.com/jbaikge/disgo"
	"net/url"
)

type PubsArr []types.Pub

type Service struct {
	client *disgo.Client
}

var _ service.Service = new(Service)

func init() {
	service.Register("Publication", new(Service))
}

// Funcs required for Service

func (s *Service) Start(client *disgo.Client) (err error) {
	s.client = client
	return
}

// Service funcs

func (s *Service) Add(in *types.Pub, out *coverage.Publication) (err error) {
	p := coverage.NewPublication()
	p.Title = in.Title
	p.Readership = in.Readership
	if p.URL, err = url.Parse(in.URL); err != nil {
		return
	}
	feeds := make([]*coverage.Feed, len(in.Feeds))
	for i, feedUrl := range in.Feeds {
		feeds[i] = coverage.NewFeed()
		feeds[i].PublicationId = p.ID
		if feeds[i].URL, err = url.Parse(feedUrl); err != nil {
			return
		}
		p.NumFeeds++
	}
	if err = s.client.Call("StorageWriter.Publication", p, disgo.Null); err != nil {
		return
	}
	for _, f := range feeds {
		if err = s.client.Call("StorageWriter.Feed", f, disgo.Null); err != nil {
			continue
		}
	}
	*out = *p
	return
}

func (s *Service) AddAll(in *PubsArr, out *disgo.NullType) (err error) {
	p := new(coverage.Publication)
	for _, pub := range []types.Pub(*in) {
		if err = s.Add(&pub, p); err != nil {
			return
		}
	}
	return
}
