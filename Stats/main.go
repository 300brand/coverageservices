package main

import (
	"fmt"
	"git.300brand.com/coverage/config"
	"git.300brand.com/coverage/skytypes"
	"git.300brand.com/coverageservices/skynetstats"
	"github.com/cyberdelia/statsd"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"log"
	"strings"
	"time"
)

type Service struct {
	Config *skynet.ServiceConfig
}

const (
	ServiceName = "Stats"
	Rate        = float64(1)
)

var (
	_     service.ServiceDelegate = &Service{}
	stats *statsd.Client
	Stats *client.ServiceClient
)

// Funcs required for ServiceDelegate

func (s *Service) MethodCalled(m string) {}

func (s *Service) MethodCompleted(m string, d int64, err error) {}

func (s *Service) Registered(service *service.Service) {
	log.Printf("Connecting to %s", config.StatsdServer.Address)
	var err error
	if stats, err = statsd.Dial(config.StatsdServer.Address); err != nil {
		log.Fatalf("Could not connect to Statsd Server: %s", err)
	}
	skynetstats.Start(s.Config, Stats)
}

func (s *Service) Started(service *service.Service) {}

func (s *Service) Stopped(service *service.Service) {
	stats.Close()
}

func (s *Service) Unregistered(service *service.Service) {
	skynetstats.Stop()
}

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
	base := statJoin(statBase(stat), "Completed")
	stats.Increment(statJoin(base, "Calls"), 1, Rate)
	if stat.Error != nil {
		stats.Increment(statJoin(base, "Errors"), 1, Rate)
	}
	stats.Timing(statJoin(base, "Duration"), int(time.Duration(stat.Nanos)/time.Millisecond), Rate)
	return
}

func (s *Service) Resources(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	base := statBase(stat)
	stats.Gauge(statJoin(base, "Alloc"), int(stat.Mem.Alloc), Rate)
	stats.Gauge(statJoin(base, "TotalAlloc"), int(stat.Mem.TotalAlloc), Rate)
	stats.Gauge(statJoin(base, "Mallocs"), int(stat.Mem.Mallocs), Rate)
	stats.Gauge(statJoin(base, "HeapAlloc"), int(stat.Mem.HeapAlloc), Rate)
	stats.Gauge(statJoin(base, "HeapSys"), int(stat.Mem.HeapSys), Rate)
	stats.Gauge(statJoin(base, "HeapInuse"), int(stat.Mem.HeapInuse), Rate)
	stats.Gauge(statJoin(base, "StackSys"), int(stat.Mem.StackSys), Rate)
	stats.Gauge(statJoin(base, "StackInuse"), int(stat.Mem.StackInuse), Rate)
	stats.Gauge(statJoin(base, "NumGC"), int(stat.Mem.NumGC), Rate)
	return
}

// Support funcs

func statBase(stat *skytypes.Stat) (s string) {
	if stat.Name == "" {
		s = fmt.Sprintf("%s.%s.%s", stat.Config.Name, stat.Config.Version, stat.Config.Region)
	} else {
		s = fmt.Sprintf("%s.%s.%s.%s", stat.Config.Name, stat.Name, stat.Config.Version, stat.Config.Region)
	}
	return
}

func statJoin(paths ...string) string {
	return strings.Join(paths, ".")
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	cc, _ := skynet.GetClientConfig()
	c := client.NewClient(cc)

	Stats = c.GetService("Stats", "", "", "")

	sc, _ := skynet.GetServiceConfig()
	sc.Name = ServiceName
	sc.Region = "Stats"
	sc.Version = "1"

	s := service.CreateService(&Service{sc}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
