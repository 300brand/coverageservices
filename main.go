package main

import (
	"flag"
	"fmt"
	"git.300brand.com/coverageservices/service"
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/logger"
	"launchpad.net/goyaml"

	_ "git.300brand.com/coverageservices/Stats"
	_ "git.300brand.com/coverageservices/StorageReader"
)

var (
	configFile = flag.String("config", "/etc/coverage.yaml", "Config file location")
	showConfig = flag.Bool("showconfig", false, "Show configuration and exit")
)

func main() {
	flag.Parse()
	service.InitConfig(*configFile)
	// Flag-overrides go here
	Config := service.Config

	if *showConfig {
		data, _ := goyaml.Marshal(&Config)
		fmt.Print(string(data))
		return
	}

	server := disgo.NewServer(Config.Gearman.Servers...)
	client := disgo.NewClient(Config.Gearman.Servers...)
	defer client.Close()

	for name, s := range service.GetServices() {
		if b := service.Config.Personalities[name]; !*b {
			logger.Debug.Printf("Service disabled: %s", name)
			continue
		}

		logger.Info.Printf("Registering service: %s", name)
		server.RegisterName(name, s)
		if err := s.Start(client); err != nil {
			logger.Error.Fatal("Failed to start %s: %s", name, err)
		}
	}

	server.Serve()
}
