package main

import (
	"fmt"
	"git.300brand.com/coverage/config"
	"git.300brand.com/coverageservices/skynetstats"
	"git.300brand.com/coverageservices/skytypes"
	"github.com/jbaikge/logger"
	"github.com/jbaikge/statsd"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"os"
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

func (s *Service) MethodCompleted(m string, d int64, err error) {
	if err != nil {
		logger.Error.Printf("Stats.%s error: %s", m, err)
	}
}

func (s *Service) Registered(service *service.Service) {
	go func(addr string) {
		for {
			logger.Debug.Printf("Connecting to %s", addr)
			newConn, err := statsd.Dial(addr)
			if err != nil {
				logger.Error.Printf("Could not connect to Statsd Server: %s", err)
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
	stats.Duration(statJoin(base, "Duration"), stat.Duration, Rate)
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
		"Goroutines": stat.Goroutines,
	}
	for suffix, value := range attr {
		stats.Gauge(statJoin(base, suffix), value, Rate)
	}
	stats.Increment(statJoin(base, "Heartbeat"), 1, Rate)
	return
}

func (s *Service) Duration(ri *skynet.RequestInfo, stat *skytypes.Stat, out *skytypes.NullType) (err error) {
	return stats.Duration(statBase(stat), stat.Duration, Rate)
}

// Support funcs

func statBase(stat *skytypes.Stat) (s string) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = fmt.Sprintf("Unknown-%d", os.Getegid())
	}
	if stat.Name == "" {
		s = fmt.Sprintf("%s.%s.%s.%s", stat.Config.Region, stat.Config.Version, stat.Config.Name, hostname)
	} else {
		s = fmt.Sprintf("%s.%s.%s.%s.%s", stat.Config.Region, stat.Config.Version, stat.Config.Name, stat.Name, hostname)
	}
	return
}

func statJoin(paths ...string) string {
	return strings.Join(paths, ".")
}

// Main

func main() {
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
