package versionvector

import (
	"strings"
	"testing"
)

func TestVersionVectorCompare(t *testing.T) {
	v1 := VersionVector{2, 1, 0}
	v2 := VersionVector{1, 1, 0}
	v3 := VersionVector{1, 2, 0}

	if rel := Compare(v1, v2); rel != Newer {
		t.Fatalf("expected v1 to be Newer than v2, got %s", rel)
	}
	if rel := Compare(v2, v1); rel != Older {
		t.Fatalf("expected v2 to be Older than v1, got %s", rel)
	}
	if rel := Compare(v1, v3); rel != Concurrent {
		t.Fatalf("expected v1 and v3 to be Concurrent (diverged), got %s", rel)
	}
	if rel := Compare(v1, v1); rel != Equal {
		t.Fatalf("expected self-comparison to be Equal, got %s", rel)
	}
}

func TestVersionVectorSyncAndConflict(t *testing.T) {
	rep0 := NewReplica(0, 2)
	rep1 := NewReplica(1, 2)

	// Step 1: rep0 writes
	rep0.Write("first_version") // [1, 0]

	// rep1 syncs from rep0 -> should accept
	res := rep1.Sync(rep0.Value, rep0.VersionVector)
	if !strings.Contains(res, "took other replica's value") {
		t.Fatalf("expected rep1 to accept newer value, got: %s", res)
	}
	if rep1.Value != "first_version" {
		t.Fatalf("expected value 'first_version', got %s", rep1.Value)
	}

	// Step 2: Concurrent writes on rep0 and rep1
	rep0.Write("branch_A") // [2, 0]
	rep1.Write("branch_B") // [1, 1]

	// rep1 attempts to sync branch_A from rep0 -> must detect conflict!
	resConflict := rep1.Sync(rep0.Value, rep0.VersionVector)
	if !strings.Contains(resConflict, "CONFLICT") {
		t.Fatalf("expected conflict detection on divergent writes, got: %s", resConflict)
	}

	// Step 3: Conflict resolution with Merge
	mergedVector := Merge(rep0.VersionVector, rep1.VersionVector)
	if mergedVector[0] != 2 || mergedVector[1] != 1 {
		t.Fatalf("expected merged vector [2, 1], got %v", mergedVector)
	}
}
