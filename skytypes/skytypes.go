package skytypes

import (
	"git.300brand.com/coverage"
	"github.com/skynetservices/skynet"
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

type DateSearch struct {
	Id    bson.ObjectId
	Date  time.Time
	Query string
}

type Inc struct {
	Id    bson.ObjectId
	Delta int
}

type NullType struct{}

type ObjectIds struct {
	Ids []bson.ObjectId
}

type ObjectId struct {
	Id bson.ObjectId
}

type MultiQuery struct {
	Query interface{}
	Sort  string
	Skip  int
	Limit int
}

type MultiPubs struct {
	Query        MultiQuery
	Total        int
	Publications []*coverage.Publication
}

type SearchQuery struct {
	Q      string
	Notify string
	Dates  struct {
		Start, End time.Time
	}
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

type Stat struct {
	Config     *skynet.ServiceConfig
	Name       string
	Count      int
	Duration   time.Duration
	Error      error
	Goroutines int
	Mem        runtime.MemStats
}

var Null = &NullType{}
