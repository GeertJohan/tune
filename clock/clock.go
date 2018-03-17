package clock

import (
	"time"
)

// Clock is a custom time clock.
// Most users won't run ntpd or some thing simliar.
// Since the AudioAddict servers' time might not be in sync with the client machine,
// we need a custom clock to stay in sync with the server time.
// TODO: Go 1.8(?) introduced monotonic clock in time.Time.
// Use the walltime from server together with a local monotonic time and apply the diff based on local monotonic time.
type Clock struct {
	t time.Time

	chGet chan chan time.Time
}

// New creates a new Clock and assume t is the current time.
func New(t time.Time) *Clock {
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
		case <-ticker.C:
			// TODO: this probably needs buffering for more accuracy? time.Ticker does not accomodate for read-lag larger than the interval duration (1 sec).
			// https://play.golang.org/p/Ixmx_yvb35
			// Or don't use a Ticker like this, but calculate offset with machine time.
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
