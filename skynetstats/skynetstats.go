package skynetstats

import (
	"git.300brand.com/coverageservices/skytypes"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"runtime"
	"strings"
	"time"
)

var (
	ticker *time.Ticker
	chStop = make(chan bool)
	c      *client.ServiceClient
	sc     *skynet.ServiceConfig
)

func Completed(m string, d int64, err error) {
	stat := skytypes.Stat{
		Config:   sc,
		Name:     m,
		Duration: time.Duration(d),
		Error:    err,
	}
	c.SendOnce(nil, "Completed", stat, skytypes.Null)
}

func Count(count int, name ...string) {
	stat := skytypes.Stat{
		Config: sc,
		Name:   strings.Join(name, "."),
		Count:  count,
	}
	c.SendOnce(nil, "Increment", stat, skytypes.Null)
}

func Duration(d time.Duration, name ...string) {
	stat := skytypes.Stat{
		Config:   sc,
		Name:     strings.Join(name, "."),
		Duration: d,
	}
	c.SendOnce(nil, "Duration", stat, skytypes.Null)
}

func Start(serviceConfig *skynet.ServiceConfig, statsClient *client.ServiceClient) {
	c = statsClient
	sc = serviceConfig

	ticker = time.NewTicker(10 * time.Second)
	go func(sc *skynet.ServiceConfig) {
		for {
			select {
			case <-ticker.C:
				report(sc)
			case <-chStop:
				return
			}
		}
	}(serviceConfig)
}

func Stop() {
	ticker.Stop()
	chStop <- true
}

func report(sc *skynet.ServiceConfig) {
	stat := skytypes.Stat{
		Config:     sc,
		Goroutines: runtime.NumGoroutine(),
	}
	runtime.ReadMemStats(&stat.Mem)
	c.SendOnce(nil, "Resources", stat, skytypes.Null)
}
