package lww

import (
	"fmt"
	"sync"
	"time"
)

// WallEntry is one value in a wall-clock LWW register.
type WallEntry struct {
	Value     string
	Timestamp time.Time
}

// WallRegister protects a single value using wall-clock LWW.
type WallRegister struct {
	lock    sync.Mutex
	current WallEntry
}

func NewWallRegister() *WallRegister {
	return &WallRegister{}
}

// Write stores the value only if its timestamp is newer than what we currently have.
func (register *WallRegister) Write(value string, timestamp time.Time) bool {
	register.lock.Lock()
	defer register.lock.Unlock()

	if timestamp.After(register.current.Timestamp) {
		register.current = WallEntry{Value: value, Timestamp: timestamp}
		return true
	}
	return false
}

func (register *WallRegister) Read() WallEntry {
	register.lock.Lock()
	defer register.lock.Unlock()
	return register.current
}

// LamportTimestamp tie-breaks equal counters using NodeID.
type LamportTimestamp struct {
	Counter int
	NodeID  int
}

func (first LamportTimestamp) IsNewerThan(second LamportTimestamp) bool {
	if first.Counter != second.Counter {
		return first.Counter > second.Counter
	}
	return first.NodeID > second.NodeID
}

func (timestamp LamportTimestamp) String() string {
	return fmt.Sprintf("(counter=%d, node=%d)", timestamp.Counter, timestamp.NodeID)
}

// LamportEntry is one value in a Lamport LWW register.
type LamportEntry struct {
	Value     string
	Timestamp LamportTimestamp
}

// LamportRegister protects a single value using Lamport timestamps.
type LamportRegister struct {
	lock    sync.Mutex
	current LamportEntry
}

func NewLamportRegister() *LamportRegister {
	return &LamportRegister{}
}

func (register *LamportRegister) Write(value string, timestamp LamportTimestamp) bool {
	register.lock.Lock()
	defer register.lock.Unlock()

	if timestamp.IsNewerThan(register.current.Timestamp) {
		register.current = LamportEntry{Value: value, Timestamp: timestamp}
		return true
	}
	return false
}

func (register *LamportRegister) Read() LamportEntry {
	register.lock.Lock()
	defer register.lock.Unlock()
	return register.current
}