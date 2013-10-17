package Stats

import (
	"fmt"
	"github.com/300brand/coverageservices/service"
	"github.com/300brand/coverageservices/types"
	"github.com/jbaikge/disgo"
	"github.com/jbaikge/go-toml-config"
	"github.com/jbaikge/logger"
	"github.com/jbaikge/statsd"
	"os"
	"strings"
	"time"
)

type Stats struct {
	client *disgo.Client
	stats  *statsd.Client
}

var _ service.Service = new(Stats)

var (
	cfgRate      = config.Float64("Stats.rate", 1)
	cfgReconnect = config.Duration("Stats.reconnect", time.Minute)
	cfgStatsd    = config.String("Stats.statsd", "127.0.0.1:8125")
)

func init() {
	service.Register("Stats", new(Stats))
}

// Funcs required by Service

func (s *Stats) Start(client *disgo.Client) (err error) {
	s.client = client

	go func(addr string) {
		for {
			logger.Trace.Printf("Connecting to %s", addr)
			newConn, err := statsd.Dial(addr)
			if err != nil {
				logger.Error.Printf("Could not connect to Statsd Server: %s", err)
				<-time.After(3 * time.Second)
				continue
			}
			// Swap connections, close out the old one and start using the fresh
			// connection
			oldConn := s.stats
			s.stats = newConn
			if oldConn != nil {
				oldConn.Close()
			}
			<-time.After(*cfgReconnect)
		}
	}(*cfgStatsd)
	return
}

// Service funcs

func (s *Stats) Completed(stat *types.Stat, out *disgo.NullType) (err error) {
	base := statBase(stat)
	s.stats.Increment(statJoin(base, "Calls"), 1, *cfgRate)
	if stat.Error != nil {
		s.stats.Increment(statJoin(base, "Errors"), 1, *cfgRate)
	}
	s.stats.Duration(statJoin(base, "Duration"), stat.Duration, *cfgRate)
	return
}

func (s *Stats) Decrement(stat *types.Stat, out *disgo.NullType) (err error) {
	return s.stats.Decrement(statBase(stat), stat.Count, *cfgRate)
}

func (s *Stats) Gauge(stat *types.Stat, out *disgo.NullType) (err error) {
	return s.stats.Gauge(statBase(stat), stat.Count, *cfgRate)
}

func (s *Stats) Increment(stat *types.Stat, out *disgo.NullType) (err error) {
	return s.stats.Increment(statBase(stat), stat.Count, *cfgRate)
}

func (s *Stats) Resources(stat *types.Stat, out *disgo.NullType) (err error) {
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
		s.stats.Gauge(statJoin(base, suffix), value, *cfgRate)
	}
	s.stats.Increment(statJoin(base, "Heartbeat"), 1, *cfgRate)
	return
}

func (s *Stats) Duration(stat *types.Stat, out *disgo.NullType) (err error) {
	return s.stats.Duration(statBase(stat), stat.Duration, *cfgRate)
}

// Support funcs

func statBase(stat *types.Stat) (s string) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = fmt.Sprintf("Unknown-%d", os.Getegid())
	}
	if stat.Name == "" {
		s = fmt.Sprintf("%s", hostname)
	} else {
		s = fmt.Sprintf("%s.%s", stat.Name, hostname)
	}
	return
}

func statJoin(paths ...string) string {
	return strings.Join(paths, ".")
}
