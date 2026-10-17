package versionvector

import (
	"fmt"
)

// VersionVector tracks writes to a piece of data across replicas.
type VersionVector []int

func NewVersionVector(numberOfReplicas int) VersionVector {
	return make(VersionVector, numberOfReplicas)
}

func CopyOf(original VersionVector) VersionVector {
	copied := make(VersionVector, len(original))
	copy(copied, original)
	return copied
}

// Relation describes how two version vectors compare.
type Relation string

const (
	Equal      Relation = "equal"
	Newer      Relation = "newer"      // first is strictly ahead of second
	Older      Relation = "older"      // first is strictly behind second
	Concurrent Relation = "concurrent" // diverged: neither is a descendant of the other
)

// Compare returns the relation between firstVector and secondVector.
func Compare(firstVector VersionVector, secondVector VersionVector) Relation {
	firstHasSmaller := false
	firstHasLarger := false

	for slot := range firstVector {
		if firstVector[slot] < secondVector[slot] {
			firstHasSmaller = true
		}
		if firstVector[slot] > secondVector[slot] {
			firstHasLarger = true
		}
	}

	if firstHasSmaller && firstHasLarger {
		return Concurrent
	}
	if firstHasSmaller {
		return Older
	}
	if firstHasLarger {
		return Newer
	}
	return Equal
}

// Merge combines two version vectors by taking the element-wise maximum.
func Merge(firstVector VersionVector, secondVector VersionVector) VersionVector {
	merged := make(VersionVector, len(firstVector))
	for slot := range firstVector {
		if firstVector[slot] > secondVector[slot] {
			merged[slot] = firstVector[slot]
		} else {
			merged[slot] = secondVector[slot]
		}
	}
	return merged
}

// Replica holds one replica's copy of a value with its associated version vector.
type Replica struct {
	ReplicaID     int
	Value         string
	VersionVector VersionVector
}

func NewReplica(replicaID int, numberOfReplicas int) *Replica {
	return &Replica{
		ReplicaID:     replicaID,
		VersionVector: NewVersionVector(numberOfReplicas),
	}
}

// Write stores a new value on this replica and increments its own slot.
func (replica *Replica) Write(newValue string) {
	replica.Value = newValue
	replica.VersionVector[replica.ReplicaID]++
}

// Sync merges state from another replica.
func (replica *Replica) Sync(otherValue string, otherVector VersionVector) string {
	relation := Compare(otherVector, replica.VersionVector)

	switch relation {
	case Newer:
		replica.Value = otherValue
		replica.VersionVector = CopyOf(otherVector)
		return "took other replica's value (it was newer)"

	case Older, Equal:
		return "kept own value (other replica was older or equal)"

	case Concurrent:
		return fmt.Sprintf("CONFLICT: own value=%q vector=%v vs other value=%q vector=%v",
			replica.Value, replica.VersionVector, otherValue, otherVector)
	}

	return ""
}

// ResolveConflict applies a merged value and merges the two version vectors.
func (replica *Replica) ResolveConflict(resolvedValue string, otherVector VersionVector) {
	replica.Value = resolvedValue
	replica.VersionVector = Merge(replica.VersionVector, otherVector)
	replica.VersionVector[replica.ReplicaID]++ // bump our slot for the resolution write
}
