//simple. yk it
//jus pick up a winner, whichever has the largest timestamp wins, and all others are silently discarded.
//cassandra uses this. redis uses it for expiring keys. analytical apps use it because they'd rather have a fast answer than a correct one
//two types of LWW clocks possible -- Wall-Clock LWW and Lamport LWW (using those 2 different kinds of clocks(physical clock suffers from skew problem, lamport suffers from the fact that it can solve causality, but cant display "real time"))

package main

import (
	"fmt"
	"sync"
	"time"
)

// 1: wall clock LWW
type WallEntry struct {
	Value     string
	Timestamp time.Time
}

type WallRegister struct {
	lock    sync.Mutex
	current WallEntry
}

func newWallRegister() *WallRegister {
	return &WallRegister{}
}

// write : stores the value only if its timestamp is newer than what we have rn
func (register *WallRegister) Write(value string, timestamp time.Time) bool {
	register.lock.Lock()
	defer register.lock.Unlock()

	if timestamp.After(register.current.Timestamp) {
		register.current = WallEntry{Value: value, Timestamp: timestamp}
		return true // write accepted
	}
	return false //write rejected, older timestamp lost

}

func (register *WallRegister) Read() WallEntry {
	register.lock.Lock()
	defer register.lock.Unlock()
	return register.current
}

//Lamport LWW: so a counter + nodeID for tie-breaking

type LamportTimestamp struct {
	Counter int
	NodeID  int
}


// LamportEntry is one value in a Lamport LWW register.
type LamportEntry struct {
	Value     string
	Timestamp LamportTimestamp
}

// LamportRegister is a single value protected by LWW using Lamport timestamps.
type LamportRegister struct {
	lock    sync.Mutex
	current LamportEntry
}


func newLamportRegister() *LamportRegister {
	return &LamportRegister{}
}

//helpers:
func (first LamportTimestamp) isNewerThan(second LamportTimestamp) bool {
	if first.Counter != second.Counter {
		return first.Counter > second.Counter
	}
	return first.NodeID > second.NodeID
}

func (timestamp LamportTimestamp) String() string {
	return fmt.Sprintf("(counter=%d, node=%d)", timestamp.Counter, timestamp.NodeID)
}



func (register *LamportRegister) Write(value string, timestamp LamportTimestamp) bool {
	register.lock.Lock()
	defer register.lock.Unlock()

	if timestamp.isNewerThan(register.current.Timestamp) {  
		register.current = LamportEntry{Value: value, Timestamp: timestamp}
		return true  //opdate
	}
	return false
}

//read obviously returns the current value only
func (register *LamportRegister) Read() LamportEntry {
	register.lock.Lock()
	defer register.lock.Unlock()
	return register.current
}

func main() {
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 wall-clock LWW: normal case👇🏼👇🏼👇🏼👇🏼 ")
	fmt.Println()

	wallRegister := newWallRegister()

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	time1 := baseTime.Add(100 * time.Millisecond)
	time2 := baseTime.Add(200 * time.Millisecond)
	time3 := baseTime.Add(150 * time.Millisecond) // arrives late, older timestamp

	accepted := wallRegister.Write("write-A", time1)
	fmt.Printf("write-A at %dms: accepted=%v\n", time1.Sub(baseTime).Milliseconds(), accepted)

	accepted = wallRegister.Write("write-B", time2)
	fmt.Printf("write-B at %dms: accepted=%v\n", time2.Sub(baseTime).Milliseconds(), accepted)

	accepted = wallRegister.Write("write-C", time3)
	fmt.Printf("write-C at %dms (arrived late): accepted=%v\n", time3.Sub(baseTime).Milliseconds(), accepted)

	result := wallRegister.Read()
	fmt.Println("final value:", result.Value)
	fmt.Println("write-C was silently dropped even though it arrived after write-B")

	fmt.Println()
	fmt.Println("👇🏼👇🏼👇🏼👇🏼 wall-clock LWW: the skew problem 👇🏼👇🏼👇🏼👇🏼")
	fmt.Println()

	skewedRegister := newWallRegister()



	//simulating: 

	// node 0's clock is accurate
	node0WriteTime := baseTime.Add(200 * time.Millisecond)
	// node 1's clock is 500ms ahead (skewed forward)
	node1WriteTime := baseTime.Add(700 * time.Millisecond)

	// node 0 writes first in real time
	skewedRegister.Write("node-0-value", node0WriteTime)
	fmt.Println("node 0 writes 'node-0-value' at real time 200ms")

	// node 1 writes second in real time, but its skewed clock gives it a bigger timestamp
	skewedRegister.Write("node-1-value", node1WriteTime)
	fmt.Println("node 1 writes 'node-1-value' at real time 200ms too, but clock says 700ms")

	fmt.Println("winner:", skewedRegister.Read().Value)
	fmt.Println("node 1 wins purely because its clock runs fast, not because it wrote later")

	fmt.Println()



	fmt.Println("👇🏼👇🏼👇🏼👇🏼Lamport LWW: skew doesn't matter👇🏼👇🏼👇🏼👇🏼")
	fmt.Println()

	lamportRegister := newLamportRegister()

	// same scenario, but with Lamport timestamps instead of wall time
	// node 0 writes at counter 5, node 1 at counter 3
	// (node 1 happens to have a lower counter even though its wall clock was ahead)
	node0Timestamp := LamportTimestamp{Counter: 5, NodeID: 0}
	node1Timestamp := LamportTimestamp{Counter: 3, NodeID: 1}

	lamportRegister.Write("node-0-value", node0Timestamp)
	fmt.Println("node 0 writes with", node0Timestamp)

	lamportRegister.Write("node-1-value", node1Timestamp)
	fmt.Println("node 1 writes with", node1Timestamp)

	fmt.Println("winner:", lamportRegister.Read().Value)
	fmt.Println("node 0 wins because its Lamport counter is higher, clock skew is irrelevant")

	fmt.Println()
	fmt.Println("=== the fundamental tradeoff ===")
	fmt.Println()
	fmt.Println("LWW always picks a winner, so you never get a conflict to resolve.")
	fmt.Println("The cost: one write is always silently dropped.")
	fmt.Println("Version vectors tell you about the conflict instead of hiding it.")
	fmt.Println("LWW is right for: caches, metrics, analytics, sessions.")
	fmt.Println("Version vectors are right for: user data, financial records, anything that matters.")
}
