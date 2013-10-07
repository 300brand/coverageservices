package service

import (
	"flag"
	"fmt"
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/go-toml-config"
)

type Service interface {
	// Called after all services are registered and before the disgo server
	// starts
	Start(client *disgo.Client) error
}

type enableDisable string

type serviceInfo struct {
	Service       Service
	FlagEnabled   *enableDisable
	ConfigEnabled *bool
}

var services = make(map[string]serviceInfo)

var _ flag.Value = new(enableDisable)

func (ed *enableDisable) Set(v string) (err error) {
	switch v {
	case "enable", "disable", "":
		*ed = enableDisable(v)
	default:
		return fmt.Errorf("Invalid value \"%s\"", v)
	}
	return
}

func (ed *enableDisable) String() string { return string(*ed) }

func (si serviceInfo) Enabled() bool {
	if si.FlagEnabled.String() == "disable" {
		return false
	}
	return si.FlagEnabled.String() == "enable" || *si.ConfigEnabled
}

func Register(name string, service Service) {
	if service == nil {
		panic("service: Register service is nil")
	}
	if _, dup := services[name]; dup {
		panic("service: Register called twice for " + name)
	}

	// Set up flag
	fe := new(enableDisable)
	//flag.Var(fe, name, "Valid values: \"enable\", \"disable\", or \"\" (uses value in config file)")

	// Add service
	services[name] = serviceInfo{
		Service:       service,
		FlagEnabled:   fe,
		ConfigEnabled: config.Bool(name+".enabled", false),
	}

}

func GetServices() (m map[string]Service) {
	m = make(map[string]Service)
	for name := range services {
		if s := services[name]; s.Enabled() {
			m[name] = s.Service
		}
	}
	return
}
