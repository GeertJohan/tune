package main

import (
	"time"
)

// Clock is a custom time clock.
// Most users won't run ntpd or some thing simliar.
// Since the AudioAddict servers' time might not be in sync with the client machine,
// we need a custom clock.
type Clock struct {
	t time.Time

	chGet chan chan time.Time
}

// NewClock creates a new Clock and assume t is the current time.
func NewClock(t time.Time) *Clock {
	c := &Clock{
		t:     t,
		chGet: make(chan chan time.Time),
	}
	go c.run()
	return c
}

func (c *Clock) run() {
	sec := 1 * time.Second
	ticker := time.NewTicker(sec)
	for {
		select {
		case <-ticker.C: // TODO: does this need buffering for more accuracy?
			c.t = c.t.Add(sec)
		case retCh := <-c.chGet:
			retCh <- c.t
		}
	}
}

// Now returns the current time according to clock
func (c *Clock) Now() time.Time {
	retCh := make(chan time.Time)
	c.chGet <- retCh
	t := <-retCh
	return t
}
