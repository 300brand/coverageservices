package service

import (
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/go-toml-config"
)

type Service interface {
	// Called after all services are registered and before the disgo server
	// starts
	Start(client *disgo.Client) error
}

type serviceInfo struct {
	Service       Service
	ConfigEnabled *bool
}

var services = make(map[string]serviceInfo)

func (si serviceInfo) Enabled() bool {
	return *si.ConfigEnabled
}

func Register(name string, service Service) {
	if service == nil {
		panic("service: Register service is nil")
	}
	if _, dup := services[name]; dup {
		panic("service: Register called twice for " + name)
	}

	// Add service
	services[name] = serviceInfo{
		Service:       service,
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
