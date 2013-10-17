package main

import (
	"bytes"
	"database/sql"
	jsonenc "encoding/json"
	"flag"
	"github.com/300brand/coverageservices/skytypes"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/rpc/json"
	"log"
	"net/http"
	"os"
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

var Config = struct {
	Export, Save *bool
	Import       *string
}{
	Export: flag.Bool("export", false, "Print results as JSON to STDOUT"),
	Save:   flag.Bool("save", false, "Send processed publications to coverage network via JSON RPC API"),
	Import: flag.String("import", "", "JSON file to import (used in place of MySQL data)"),
}

func Export(pubs []skytypes.Pub) {
	b, err := jsonenc.MarshalIndent(pubs, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(b)
}

func JSONImport(file string) (pubs []skytypes.Pub, err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	dec := jsonenc.NewDecoder(f)
	dec.UseNumber()
	err = dec.Decode(&pubs)
	return
}

func MySQLImport() (pubs []skytypes.Pub, err error) {
	pubs = make([]skytypes.Pub, 0, 512)
	var (
		pubMap            = make(map[int64]skytypes.Pub, 512)
		pId               int64
		pName, pUrl, fUrl string
	)
	db, err := sql.Open("mysql", "root:@/coverage_db")
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&pId, &pUrl, &pName, &fUrl); err != nil {
			return
		}

		p, ok := pubMap[pId]
		if !ok {
			p.URL = pUrl
			p.Title = pName
		}
		p.Feeds = append(p.Feeds, fUrl)
		pubMap[pId] = p
	}
	for _, pub := range pubMap {
		pubs = append(pubs, pub)
	}
	return
}

func Save(pubs []skytypes.Pub) {
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

func main() {
	flag.Parse()

	var (
		pubs []skytypes.Pub
		err  error
	)

	if file := *Config.Import; file != "" {
		pubs, err = JSONImport(file)
	} else {
		pubs, err = MySQLImport()
	}

	if err != nil {
		log.Fatal(err)
	}

	if *Config.Export {
		Export(pubs)
	}
	if *Config.Save {
		Save(pubs)
	}
}
