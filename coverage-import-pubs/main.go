package main

import (
	"bytes"
	"database/sql"
	"git.300brand.com/coverageservices/skytypes"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/rpc/json"
	"log"
	"net/http"
)

const APIURL = "http://192.168.20.20:8080/rpc"

var query = `
SELECT
	p.object_id AS pub_id,
	pu.attribute_value AS pub_url,
	REPLACE(pn.attribute_value, '&#39;', '\'') AS pub_name,
	fu.attribute_value AS feed_url
FROM
	objects AS p
	LEFT JOIN objects AS f ON(f.object_parent_id = p.object_id)
	LEFT JOIN attributes AS pu ON(
		p.object_id = pu.object_id
		AND pu.attribute_key = 'url'
		AND pu.attribute_archived = 0
	)
	LEFT JOIN attributes AS pn ON(
		p.object_id = pn.object_id
		AND pn.attribute_key = 'name'
		AND pn.attribute_archived = 0
	)
	LEFT JOIN attributes AS fu ON(
		f.object_id = fu.object_id
		AND fu.attribute_key = 'url'
		AND fu.attribute_archived = 0
	)
WHERE
	p.object_type = 'CoveragePublication'
	AND f.object_type = 'CoverageFeed'
ORDER BY p.object_id
`

func main() {
	db, err := sql.Open("mysql", "root:@/coverage_db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}

	var (
		pubs              = make(map[int64]skytypes.Pub, 512)
		pId               int64
		pName, pUrl, fUrl string
	)
	for rows.Next() {
		if err := rows.Scan(&pId, &pUrl, &pName, &fUrl); err != nil {
			log.Fatal(err)
		}

		p, ok := pubs[pId]
		if !ok {
			p.URL = pUrl
			p.Title = pName
		}
		p.Feeds = append(p.Feeds, fUrl)
		pubs[pId] = p
	}

	for _, pub := range pubs {
		b, err := json.EncodeClientRequest("Publication.Add", pub)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := http.Post(APIURL, "application/json", bytes.NewReader(b))
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		var response interface{}
		if err = json.DecodeClientResponse(resp.Body, &response); err != nil {
			log.Fatal(err)
		}
	}
}
