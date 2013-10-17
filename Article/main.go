package Article

import (
	"github.com/300brand/coverage"
	"github.com/300brand/coverage/article/body"
	"github.com/300brand/coverage/article/lexer"
	"github.com/300brand/coverage/downloader"
	"github.com/300brand/coverageservices/service"
	"github.com/300brand/coverageservices/types"
	"github.com/300brand/disgo"
	"github.com/300brand/logger"
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

	hostChunks := strings.Split(in.URL.Host, ".")
	if len(hostChunks) > 2 {
		hostChunks = hostChunks[len(hostChunks)-2:]
	}
	domain := strings.Join(hostChunks, "_")

	// Download article
	if err = downloader.Article(in); err != nil {
		logger.Debug.Printf("%s %s.%s", time.Since(start), "Process.Download.Failure", domain)
		logger.Error.Printf("%s[%s] Error downloading: %s", in.ID.Hex(), in.URL, err)
		return
	}
	logger.Debug.Printf("%s %s.%s", time.Since(start), "Process.Download.Success", domain)
	logger.Debug.Printf("%d %s.%s", len(in.Text.HTML), "Process.Bandwidth", domain)

	// If any step fails along the way, save the article's state
	defer func() {
		if err = s.client.Call("StorageWriter.Article", in, in); err != nil {
			logger.Error.Printf("%s[%s] Error saving: %s", in.ID.Hex(), in.URL, err)
			return
		}
		inc := types.Inc{
			Id:    in.PublicationId,
			Delta: 1,
		}
		if err = s.client.Call("StorageWriter.PubIncArticles", inc, disgo.Null); err != nil {
			logger.Error.Printf("%s[%s] Error incrementing article count for pub [%s]: %s", in.ID.Hex(), in.URL, in.PublicationId.Hex(), err)
		}
	}()

	// Extract body
	start = time.Now()
	if err = body.SetBody(in); err != nil || in.Text.Body.Text == nil || len(in.Text.Body.Text) == 0 {
		logger.Debug.Printf("%s %s.%s", time.Since(start), "Process.Body.Failure", domain)
		logger.Error.Printf("%s[%s] Error setting body: %s", in.ID.Hex(), in.URL, err)
		return
	}
	logger.Debug.Printf("%s %s.%s", time.Since(start), "Process.Body.Success", domain)
	logger.Debug.Printf("%d %s.%s", len(in.Text.Body.Text), "Process.BodyLength", domain)

	// Filter out individual words
	in.Text.Words.All = lexer.Words(in.Text.Body.Text)
	logger.Debug.Printf("%d %s", len(in.Text.Words.All), "Process.Words")

	// Filter out Keywords
	in.Text.Words.Keywords = lexer.Keywords(in.Text.Body.Text)
	logger.Debug.Printf("%d %s", len(in.Text.Words.Keywords), "Process.Keywords")
	return
}
