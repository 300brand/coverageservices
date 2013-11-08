package Article

import (
	"fmt"
	"github.com/300brand/coverage"
	"github.com/300brand/coverage/article/body"
	"github.com/300brand/coverage/article/lexer"
	"github.com/300brand/coverage/downloader"
	"github.com/300brand/coverageservices/service"
	"github.com/300brand/coverageservices/types"
	"github.com/300brand/disgo"
	"github.com/300brand/logger"
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
	prefix := fmt.Sprintf("Article.Process: [P:%s] [F:%s] [A:%s] [U:%s]", in.PublicationId.Hex(), in.FeedId.Hex(), in.ID.Hex(), in.URL)

	// Download article
	if err = downloader.Article(in); err != nil {
		s.client.Call("Stats.Increment", &types.Stat{Name: "Article.Process.Errors.Download", Count: 1}, disgo.Null)
		logger.Error.Printf("%s Download error: %s", prefix, err)
		return
	}
	if l := len(in.Text.HTML); int64(l) == downloader.MaxFileSize {
		s.client.Call("Stats.Increment", &types.Stat{Name: "Article.Process.Errors.DownloadTooBig", Count: 1}, disgo.Null)
		err = fmt.Errorf("Document larger than max file size (%d)", downloader.MaxFileSize)
		logger.Error.Printf("%s Download failure: %s", prefix, err)
		return
	}
	s.client.Call("Stats.Increment", &types.Stat{Name: "Article.Process.HTML.Size", Count: len(in.Text.HTML)}, disgo.Null)
	logger.Debug.Printf("%s Download success. %d bytes took %s", prefix, len(in.Text.HTML), time.Since(start))

	// If any step fails along the way, save the article's state
	defer func() {
		if err = s.client.Call("StorageWriter.Article", in, in); err != nil {
			s.client.Call("Stats.Increment", &types.Stat{Name: "Article.Process.Errors.Database", Count: 1}, disgo.Null)
			logger.Error.Printf("%s Error saving: %s", prefix, err)
		}
	}()

	// Extract body
	start = time.Now()
	if err = body.SetBody(in); err != nil || in.Text.Body.Text == nil || len(in.Text.Body.Text) == 0 {
		s.client.Call("Stats.Increment", &types.Stat{Name: "Article.Process.Errors.BodyExtraction", Count: 1}, disgo.Null)
		logger.Error.Printf("%s Body extraction error: %s", prefix, err)
		return
	}
	s.client.Call("Stats.Increment", &types.Stat{Name: "Article.Process.Body.Size", Count: len(in.Text.Body.Text)}, disgo.Null)

	// Filter out individual words
	in.Text.Words.All = lexer.Words(in.Text.Body.Text)
	s.client.Call("Stats.Increment", &types.Stat{Name: "Article.Process.Body.Words", Count: len(in.Text.Words.All)}, disgo.Null)

	// Filter out Keywords
	in.Text.Words.Keywords = lexer.Keywords(in.Text.Body.Text)
	s.client.Call("Stats.Increment", &types.Stat{Name: "Article.Process.Body.Keywords", Count: len(in.Text.Words.Keywords)}, disgo.Null)

	s.client.Call("Stats.Duration", &types.Stat{Name: "Article.Process", Duration: time.Since(start)}, disgo.Null)
	logger.Debug.Printf("%s Body Length: %d; Words: %d; Keywords: %d; Took: %s", prefix, len(in.Text.Body.Text), len(in.Text.Words.All), len(in.Text.Words.Keywords), time.Since(start))
	return
}
