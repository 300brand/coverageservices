package main

import (
	"errors"
	"git.300brand.com/coverageservices/skytypes"
	"sync/atomic"
	"time"
)

type Ticker struct {
	F       func() error
	Running int64
	Start   chan bool
	Stop    chan bool
	Tick    time.Duration
	Ticker  *time.Ticker
}

func NewTicker(f func() error, d time.Duration) *Ticker {
	return &Ticker{
		F:      f,
		Start:  make(chan bool),
		Stop:   make(chan bool),
		Tick:   d,
		Ticker: &time.Ticker{},
	}
}

func (t *Ticker) ProcessCommand(cmd *skytypes.ClockCommand) (err error) {
	switch cmd.Command {
	case "once":
		return t.F()
	case "start":
		t.Start <- true
	case "stop":
		t.Stop <- true
	default:
		err = errors.New("Unknown command: " + cmd.Command)
	}
	return
}

func (t *Ticker) Run() {
	for {
		select {
		case <-t.Start:
			t.Ticker = time.NewTicker(t.Tick)
		case <-t.Stop:
			t.Ticker.Stop()
		case <-t.Ticker.C:
			go func(t *Ticker) {
				atomic.AddInt64(&t.Running, 1)
				t.F()
				atomic.AddInt64(&t.Running, -1)
			}(t)
		}
	}
}
