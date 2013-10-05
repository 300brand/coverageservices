package main

import (
	"flag"
	"git.300brand.com/coverageservices/service"
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/logger"
	"github.com/stvp/go-toml-config"
	"strings"

	_ "git.300brand.com/coverageservices/Article"
	_ "git.300brand.com/coverageservices/Social"
	_ "git.300brand.com/coverageservices/Stats"
	_ "git.300brand.com/coverageservices/StorageReader"
	_ "git.300brand.com/coverageservices/StorageWriter"
)

var (
	configFile     = flag.String("config", "/etc/coverage.yaml", "Config file location")
	showConfig     = flag.Bool("showconfig", false, "Show configuration and exit")
	gearmanServers = config.String("gearman.servers", ":4730")
)

func main() {
	flag.Parse()

	if err := config.Parse(*configFile); err != nil {
		logger.Error.Fatal(err)
	}

	if *showConfig {
		return
	}

	addrs := strings.Split(*gearmanServers, ",")
	server := disgo.NewServer(addrs...)

	for name, s := range service.GetServices() {
		logger.Info.Printf("Registering service: %s", name)
		server.RegisterName(name, s)

		client := disgo.NewClient(addrs...)
		if err := s.Start(client); err != nil {
			logger.Error.Fatal("Failed to start %s: %s", name, err)
		}
		//defer client.Close()
	}
	return
	server.Serve()
}
