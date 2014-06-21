package types

import (
	"github.com/300brand/coverage"
	"github.com/300brand/coverage/storage/mongo"
	"labix.org/v2/mgo/bson"
	"runtime"
	"time"
)

type ClockCommand struct {
	Command string
	Tick    time.Duration
}

type ClockResult struct {
	Message string
}

type DateThreshold struct {
	Threshold time.Time
}

type DateSearch struct {
	Id    bson.ObjectId
	Date  time.Time
	Query string
}

type NewFeed struct {
	PublicationId bson.ObjectId
	URL           string
}

type Inc struct {
	Id    bson.ObjectId
	Delta int
}

type ObjectIds struct {
	Ids []bson.ObjectId
}

type ObjectId struct {
	Id bson.ObjectId
}

type MultiQuery struct {
	Query  bson.M
	Select interface{}
	Sort   string
	Skip   int
	Limit  int
}

type MultiArticles struct {
	Query    MultiQuery
	Total    int
	Articles []*coverage.Article
}

type MultiFeeds struct {
	Query MultiQuery
	Total int
	Feeds []*coverage.Feed
}

type MultiPubs struct {
	Query        MultiQuery
	Total        int
	Publications []*coverage.Publication
}

type Pub struct {
	Title      string
	URL        string
	Readership int64
	Feeds      []string
}

type SearchQuery struct {
	Q              string
	Label          string
	Notify         notify
	Dates          startend
	PublicationIds []bson.ObjectId
	Foreground     bool // During group queries, don't background the processing
	Version        int  // Version 0 or 1: convert simple query format; 2: Use complex format
}

type SearchQueryResponse struct {
	Id    bson.ObjectId
	Start time.Time
}

type SearchStatus struct {
	Id bson.ObjectId
}

type SearchResults struct {
	Id        bson.ObjectId
	Ready     bool
	Completed time.Time
	Articles  []coverage.Article
}

type Set struct {
	Id    bson.ObjectId
	Key   string
	Value interface{}
}

type Stat struct {
	Name       string
	Count      int
	Duration   time.Duration
	Error      error
	Goroutines int
	Mem        runtime.MemStats
}

type Stats struct {
	Database mongo.Stats
}

type GroupQuery struct {
	Queries        []query
	Notify         notify
	Dates          startend
	PublicationIds []bson.ObjectId
}

type ViewPub struct {
	Publication *coverage.Publication
	Feeds       MultiFeeds
	Articles    MultiArticles
}

type ViewPubQuery struct {
	Publication bson.ObjectId
	Feeds       MultiQuery
	Articles    MultiQuery
}

// Private types
type notify struct {
	Done   string
	Social string
}

type query struct {
	Q     string
	Label string
}

type startend struct {
	Start, End time.Time
}
