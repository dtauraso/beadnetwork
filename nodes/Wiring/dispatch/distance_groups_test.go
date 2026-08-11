// distance_groups_test.go — the "distance home button" controller's synchronous bounds
// check: ApplyDistanceGroupTarget with an out-of-range group index must return false and
// touch nothing, without needing the mover network running.
package dispatch_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/build"
	"github.com/dtauraso/wirefold/nodes/Wiring/distancegroups"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

func repoRootForDistanceGroupsTest(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	// nodes/Wiring/dispatch/distance_groups_test.go -> repo root is three levels up.
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
}

// TestDistanceGroupOutOfRangeIndexIsNoOp guards ApplyDistanceGroupTarget's bounds
// check: an out-of-range group index must return false and move nothing. This is a pure
// bounds check (out-of-range return happens before any mover is touched), so the mover
// network never needs to be started.
func TestDistanceGroupOutOfRangeIndexIsNoOp(t *testing.T) {
	root := filepath.Join(repoRootForDistanceGroupsTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)
	_, _, md, _, err := build.LoadTopology(context.Background(), root, tr, clock.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(production topology): %v", err)
	}
	// ApplyDistanceGroupTarget takes ctx as an explicit parameter (no MoveDispatch field
	// backs it — §35, docs/planning/movedispatch-decomposition.md); context.Background()
	// is the same "no cancellation" value this test needs.
	if ok := distancegroups.ApplyDistanceGroupTarget(context.Background(), &md.UI, &md.MR, &md.LQ, 99, 1); ok {
		t.Fatal("ApplyDistanceGroupTarget(99, up) = true, want false (out of range)")
	}
}
