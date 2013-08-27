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
		Config: sc,
		Name:   m,
		Nanos:  d,
		Error:  err,
	}
	c.SendOnce(nil, "Completed", stat, skytypes.Null)
}

func Report(sc *skynet.ServiceConfig) {
	stat := skytypes.Stat{
		Config:     sc,
		Goroutines: runtime.NumGoroutine(),
	}
	runtime.ReadMemStats(&stat.Mem)
	c.SendOnce(nil, "Resources", stat, skytypes.Null)
}

func Start(serviceConfig *skynet.ServiceConfig, statsClient *client.ServiceClient) {
	c = statsClient
	sc = serviceConfig

	ticker = time.NewTicker(10 * time.Second)
	go func(sc *skynet.ServiceConfig) {
		for {
			select {
			case <-ticker.C:
				Report(sc)
			case <-chStop:
				return
			}
		}
	}(sc)
}

func Stop() {
	ticker.Stop()
	chStop <- true
}
