package main

import (
	"fmt"
	"git.300brand.com/coverage/config"
	"git.300brand.com/coverage/skytypes"
	"github.com/cyberdelia/statsd"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/service"
	"log"
	"strings"
	"time"
)

type Service struct{}

const (
	ServiceName = "Stats"
)

var (
	_      service.ServiceDelegate = &Service{}
	client *statsd.Client
)

// Funcs required for ServiceDelegate

func (s *Service) MethodCalled(m string) {}

func (s *Service) MethodCompleted(m string, d int64, err error) {}

func (s *Service) Registered(service *service.Service) {
	log.Printf("Connecting to %s", config.StatsdServer.Address)
	var err error
	if client, err = statsd.Dial(config.StatsdServer.Address); err != nil {
		log.Fatalf("Could not connect to Statsd Server: %s", err)
	}
}

func (s *Service) Started(service *service.Service) {}

func (s *Service) Stopped(service *service.Service) {
	client.Close()
}

func (s *Service) Unregistered(service *service.Service) {}

// Service funcs

func (s *Service) Decrement(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	return
}

func (s *Service) Gauge(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	return
}

func (s *Service) Increment(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	return
}

func (s *Service) Completed(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	log.Printf("Stat Info: %v", stat)
	rate := float64(1)
	base := statJoin(statBase(stat), "MethodCompleted")
	client.Increment(statJoin(base, "count"), 1, rate)
	client.Timing(statJoin(base, "duration"), int(time.Duration(stat.Nanos)/time.Millisecond), rate)
	if stat.Error != nil {
		client.Increment(statJoin(base, "errors"), 1, rate)
	}
	client.Gauge(statJoin(base, "alloc"), int(stat.Mem.Alloc), rate)
	client.Gauge(statJoin(base, "totalAlloc"), int(stat.Mem.TotalAlloc), rate)
	client.Gauge(statJoin(base, "mallocs"), int(stat.Mem.Mallocs), rate)
	client.Gauge(statJoin(base, "heapAlloc"), int(stat.Mem.HeapAlloc), rate)
	client.Gauge(statJoin(base, "heapSys"), int(stat.Mem.HeapSys), rate)
	client.Gauge(statJoin(base, "heapInuse"), int(stat.Mem.HeapInuse), rate)
	client.Gauge(statJoin(base, "stackSys"), int(stat.Mem.StackSys), rate)
	client.Gauge(statJoin(base, "stackInuse"), int(stat.Mem.StackInuse), rate)
	client.Gauge(statJoin(base, "numGC"), int(stat.Mem.NumGC), rate)
	return
}

// Support funcs

func statBase(stat *skytypes.Stat) string {
	return fmt.Sprintf("%s.%s.%s.%s", stat.Config.Name, stat.Name, stat.Config.Version, stat.Config.Region)
}

func statJoin(paths ...string) string {
	return strings.Join(paths, ".")
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	sc, _ := skynet.GetServiceConfig()
	sc.Name = ServiceName
	sc.Region = "Stats"
	sc.Version = "1"

	s := service.CreateService(&Service{}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
