package StorageWriter

import (
	"git.300brand.com/coverage"
	"git.300brand.com/coverage/storage/mongo"
	"git.300brand.com/coverageservices/service"
	"git.300brand.com/coverageservices/types"
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/go-toml-config"
	"github.com/jbaikge/logger"
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
	defer func() {
		*out = *in
	}()

	if err = s.m.AddURL(in.URL, in.ID); err != nil {
		logger.Info.Printf("Duplicate URL: %s", in.URL)
		return
	}
	if err = s.m.UpdateArticle(in); err != nil {
		return
	}
	go func(a *coverage.Article) {
		start := time.Now()
		if err := s.m.AddKeywords(in); err != nil {
			logger.Error.Printf("Error saving keywords: %s", err)
		}
		s.client.Call("Stats.Duration", types.Stat{
			Name:     "StorageWriter.Article.AddKeywords",
			Duration: time.Since(start),
		}, disgo.Null)
	}(in)
	return
}

func (s *StorageWriter) Feed(in *coverage.Feed, out *coverage.Feed) (err error) {
	defer func() {
		*out = *in
	}()

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
