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
	Q      string
	Notify struct {
		Done   string
		Social string
	}
	Dates struct {
		Start, End time.Time
	}
	PublicationIds []bson.ObjectId
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
