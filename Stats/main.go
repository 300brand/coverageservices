package main

import (
	"fmt"
	"git.300brand.com/coverage/config"
	"git.300brand.com/coverage/skytypes"
	"git.300brand.com/coverageservices/skynetstats"
	"github.com/jbaikge/statsd"
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
	go func(addr string) {
		for {
			log.Printf("Connecting to %s", addr)
			newConn, err := statsd.Dial(addr)
			if err != nil {
				log.Printf("Could not connect to Statsd Server: %s", err)
			}
			// Swap connections, close out the old one and start using the fresh
			// connection
			oldConn := stats
			stats = newConn
			if oldConn != nil {
				oldConn.Close()
			}
			<-time.After(time.Minute)
		}
	}(config.StatsdServer.Address)
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

func (s *Service) Completed(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	base := statBase(stat)
	stats.Increment(statJoin(base, "Calls"), 1, Rate)
	if stat.Error != nil {
		stats.Increment(statJoin(base, "Errors"), 1, Rate)
	}
	stats.Timing(statJoin(base, "Duration"), stat.Duration, Rate)
	return
}

func (s *Service) Decrement(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	return stats.Decrement(statBase(stat), stat.Count, Rate)
}

func (s *Service) Gauge(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	return stats.Gauge(statBase(stat), stat.Count, Rate)
}

func (s *Service) Increment(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	return stats.Increment(statBase(stat), stat.Count, Rate)
}

func (s *Service) Resources(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	base := statBase(stat)
	attr := map[string]int{
		"Alloc":      int(stat.Mem.Alloc),
		"TotalAlloc": int(stat.Mem.TotalAlloc),
		"Mallocs":    int(stat.Mem.Mallocs),
		"HeapAlloc":  int(stat.Mem.HeapAlloc),
		"HeapSys":    int(stat.Mem.HeapSys),
		"HeapInuse":  int(stat.Mem.HeapInuse),
		"StackSys":   int(stat.Mem.StackSys),
		"StackInuse": int(stat.Mem.StackInuse),
		"NumGC":      int(stat.Mem.NumGC),
	}
	for suffix, value := range attr {
		stats.Gauge(statJoin(base, suffix), value, Rate)
	}
	return
}

func (s *Service) Timing(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	return stats.Timing(statBase(stat), stat.Duration, Rate)
}

// Support funcs

func statBase(stat *skytypes.Stat) (s string) {
	if stat.Name == "" {
		s = fmt.Sprintf("%s.%s.%s", stat.Config.Region, stat.Config.Version, stat.Config.Name)
	} else {
		s = fmt.Sprintf("%s.%s.%s.%s", stat.Config.Region, stat.Config.Version, stat.Config.Name, stat.Name)
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
