//normal version clock and version vector clock -- both look same visually. as in, [x,y,z].
//the difference is in what they track and what questions do they answer
//vector clock: tracks events on nodes. answers "did event A happen before event B"
//version vector: tracks writes to a single piece of data across replicas. answers "is replica A's copy of this data newer than, older than, or in conflict with replica B's copy "

//classic example: dynamo - style storage . you have one key (user:69) and three replicas. each replica keeps a version vector alongside the value. when a write happens, that replica bumps its own slot
//when replicas sync, they compare vectors to decide : one is newer(take it) , both are the same (fine), or they diverged (conflict, needs resolution)

package main

import "fmt"

type VersionVector []int

func newVersionVector(numberOfReplicas int) VersionVector {
	return make(VersionVector, numberOfReplicas)

}

// make a copyof function:
func copyOf(original VersionVector) VersionVector {
	copied := make(VersionVector, len(original))
	copy(copied, original)
	return copied
}

// relation describes how two version vectors compare
type Relation string

const (
	Equal      Relation = "equal"
	Newer      Relation = "newer"      //first is strictly ahead of second
	Older      Relation = "older"      //when first is strictly behind of second
	Concurrent Relation = "concurrent" //diverged: neither is a descendant of the other
)

// compare asks ki how does the firstVector "relate" to secondVector
func Compare(firstVector VersionVector, secondVector VersionVector) Relation {
	firstKaChotu := false
	firstKaBada := false

	for slot := range firstVector {
		if firstVector[slot] < secondVector[slot] {
			firstKaChotu = true
		}
		if firstVector[slot] > secondVector[slot] {
			firstKaBada = true
		}

	}

	if firstKaChotu && firstKaBada {
		return Concurrent
	}
	if firstKaChotu {
		return Older
	}
	if firstKaBada {
		return Newer
	}

	//else:
	return Equal //very similar to Vector clocks till here
}

// Merge is used when two replicas have diverged(conflicted) and you resolve that conflict.
func Merge(firstVector VersionVector, secondVector VersionVector) VersionVector {
	mergedVector := make(VersionVector, len(firstVector))
	for slot := range firstVector {
		if firstVector[slot] > secondVector[slot] {
			mergedVector[slot] = firstVector[slot]
		} else {
			mergedVector[slot] = secondVector[slot]
		}
	}

	return mergedVector
}

// replica yoyo honey singh holds one replica's copy of a value, with its version vector
type Replica struct {
	replicaID     int
	value         string
	versionVector VersionVector
}

func newReplica(replicaID int, numberofReplicas int) *Replica {
	return &Replica{
		replicaID:     replicaID,
		versionVector: newVersionVector(numberofReplicas),
	}

}

// write stores a new value on thsi replica and bumps its own slot:
func (replica *Replica) Write(newValue string) {
	replica.value = newValue
	replica.versionVector[replica.replicaID]++

}

//sync tries to update this replica from another replica's state. ts returns what happened: took it/ignored it/conflict

func (replica *Replica) Sync(otherValue string, otherVector VersionVector) string {
	relation := Compare(otherVector, replica.versionVector)

	switch relation {
	case Newer:
		//the other replica is strictly ahead: take its value
		replica.value = otherValue
		replica.versionVector = copyOf(otherVector)
		return "took other replica's value (it was newer)"

	case Older, Equal:
		//we're ahead or the same , therefore nothing to do
		return "kept own value (other replica was older or equal)"

	case Concurrent:
		//both sides made independent writes -> therefore , this is a conflict. in real life systems you'd keep both values and let the application code resolve it
		return fmt.Sprintf("CONFLICT ⚠️🤪🥀:  own value=%q vector=%v vs other value=%q vector=%v",
			replica.value, replica.versionVector, otherValue, otherVector)

	}

	return ""

}

func main() {
	fmt.Println("=== normal case: one replica writes, others sync ===")
	fmt.Println()

	replicaA := newReplica(0, 3)
	replicaB := newReplica(1, 3)
	replicaC := newReplica(2, 3)

	replicaA.Write("value-1")
	fmt.Println("replica A writes 'value-1':", replicaA.versionVector)

	result := replicaB.Sync(replicaA.value, replicaA.versionVector)
	fmt.Println("replica B syncs from A:", result)
	fmt.Println("replica B vector now:", replicaB.versionVector)

	result = replicaC.Sync(replicaA.value, replicaA.versionVector)
	fmt.Println("replica C syncs from A:", result)

	fmt.Println()
	fmt.Println("=== conflict case: two replicas write independently ===")
	fmt.Println()

	replicaX := newReplica(0, 2)
	replicaY := newReplica(1, 2)

	// both replicas start in sync
	replicaX.Write("shared-start")
	replicaY.Sync(replicaX.value, replicaX.versionVector)
	fmt.Println("both replicas in sync at:", replicaX.versionVector)

	// now they diverge: network partition, both write independently
	replicaX.Write("x-wrote-this")
	replicaY.Write("y-wrote-this")
	fmt.Println("replica X vector after its write:", replicaX.versionVector)
	fmt.Println("replica Y vector after its write:", replicaY.versionVector)

	// partition heals, they try to sync
	fmt.Println()
	result = replicaX.Sync(replicaY.value, replicaY.versionVector)
	fmt.Println("replica X tries to sync from Y:", result)
	result = replicaY.Sync(replicaX.value, replicaX.versionVector)
	fmt.Println("replica Y tries to sync from X:", result)
}



//so yeh conflict bata paarha
//A vector clock tags an event and asks “what caused this?” A version vector tags a value and asks “which copy is newer?”

//vector clock: "did event X cause event Y"
//version vector : "which replica's copy of this key is newer, or did they conflict"


