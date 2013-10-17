package main

import (
	"flag"
	"fmt"
	"github.com/300brand/coverageservices/service"
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/go-toml-config"
	"github.com/jbaikge/logger"
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
)

var (
	configFile     = flag.String("config", "config.toml", "Config file location")
	showConfig     = flag.Bool("showconfig", false, "Show configuration and exit")
	gearmanServers = config.String("gearman.servers", ":4730")
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

	// Prepare for Gearman
	addrs := strings.Split(*gearmanServers, ",")
	server := disgo.NewServer(addrs...)

	for name, s := range service.GetServices() {
		logger.Info.Printf("Registering service: %s", name)
		server.RegisterName(name, s)

		client := disgo.NewClient(addrs...)
		if err := s.Start(client); err != nil {
			logger.Error.Fatal("Failed to start %s: %s", name, err)
		}
		defer client.Close()
	}
	//return
	server.Serve()
}
