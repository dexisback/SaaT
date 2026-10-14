//why?
//physical need:(yes) because physical clocks stay close to real time but break causality, because skew can stamp an effect before a cause
//lamport need: lamport clock keeps causality but loses real time, so its numbers mean nothing as clock times
//vector need: with a lamport clock, causality only worked in 1 direction(no pun). as in, if A->B => Lamport(A) < Lamport(B). But the reverse was NOT true. If Lamport(A) < Lamport(B) !=> doesnt mean A-> B
// a vector clock sits right between lamport clock and full causality graphs. instead of keeping a single counter, every node keeps an array(a vector, so an information of each other).
//HLC (keeps both) -- its timestamp has two parts :
//1: walltime: a physical time in milliseconds, through it can run slightly ahead of the node's real clock. but ok ig
//2: counter: breaks ties when two events share the same WallTime

package main

import (
	"fmt"
	"sync"
)

//HLC timestamp = [walltime, counter]

type TimeStamp struct {
	WallTime int64
	Counter  int
}

// string makes Println show a timestamp as (Wall, Counter)
func (timestamp TimeStamp) String() string {
	return fmt.Sprintf("(%d, %d)", timestamp.WallTime, timestamp.Counter)
}

// IsBefore compares WallTime first, then Counter:
func (timestamp TimeStamp) IsBefore(other TimeStamp) bool {
	if timestamp.WallTime != other.WallTime {
		return timestamp.WallTime < other.WallTime

	}
	//else:
	return timestamp.Counter < other.Counter //return jo iski boolean value hogi
}

func whosTheBiggerPokemon(first int64, second int64) int64 {
	if first > second {
		return first
	}
	//else:
	return second
}

// clock is one node's hlc(like other clocks, independent per node hai)
type Clock struct {
	lock     sync.Mutex
	wallTime int64
	counter  int
}

func NewClock() *Clock {
	return &Clock{wallTime: 0, counter: 0}
}

// defining locakevent:
func (clock *Clock) LocalEvent(physicalTimeNow int64) TimeStamp {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	previousWallTime := clock.wallTime
	clock.wallTime = whosTheBiggerPokemon(previousWallTime, physicalTimeNow)

	if clock.wallTime == previousWallTime {
		clock.counter++ //physical time hasnt moved past us, so we use counter instead

	} else {
		clock.counter = 0 //phy time alr moved ahead (like she did), so start fresh
	}

	return TimeStamp{WallTime: clock.wallTime, Counter: clock.counter}

}

// writing sendMessage: (sending is also an event. The returned timestamp is attached to the message)
func (clock *Clock) SendMessage(physicalTimeNow int64) TimeStamp {
	return clock.LocalEvent(physicalTimeNow)
}

// receive: take the biggest walltime of the 3, then fix the counter
func (clock *Clock) ReceiveMessage(incomingTimeStamp TimeStamp, physicalTimeNow int64) TimeStamp {
	clock.lock.Lock()
	defer clock.lock.Unlock()

	previousWallTime := clock.wallTime
	clock.wallTime = whosTheBiggerPokemon(whosTheBiggerPokemon(previousWallTime, incomingTimeStamp.WallTime), physicalTimeNow)

	sameAsMine := clock.wallTime == previousWallTime
	sameAsMessage := clock.wallTime == incomingTimeStamp.WallTime

	if sameAsMessage && sameAsMine {
		//all tied: then take the bigger counter, and add +1
		if incomingTimeStamp.Counter > clock.counter {
			clock.counter = incomingTimeStamp.Counter
		}
		clock.counter++

	} else if sameAsMine {
		clock.counter++
	} else if sameAsMessage {
		clock.counter = incomingTimeStamp.Counter + 1
	} else {
		clock.counter = 0 //our own physical clock was the newest, so we start fresh

	}

	return TimeStamp{WallTime: clock.wallTime, Counter: clock.counter}

}

// main tells the same story again (3 nodes, same everything )
func main() {
	clockOfNode0 := NewClock()
	clockOfNode1 := NewClock()
	clockOfNode2 := NewClock()

	// Node 0 does something locally, then sends a message to node 1
	node0Local := clockOfNode0.LocalEvent(1000)
	node0Send := clockOfNode0.SendMessage(1010)

	// Node 1 receives that message (its slow clock says 980), then sends one to node 2
	node1Receive := clockOfNode1.ReceiveMessage(node0Send, 980)
	node1Send := clockOfNode1.SendMessage(990)

	// Node 2 does something on its own, then receives node 1's message
	node2Local := clockOfNode2.LocalEvent(1005)
	node2Receive := clockOfNode2.ReceiveMessage(node1Send, 1065)

	fmt.Println("node 0 local   :", node0Local)
	fmt.Println("node 0 send    :", node0Send)
	fmt.Println("node 1 receive :", node1Receive, " <- its own clock said 980")
	fmt.Println("node 1 send    :", node1Send)
	fmt.Println("node 2 local   :", node2Local)
	fmt.Println("node 2 receive :", node2Receive)

	fmt.Println()
	fmt.Println("node 0 send is before node 1 receive?", node0Send.IsBefore(node1Receive))
	fmt.Println("node 1 send is before node 2 receive?", node1Send.IsBefore(node2Receive))
	fmt.Println("node 2 local is before node 0 send?  ", node2Local.IsBefore(node0Send))
}



// //what HLC fixes and what it doesnt:
// Fixed: causality, like Lamport. The timestamps also look like real times (1010 means roughly 1010 ms), they stay close to the real clock, and they stay two numbers no matter how many nodes there are. That’s the advantage over vector clocks, and it’s why CockroachDB and MongoDB use it.
// Not fixed: concurrency detection. Look at the third comparison. Node 2’s local event is unrelated to node 0’s send, yet HLC ranks it before, just as Lamport did. Only vector clocks can say “concurrent.”



