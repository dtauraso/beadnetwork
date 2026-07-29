// distance_groups_test.go — the "distance home button" controller's synchronous bounds
// check: ApplyDistanceGroupTarget with an out-of-range group index must return false and
// touch nothing, without needing the mover network running.
package Wiring

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

func repoRootForDistanceGroupsTest(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	// nodes/Wiring/distance_groups_test.go -> repo root is two levels up.
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// TestDistanceGroupOutOfRangeIndexIsNoOp guards ApplyDistanceGroupTarget's bounds
// check: an out-of-range group index must return false and move nothing. This is a pure
// bounds check (out-of-range return happens before any mover is touched), so the mover
// network never needs to be started.
func TestDistanceGroupOutOfRangeIndexIsNoOp(t *testing.T) {
	root := filepath.Join(repoRootForDistanceGroupsTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)
	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(production topology): %v", err)
	}
	if ok := md.ApplyDistanceGroupTarget(99, 1); ok {
		t.Fatal("ApplyDistanceGroupTarget(99, up) = true, want false (out of range)")
	}
}
