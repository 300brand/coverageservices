package StorageReader

import (
	"github.com/300brand/coverage"
	"github.com/300brand/coverage/storage/mongo"
	"github.com/300brand/coverageservices/service"
	"github.com/300brand/coverageservices/types"
	"github.com/300brand/disgo"
	"github.com/300brand/go-toml-config"
	"github.com/300brand/logger"
)

type StorageReader struct {
	client *disgo.Client
	m      *mongo.Mongo
}

var _ service.Service = new(StorageReader)

var (
	// Prefix for database names (used when running production and testing in
	// same environment)
	cfgPrefix = config.String("StorageReader.prefix", "A_")
	// Addresses of MongoDB, see labix.org/mgo for format details
	cfgMongoDB = config.String("StorageReader.mongodb", "127.0.0.1")
)

func init() {
	service.Register("StorageReader", new(StorageReader))
}

// Funcs required for Service

func (s *StorageReader) Start(client *disgo.Client) (err error) {
	s.client = client

	logger.Debug.Printf("StorageReader: Connecting to MongoDB %s", *cfgMongoDB)
	s.m = mongo.New(*cfgMongoDB)
	s.m.Prefix = *cfgPrefix
	if err = s.m.Connect(); err != nil {
		logger.Error.Printf("StorageReader: Failed to connect to MongoDB: %s", err)
		return
	}
	logger.Debug.Println("StorageReader: Connected to MongoDB")
	return
}

// Service funcs

func (s *StorageReader) Article(in *types.ObjectId, out *coverage.Article) error {
	return s.m.GetArticle(in.Id, out)
}

func (s *StorageReader) Articles(in *types.MultiQuery, out *types.MultiArticles) (err error) {
	objectIdify(&in.Query)

	if out.Total, err = s.m.C.Articles.Find(in.Query).Count(); err != nil {
		return
	}
	out.Query = *in
	out.Articles = make([]*coverage.Article, 0, in.Limit)
	return s.m.GetArticles(in.Query, in.Sort, in.Skip, in.Limit, in.Select, &out.Articles)
}

func (s *StorageReader) Feed(in *types.ObjectId, out *coverage.Feed) error {
	return s.m.GetFeed(in.Id, out)
}

func (s *StorageReader) Feeds(in *types.MultiQuery, out *types.MultiFeeds) (err error) {
	objectIdify(&in.Query)

	if out.Total, err = s.m.C.Feeds.Find(in.Query).Count(); err != nil {
		return
	}
	logger.Debug.Printf("[StorageReader.Feeds] Query: %+v Total Feeds: %d", in.Query, out.Total)
	out.Query = *in
	out.Feeds = make([]*coverage.Feed, 0, in.Limit)
	return s.m.GetFeeds(in.Query, in.Sort, in.Skip, in.Limit, in.Select, &out.Feeds)
}

func (s *StorageReader) OldestFeed(in *types.ObjectIds, out *coverage.Feed) error {
	return s.m.GetOldestFeed(in.Ids, out)
}

func (s *StorageReader) Publication(in *types.ObjectId, out *coverage.Publication) error {
	return s.m.GetPublication(in.Id, out)
}

func (s *StorageReader) Publications(in *types.MultiQuery, out *types.MultiPubs) (err error) {
	objectIdify(&in.Query)

	if out.Total, err = s.m.C.Publications.Find(in.Query).Count(); err != nil {
		return
	}
	out.Query = *in
	out.Publications = make([]*coverage.Publication, 0, in.Limit)
	return s.m.GetPublications(in.Query, in.Sort, in.Skip, in.Limit, &out.Publications)
}

func (s *StorageReader) Search(in *types.ObjectId, out *coverage.Search) error {
	return s.m.GetSearch(in.Id, out)
}

func (s *StorageReader) GroupSearch(in *types.ObjectId, out *coverage.GroupSearch) error {
	return s.m.GetGroupSearch(in.Id, out)
}

func (s *StorageReader) Stats(in *disgo.NullType, out *mongo.Stats) error {
	return s.m.GetStats(out)
}
