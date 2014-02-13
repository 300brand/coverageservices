package main

import (
	"flag"
	"fmt"
	"github.com/300brand/coverageservices/service"
	"github.com/300brand/disgo"
	"github.com/300brand/go-toml-config"
	"github.com/300brand/logger"
	"net/http"
	"os"

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
	configFile      = flag.String("config", "config.toml", "Config file location")
	showConfig      = flag.Bool("showconfig", false, "Show configuration and exit")
	gearmanServers  = config.String("gearman.servers", ":4730")
	beanstalkServer = config.String("beanstalk.server", "127.0.0.1:11300")
	pprofListen     = config.String("pprof.listen", ":6060")
)

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

	server, err := disgo.NewServer(*beanstalkServer)
	if err != nil {
		logger.Error.Fatalf("Error connecting server: %s", err)
	}

	client, err := disgo.NewClient(*beanstalkServer)
	if err != nil {
		logger.Error.Fatalf("Error connecting client: %s", err)
	}

	for name, s := range service.GetServices() {
		logger.Info.Printf("Registering service: %s", name)
		server.RegisterName(name, s)

		if err := s.Start(client); err != nil {
			logger.Error.Fatal("Failed to start %s: %s", name, err)
		}
		defer client.Close()
	}

	logger.Error.Fatal(server.Serve())
}
