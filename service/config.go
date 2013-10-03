package main

import (
	"git.300brand.com/coverageservices/service"
	"github.com/jbaikge/logger"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
)

type cfgGearman struct{ Servers []string }
type cfgMongo struct{ Servers []string }

var Config = struct {
	Gearman       cfgGearman
	Mongo         cfgMongo
	Personalities map[string]bool
}{
	cfgGearman{[]string{":4730"}},
	cfgMongo{[]string{":4730"}},
	make(map[string]bool),
}

func InitConfig(filename string) {
	// Prefill the personalities
	for name := range service.GetServices() {
		Config.Personalities[name] = true
	}
	// Attempt to read config from filename
	f, err := os.Open(filename)
	if err != nil {
		logger.Warn.Print(err)
		data, err := goyaml.Marshal(&Config)
		if err != nil {
			logger.Error.Fatalf("Something went terribly wrong with the config: %s", err)
		}
		logger.Warn.Printf("Using config (Place in %s):\n%s", filename, data)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err := goyaml.Unmarshal(data, &Config); err != nil {
		logger.Error.Fatalf("There was an error parsing %s: %s", filename, err)
	}
}
