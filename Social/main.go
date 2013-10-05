package Social

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/social"
	"git.300brand.com/coverageservices/service"
	"github.com/jbaikge/disgo"
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
	return social.Fetch(in.URL, out)
}
