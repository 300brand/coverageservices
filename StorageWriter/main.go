package StorageWriter

import (
	"github.com/300brand/coverage"
	"github.com/300brand/coverage/storage/mongo"
	"github.com/300brand/coverageservices/service"
	"github.com/300brand/coverageservices/types"
	"github.com/300brand/disgo"
	"github.com/300brand/go-toml-config"
	"github.com/300brand/logger"
	"time"
)

type StorageWriter struct {
	client *disgo.Client
	m      *mongo.Mongo
}

var _ service.Service = new(StorageWriter)

var (
	// Prefix for database names (used when running production and testing in
	// same environment)
	cfgPrefix = config.String("StorageWriter.prefix", "A_")
	// Addresses of MongoDB, see labix.org/mgo for format details
	cfgMongoDB = config.String("StorageWriter.mongodb", "127.0.0.1")
)

func init() {
	service.Register("StorageWriter", new(StorageWriter))
}

// Funcs required for Service

func (s *StorageWriter) Start(client *disgo.Client) (err error) {
	s.client = client

	logger.Debug.Printf("StorageWriter: Connecting to MongoDB %s", *cfgMongoDB)
	s.m = mongo.New(*cfgMongoDB)
	s.m.Prefix = *cfgPrefix
	if err = s.m.Connect(); err != nil {
		logger.Error.Printf("StorageWriter: Failed to connect to MongoDB: %s", err)
		return
	}
	logger.Debug.Println("StorageWriter: Connected to MongoDB")
	return
}

// Service funcs

func (s *StorageWriter) DateSearch(in *types.DateSearch, out *disgo.NullType) (err error) {
	return s.m.DateSearch(in.Id, in.Query, in.Date)
}

func (s *StorageWriter) NewSearch(in *coverage.Search, out *coverage.Search) (err error) {
	*out = *in
	out.Start = time.Now()
	return s.m.UpdateSearch(out)
}

func (s *StorageWriter) Article(in *coverage.Article, out *coverage.Article) (err error) {
	start := time.Now()

	defer func() {
		*out = *in
		s.client.Call("Stats.Duration", types.Stat{
			Name:     "StorageWriter.Article.AddKeywords",
			Duration: time.Since(start),
		}, disgo.Null)
		logger.Info.Printf("Added [P:%s] [F:%s] [A:%s] %s", in.PublicationId.Hex(), in.FeedId.Hex(), in.ID.Hex(), in.URL)
	}()

	if err = s.m.AddURL(in.URL, in.ID); err != nil {
		logger.Warn.Printf("StorageWriter.Article: [P:%s] [F:%s] [A:%s] Duplicate URL: %s", in.PublicationId.Hex(), in.FeedId.Hex(), in.ID.Hex(), in.URL)
		return
	}
	if err = s.m.UpdateArticle(in); err != nil {
		return
	}
	if err := s.m.AddKeywords(in); err != nil {
		logger.Error.Printf("StorageWriter.Article: [P:%s] [F:%s] [A:%s] Error saving keywords: %s", in.PublicationId.Hex(), in.FeedId.Hex(), in.ID.Hex(), err)
	}
	return
}

func (s *StorageWriter) Feed(in *coverage.Feed, out *coverage.Feed) (err error) {
	defer func() {
		*out = *in
	}()

	logger.Debug.Printf("StorageWriter.Feed: [P:%s] [F:%s] %s", in.PublicationId.Hex(), in.ID.Hex(), in.LastDownload)
	if err = s.m.UpdateFeed(in); err != nil {
		return
	}
	return
}

func (s *StorageWriter) NextDownloadFeedId(in *types.DateThreshold, out *types.ObjectId) (err error) {
	return s.m.NextDownloadFeedId(in.Threshold, &out.Id)
}

func (s *StorageWriter) Publication(in *coverage.Publication, out *coverage.Publication) (err error) {
	defer func() {
		*out = *in
	}()

	if err = s.m.UpdatePublication(in); err != nil {
		return
	}
	return
}

func (s *StorageWriter) PubIncArticles(in *types.Inc, out *disgo.NullType) (err error) {
	return s.m.PublicationIncArticles(in.Id, in.Delta)
}

func (s *StorageWriter) PubIncFeeds(in *types.Inc, out *disgo.NullType) (err error) {
	return s.m.PublicationIncFeeds(in.Id, in.Delta)
}
