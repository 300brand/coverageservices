package types

import (
	"git.300brand.com/coverage"
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
	Name       string
	Count      int
	Duration   time.Duration
	Error      error
	Goroutines int
	Mem        runtime.MemStats
}