package Article

import (
	"fmt"
	"github.com/300brand/coverage"
	"github.com/300brand/coverage/article/body"
	"github.com/300brand/coverage/article/lexer"
	"github.com/300brand/coverage/downloader"
	"github.com/300brand/coverageservices/service"
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

	// Used for stats - coming soon
	// hostChunks := strings.Split(in.URL.Host, ".")
	// if len(hostChunks) > 2 {
	// 	hostChunks = hostChunks[len(hostChunks)-2:]
	// }
	// domain := strings.Join(hostChunks, "_")

	// Download article
	if err = downloader.Article(in); err != nil {
		logger.Error.Printf("%s Download error: %s", prefix, err)
		return
	}
	if l := len(in.Text.HTML); int64(l) == downloader.MaxFileSize {
		logger.Error.Printf("%s Download failure: Document larger than max file size (%d)", prefix, l)
		return fmt.Errorf("Document larger than max file size. URL: %s", in.URL)
	}
	logger.Debug.Printf("%s Download success. %d bytes took %s", prefix, len(in.Text.HTML), time.Since(start))

	// If any step fails along the way, save the article's state
	defer func() {
		if err = s.client.Call("StorageWriter.Article", in, in); err != nil {
			logger.Error.Printf("%s Error saving: %s", prefix, err)
		}
	}()

	// Extract body
	start = time.Now()
	if err = body.SetBody(in); err != nil || in.Text.Body.Text == nil || len(in.Text.Body.Text) == 0 {
		logger.Error.Printf("%s Body extraction error: %s", prefix, err)
		return
	}

	// Filter out individual words
	in.Text.Words.All = lexer.Words(in.Text.Body.Text)

	// Filter out Keywords
	in.Text.Words.Keywords = lexer.Keywords(in.Text.Body.Text)

	logger.Debug.Printf("%s Body Length: %d; Words: %d; Keywords: %d; Took: %s", prefix, len(in.Text.Body.Text), len(in.Text.Words.All), len(in.Text.Words.Keywords), time.Since(start))
	return
}
