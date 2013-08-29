package skynetstats

import (
	"git.300brand.com/coverage/skytypes"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"runtime"
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

func Count(name string, count int) {
	stat := skytypes.Stat{
		Config: sc,
		Name:   name,
		Count:  count,
	}
	c.SendOnce(nil, "Increment", stat, skytypes.Null)
}

func Duration(name string, d time.Duration) {
	stat := skytypes.Stat{
		Config:   sc,
		Name:     name,
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
