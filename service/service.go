package service

import (
	"flag"
	"github.com/jbaikge/disgo"
)

type Service interface {
	// Called upon registration
	ConfigOptions() interface{}
	// Called after all services are registered and before the disgo server
	// starts
	Start(client *disgo.Client) error
}

var services = make(map[string]Service)

func Register(name string, service Service) {
	if service == nil {
		panic("service: Register service is nil")
	}
	if _, dup := services[name]; dup {
		panic("service: Register called twice for " + name)
	}
	services[name] = service

	// Supplies the service configuration pointer to create a dynamic config
	// file. Each service has it's own subsection under services
	Config.Services[name] = service.ConfigOptions()

	t := new(bool)
	Config.Personalities[name] = t
	flag.BoolVar(t, "personality."+name, false, "Enables a personality (service)")
}

func GetServices() map[string]Service {
	return services
}
