package WebAPI

import (
	"database/sql"
	"fmt"
	"github.com/300brand/coverage"
	"github.com/300brand/coverageservices/types"
	"github.com/300brand/logger"
	"io"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "code.google.com/p/gosqlite/sqlite3"
)

func (s *Service) HandleExport(w http.ResponseWriter, r *http.Request) {
	bits := strings.SplitN(r.URL.Path, "/", 3)
	if len(bits) < 3 {
		http.Error(w, "No searchId supplied", http.StatusBadRequest)
	}
	qSearchId := bits[2]
	if !bson.IsObjectIdHex(qSearchId) {
		http.Error(w, "Invalid searchId "+qSearchId, http.StatusBadRequest)
		return
	}
	limit := 0
	if str := r.URL.Query().Get("limit"); str != "" {
		limit, _ = strconv.Atoi(str)
	}

	searchId := bson.ObjectIdHex(qSearchId)
	logger.Debug.Printf("Exporting results for %s", searchId)

	// See if the search is a group search:
	groupSearch := new(coverage.GroupSearch)
	if err := s.client.Call("StorageReader.GroupSearch", types.ObjectId{searchId}, groupSearch); err != nil {
		groupSearch.SearchIds = []bson.ObjectId{searchId}
	}

	filename := filepath.Join(os.TempDir(), "dbs", fmt.Sprintf("%s-%d.sqlite3", qSearchId, limit))
	os.MkdirAll(filepath.Dir(filename), 0755)
	f, err := os.Open(filename)
	if os.IsNotExist(err) {
		// Generate database:
		for _, id := range groupSearch.SearchIds {
			if err := s.generateExport(id, filename, limit); err != nil {
				logger.Error.Printf("[S:%s] HandleExport: %s", qSearchId, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		f, err = os.Open(filename)
	}
	defer f.Close()

	w.Header().Add("Content-Type", "application/x-sqlite3")
	if _, err := io.Copy(w, f); err != nil {
		logger.Error.Printf("[S:%s] HandleExport: %s", qSearchId, err)
	}
}

func (s *Service) generateExport(id bson.ObjectId, filename string, limit int) (err error) {
	var (
		aInsert    *sql.Stmt
		pInsert    *sql.Stmt
		sInsert    *sql.Stmt
		sqlCreates = []string{
			`CREATE TABLE IF NOT EXISTS Searches (
				search_id CHAR(24) PRIMARY KEY,
				query     TEXT,
				label     TEXT,
				duration  INTEGER
			)`,
			`CREATE TABLE IF NOT EXISTS Articles (
				article_id     CHAR(24),
				feed_id        CHAR(24),
				publication_id CHAR(24),
				search_id      CHAR(24),
				author         TEXT,
				title          TEXT,
				url            TEXT,
				body           TEXT,
				published      DATETIME,
				PRIMARY KEY    (article_id, search_id)
			)`,
			`CREATE TABLE IF NOT EXISTS Pubs (
				publication_id CHAR(24) PRIMARY KEY,
				title          TEXT,
				url            TEXT
			)`,
		}
	)

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return
	}

	search := new(coverage.Search)
	if err = s.client.Call("StorageReader.Search", types.ObjectId{id}, search); err != nil {
		return
	}

	if limit > 0 && len(search.Articles) > limit {
		search.Articles = search.Articles[:limit]
	}

	for _, sqlCreate := range sqlCreates {
		if _, err = tx.Exec(sqlCreate); err != nil {
			return
		}
	}

	if sInsert, err = tx.Prepare("INSERT INTO Searches VALUES (?, ?, ?, ?)"); err != nil {
		return
	}
	defer sInsert.Close()

	if aInsert, err = tx.Prepare("INSERT INTO Articles VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"); err != nil {
		return
	}
	defer aInsert.Close()

	if pInsert, err = tx.Prepare("INSERT OR IGNORE INTO Pubs VALUES (?, ?, ?)"); err != nil {
		return
	}
	defer pInsert.Close()

	if search.Complete == nil {
		logger.Warn.Printf("[S:%s] search.Complete is nil, setting to now", search.Id.Hex())
		t := time.Now()
		search.Complete = &t
	}

	if _, err = sInsert.Exec(
		search.Id.Hex(),
		search.Q,
		search.Label,
		(*search.Complete).Sub(search.Start).Nanoseconds(),
	); err != nil {
		return
	}

	pubMap := make(map[bson.ObjectId]bool, len(search.Articles))
	articles := &types.MultiArticles{
		Articles: make([]*coverage.Article, 0, len(search.Articles)),
	}
	aQuery := &types.MultiQuery{
		Query: bson.M{"_id": bson.M{"$in": search.Articles}},
		Select: bson.M{
			"log":        0,
			"text.html":  0,
			"text.words": 0,
		},
	}
	if err = s.client.Call("StorageReader.Articles", aQuery, articles); err != nil {
		return
	}
	for _, a := range articles.Articles {
		logger.Info.Printf("%s", a.ID.Hex())
		pubMap[a.PublicationId] = true
		if _, err = aInsert.Exec(
			a.ID.Hex(),
			a.FeedId.Hex(),
			a.PublicationId.Hex(),
			search.Id.Hex(),
			a.Author,
			a.Title,
			a.URL,
			string(a.Text.Body.Text),
			a.Published,
		); err != nil {
			return
		}
	}

	pubIds := make([]bson.ObjectId, 0, len(pubMap))
	for id := range pubMap {
		pubIds = append(pubIds, id)
	}
	pubs := &types.MultiPubs{
		Publications: make([]*coverage.Publication, 0, len(pubIds)),
	}
	pQuery := &types.MultiQuery{
		Query: bson.M{"_id": bson.M{"$in": pubIds}},
	}
	if err = s.client.Call("StorageReader.Publications", pQuery, pubs); err != nil {
		return
	}
	for _, p := range pubs.Publications {
		if _, err = pInsert.Exec(
			p.ID.Hex(),
			p.Title,
			p.URL,
		); err != nil {
			return
		}
	}

	if err = tx.Commit(); err != nil {
		return
	}
	return
}
