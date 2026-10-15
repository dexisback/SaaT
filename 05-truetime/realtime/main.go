package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/beevik/ntp"
)

type Interval struct {
	Earliest time.Time
	Latest   time.Time
}

func (interval Interval) String() string {
	uncertainty := interval.Latest.Sub(interval.Earliest)
	return fmt.Sprintf("[%s, %s] (uncertainty: %v)",
		interval.Earliest.Format("15:04:05.000"),
		interval.Latest.Format("15:04:05.000"),
		uncertainty)
}

// clock queries a real NTP server and builds an we build an uncertaininty interval from the round-trip time
type Clock struct {
	lock       sync.Mutex
	ntpServer  string
	lastResult *ntp.Response // we gonna use this later for caching so our server doesnt get hammered
}

func NewClock(ntpServer string) *Clock {
	return &Clock{ntpServer: ntpServer}
}

// sync queries the NTP server and caches the result (because NTP server would think we're trying to ddos if multiple reqs)
func (clock *Clock) Sync() error {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	response, err := ntp.Query(clock.ntpServer)
	if err != nil {
		return err
	}
	clock.lastResult = response
	return nil
}

//now returns the uncertainity interval.
// The midpoint is our local time adjusted by the NTP offset.
//the window is the roundtrip time
//at most one full round-trip before we got it

func (clock *Clock) Now() Interval {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	if clock.lastResult == nil {
		//case where we havent yet synced with the NTP
		now := time.Now()
		return Interval{
			Earliest: now.Add(-1 * time.Second), //because we dont know the real offset or RTT, we cant trust the machine's clock. so right now returning +-1 second is a safe fallback
			Latest:   now.Add(1 * time.Second),
		}
	}
	//else if synced:
	correctedTime := time.Now().Add(clock.lastResult.ClockOffset)
	halfRoundTrip := clock.lastResult.RTT / 2

	return Interval{
		Earliest: correctedTime.Add(-halfRoundTrip),
		Latest:   correctedTime.Add(halfRoundTrip),
	}

}

// Stamp picks the Latest end of the interval as the commit timestamp.
func (clock *Clock) Stamp() time.Time {
	return clock.Now().Latest
}

// CommitWait blocks until our Earliest is past the timestamp,
// meaning we are certain the realtime has moved past it
func (clock *Clock) CommitWait(timestamp time.Time) time.Duration {
	startedWaiting := time.Now()
	for {
		if clock.Now().Earliest.After(timestamp) {
			return time.Since(startedWaiting)
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func main() {
	servers := []string{
		"time.cloudflare.com",
		"time.google.com",
		"pool.ntp.org",
	}

	fmt.Println("=== querying real NTP servers ===")
	fmt.Println()

	clocks := make([]*Clock, len(servers))
	for index, server := range servers {
		clocks[index] = NewClock(server)
		err := clocks[index].Sync()
		if err != nil {
			fmt.Printf("%-25s  could not reach: %v\n", server, err)
			continue
		}
		interval := clocks[index].Now()
		fmt.Printf("%-25s  %s\n", server, interval)
	}

	fmt.Println()
	fmt.Println("=== simulating commit wait with real uncertainty ===")
	fmt.Println()

	// pick the first clock that synced
	var activeClock *Clock
	var activeServer string
	for index, clock := range clocks {
		if clock.lastResult != nil {
			activeClock = clock
			activeServer = servers[index]
			break
		}
	}

	if activeClock == nil {
		fmt.Println("no NTP server reachable, cannot continue")
		return
	}

	fmt.Printf("using %s\n", activeServer)
	fmt.Printf("clock offset from true time: %v\n", activeClock.lastResult.ClockOffset)
	fmt.Printf("round-trip time to server:   %v\n", activeClock.lastResult.RTT)
	fmt.Println()

	timestamp := activeClock.Stamp()
	fmt.Println("transaction stamped at:", timestamp.Format("15:04:05.000000"))
	fmt.Println("waiting out uncertainty window...")

	waited := activeClock.CommitWait(timestamp)
	fmt.Printf("commit wait complete, waited %v\n", waited)
	fmt.Println()
	fmt.Println("any node that syncs after this point is guaranteed")
	fmt.Println("to see our transaction as being in the past.")
}
