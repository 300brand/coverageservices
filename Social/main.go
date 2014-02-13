package Social

import (
	"github.com/300brand/coverage"
	"github.com/300brand/coverage/social"
	"github.com/300brand/coverageservices/service"
	"github.com/300brand/disgo"
)

type Service struct{}

var _ service.Service = new(Service)

func init() {
	service.Register("Social", new(Service))
}

// Funcs required for Service

func (s *Service) Start(client *disgo.Client) (err error) { return }

// Service funcs

func (s *Service) Article(in *coverage.Article, out *social.Stats) (err error) {
	return social.FetchString(in.URL, out)
}
