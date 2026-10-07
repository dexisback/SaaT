// we cannot run real NTP in this repo, because all nodes live on the same machine (i.e local machine running the repo (my laptop), and we could only simulate a distributed environment. there's no real commn lag. )
// although ntp could be used like this
// a small program to showcase how far your machine's own clock is from the NTP server, and RTT
package main

import (
	"fmt"

	"github.com/beevik/ntp"
)

func main() {
	response, err := ntp.Query("pool.ntp.org")
	if err != nil {
		panic(err)
	}
	//else:
	fmt.Println("my laptop ki clock's offset from NTP: ", response.ClockOffset)
	fmt.Println("round trip time to server: ", response.RTT)
}
