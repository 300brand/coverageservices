package Article

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/article/body"
	"git.300brand.com/coverage/article/lexer"
	"git.300brand.com/coverage/downloader"
	"git.300brand.com/coverageservices/service"
	"git.300brand.com/coverageservices/types"
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/logger"
	"strings"
	"time"
)

type Service struct {
	client *disgo.Client
}

var _ service.Service = new(Service)

func init() {
	service.Register("Article", new(Service))
}

// Funcs required for Service

func (s *Service) Start(client *disgo.Client) (err error) {
	s.client = client
	return
}

// Service funcs

func (s *Service) Process(in *coverage.Article, out *disgo.NullType) (err error) {
	start := time.Now()

	hostChunks := strings.Split(a.URL.Host, ".")
	if len(hostChunks) > 2 {
		hostChunks = hostChunks[len(hostChunks)-2:]
	}
	domain := strings.Join(hostChunks, "_")

	// Download article
	if err := downloader.Article(a); err != nil {
		logger.Debug.Printf("%s %s.%s", time.Since(start), "Process.Download.Failure", domain)
		logger.Error.Printf("%s[%s] Error downloading: %s", a.ID.Hex(), a.URL, err)
		return
	}
	logger.Debug.Printf("%s %s.%s", time.Since(start), "Process.Download.Success", domain)
	logger.Debug.Printf("%d %s.%s", len(a.Text.HTML), "Process.Bandwidth", domain)

	// If any step fails along the way, save the article's state
	defer func() {
		if err := s.client.Call("StorageWriter.Article", a, a); err != nil {
			logger.Error.Printf("%s[%s] Error saving: %s", a.ID.Hex(), a.URL, err)
			return
		}
		inc := types.Inc{
			Id:    a.PublicationId,
			Delta: 1,
		}
		if err := s.client.Call("StorageWriter.PubIncArticles", inc, disgo.Null); err != nil {
			logger.Error.Printf("%s[%s] Error incrementing article count for pub [%s]: %s", a.ID.Hex(), a.URL, a.PublicationId.Hex(), err)
		}
	}()

	// Extract body
	start = time.Now()
	if err := body.SetBody(a); err != nil || a.Text.Body.Text == nil || len(a.Text.Body.Text) == 0 {
		logger.Debug.Printf("%s %s.%s", time.Since(start), "Process.Body.Failure", domain)
		logger.Error.Printf("%s[%s] Error setting body: %s", a.ID.Hex(), a.URL, err)
		return
	}
	logger.Debug.Printf("%s %s.%s", time.Since(start), "Process.Body.Success", domain)
	logger.Debug.Printf("%d %s.%s", len(a.Text.Body.Text), "Process.BodyLength", domain)

	// Filter out individual words
	a.Text.Words.All = lexer.Words(a.Text.Body.Text)
	logger.Debug.Printf("%d %s", len(a.Text.Words.All), "Process.Words")

	// Filter out Keywords
	a.Text.Words.Keywords = lexer.Keywords(a.Text.Body.Text)
	logger.Debug.Printf("%d %s", len(a.Text.Words.Keywords), "Process.Keywords")
	return
}
