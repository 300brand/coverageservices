package main

import (
	"flag"
	"fmt"
	"git.300brand.com/coverageservices/service"
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/go-toml-config"
	"github.com/jbaikge/logger"
	"os"
	"strings"

	_ "git.300brand.com/coverageservices/Article"
	_ "git.300brand.com/coverageservices/Feed"
	_ "git.300brand.com/coverageservices/Social"
	_ "git.300brand.com/coverageservices/Stats"
	_ "git.300brand.com/coverageservices/StorageReader"
	_ "git.300brand.com/coverageservices/StorageWriter"
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
	return
	server.Serve()
}
