package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/300brand/coverageservices/service"
	"github.com/300brand/disgo"
	"github.com/300brand/go-toml-config"
	"github.com/300brand/logger"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"strings"

	_ "github.com/300brand/coverageservices/Article"
	_ "github.com/300brand/coverageservices/Feed"
	_ "github.com/300brand/coverageservices/Manager"
	_ "github.com/300brand/coverageservices/Publication"
	_ "github.com/300brand/coverageservices/Search"
	_ "github.com/300brand/coverageservices/Social"
	_ "github.com/300brand/coverageservices/Stats"
	_ "github.com/300brand/coverageservices/StorageReader"
	_ "github.com/300brand/coverageservices/StorageWriter"
	_ "github.com/300brand/coverageservices/WebAPI"
	_ "net/http/pprof"
)

var (
	configFile     = flag.String("config", "config.toml", "Config file location")
	showConfig     = flag.Bool("showconfig", false, "Show configuration and exit")
	disgoListen    = config.String("disgo.listen", "127.0.0.1:10000")
	disgoBroadcast = config.String("disgo.broadcast", "")
	etcdServers    = config.String("disgo.etcd.servers", "127.0.0.1:4001")
	pprofListen    = config.String("pprof.listen", ":6060")
)

func init() {
	gob.Register(new(bson.M))
	gob.Register(new(bson.D))
	gob.Register(new(bson.ObjectId))
	gob.Register([]bson.ObjectId{})
}

func main() {
	// Parse flags and config
	flag.Usage = func() {
		fmt.Printf("Usage for %s:\n(use -showconfig to see parsed values)\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if err := config.Parse(*configFile); err != nil {
		logger.Error.Fatal(err)
	}

	if err := applyLogSettings(); err != nil {
		fmt.Printf("Error with logging settings: %s\n", err)
		os.Exit(1)
	}

	// Show values
	if *showConfig {
		fmt.Printf("%-32s%-32s%s\n\n", "Directive", "Value", "Default")
		config.VisitAll(func(f *flag.Flag) {
			fmt.Printf("%-32s%-32v%s\n", f.Name, f.Value, f.DefValue)
		})
		return
	}

	// Make pprof available
	go func() {
		logger.Error.Println(http.ListenAndServe(*pprofListen, nil))
	}()

	server, err := disgo.NewServer(strings.Split(*etcdServers, ","), *disgoBroadcast)
	if err != nil {
		logger.Error.Fatalf("Error connecting server: %s", err)
	}

	client, err := disgo.NewClient(strings.Split(*etcdServers, ","))
	if err != nil {
		logger.Error.Fatalf("Error connecting client: %s", err)
	}

	var haveServices bool
	for name, s := range service.GetServices() {
		logger.Info.Printf("Registering service: %s", name)
		if err := server.RegisterName(name, s); err != nil {
			logger.Warn.Printf("Error registering services for %s", name)
		} else {
			haveServices = true
		}

		if err := s.Start(client); err != nil {
			logger.Error.Fatal("Failed to start %s: %s", name, err)
		}
		defer client.Close()
	}

	if haveServices {
		// Run DisGo server!
		logger.Error.Fatal(server.Serve(*disgoListen))
	} else {
		// This only happens when just the WebAPI service is running (no
		// exported service methods)
		select {}
	}
}
