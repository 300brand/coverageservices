package service

import (
	"github.com/jbaikge/disgo"
)

type Service interface {
	// Provides a disgo Client for the service to use
	DisgoClient(client *disgo.Client)
	// Called after all services are registered and before the disgo server
	// starts
	Start() error
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
}

func GetServices() map[string]Service {
	return services
}
