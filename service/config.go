package service

import (
	"github.com/jbaikge/logger"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
)

type cfgGearman struct{ Servers []string }

var Config = struct {
	Gearman       cfgGearman
	Personalities map[string]*bool
	Services      map[string]interface{}
}{
	cfgGearman{[]string{":4730"}},
	make(map[string]*bool),
	make(map[string]interface{}),
}

func InitConfig(filename string) {
	// // Prefill the personalities
	// for name := range services {
	// 	var t *bool
	// 	Config.Personalities[name] = t
	// 	flag.BoolVar(t, "personality."+name, false, "Disables")
	// }
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
