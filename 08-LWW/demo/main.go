package main

import (
	"fmt"
	"time"

	"github.com/dexisback/SaaT/08-LWW"
)

func main() {
	fmt.Println("=== 1: Wall-Clock LWW Register ===")
	wallReg := lww.NewWallRegister()

	now := time.Now()
	t1 := now
	t2 := now.Add(10 * time.Millisecond)
	tOld := now.Add(-50 * time.Millisecond) // simulates clock skew or out-of-order write

	wallReg.Write("value A (from node 0)", t1)
	fmt.Println("write A at t1:", wallReg.Read().Value)

	wallReg.Write("value B (from node 1)", t2)
	fmt.Println("write B at t2 (newer):", wallReg.Read().Value)

	accepted := wallReg.Write("value C (skewed node)", tOld)
	fmt.Println("write C at tOld accepted?", accepted)
	fmt.Println("register value remains:", wallReg.Read().Value)

	fmt.Println()
	fmt.Println("=== 2: Lamport LWW Register ===")
	lamportReg := lww.NewLamportRegister()

	lamportReg.Write("alice", lww.LamportTimestamp{Counter: 1, NodeID: 0})
	fmt.Println("write alice (counter 1):", lamportReg.Read().Value)

	lamportReg.Write("bob", lww.LamportTimestamp{Counter: 2, NodeID: 0})
	fmt.Println("write bob (counter 2):", lamportReg.Read().Value)

	// Equal counter: node 1 wins over node 0
	lamportReg.Write("charlie", lww.LamportTimestamp{Counter: 2, NodeID: 1})
	fmt.Println("write charlie (counter 2, node 1 - tie break):", lamportReg.Read().Value)
}
