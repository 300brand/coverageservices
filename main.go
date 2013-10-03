package main

import (
	"flag"
	"git.300brand.com/coverageservices/service"
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/logger"
)

var (
	configFile = flag.String("config", "/etc/coverage.yaml", "Config file location")
)

func main() {
	flag.Parse()
	InitConfig(*configFile)

	server := disgo.NewServer(Config.Gearman.Servers...)
	client := disgo.NewClient(Config.Gearman.Servers...)
	defer client.Close()

	for name, service := range service.GetServices() {
		logger.Info.Printf("Registering service: %s", name)
		service.DisgoClient(client)
	}

	server.Serve()
}
