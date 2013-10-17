package sendstat

import (
	"github.com/300brand/coverageservices/types"
	"github.com/jbaikge/disgo"
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
	stat := types.Stat{
		Config:   sc,
		Name:     m,
		Duration: time.Duration(d),
		Error:    err,
	}
	c.SendOnce(nil, "Completed", stat, disgo.Null)
}

func Count(count int, name ...string) {
	stat := types.Stat{
		Config: sc,
		Name:   strings.Join(name, "."),
		Count:  count,
	}
	c.SendOnce(nil, "Increment", stat, disgo.Null)
}

func Duration(d time.Duration, name ...string) {
	_, file, line, ok := runtime.Caller(2)
	stat := types.Stat{
		Config:   sc,
		Name:     strings.Join(name, "."),
		Duration: d,
	}
	c.SendOnce(nil, "Duration", stat, disgo.Null)
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
	stat := types.Stat{
		Config:     sc,
		Goroutines: runtime.NumGoroutine(),
	}
	runtime.ReadMemStats(&stat.Mem)
	c.SendOnce(nil, "Resources", stat, disgo.Null)
}
