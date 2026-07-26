package Wiring

// msg_tap_test.go — proves the per-mover message tap (node_mover.go's nm.tap, installed
// via MoveDispatch.SetMsgTap before Start) still observes real mover-to-mover traffic
// after the refactor from a single shared atomic.Pointer to per-mover plain-field
// ownership (each mover reads only its OWN nm.tap, on its OWN goroutine — see
// enqueueFor in mover_registry.go).

import (
	"context"
	"sync"
	"testing"
)

// TestSetMsgTapObservesNeighborSetC drives the same real drag → neighborSetC
// propagation as neighbor_setc_test.go (writeTree's plain 2-node src/dst graph) but
// installs a recording SetMsgTap BEFORE Start and asserts the tap observed the
// dst->src neighborSetC message enqueueFor routes on dst's own goroutine.
func TestSetMsgTapObservesNeighborSetC(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	md.EnableEditPersist(root)

	var mu sync.Mutex
	var seen []struct {
		destID string
		kind   string
		sender string
	}
	// SetMsgTap must be called before Start (setup-goroutine write to every mover's
	// plain nm.tap field — happens-before every mover goroutine's later read).
	md.SetMsgTap(func(destID string, msg moveMsg) {
		mu.Lock()
		defer mu.Unlock()
		seen = append(seen, struct {
			destID string
			kind   string
			sender string
		}{destID, msg.Kind, msg.SenderID})
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	md.Start(ctx)

	// Sync point: src (the neighborSetC recipient) logs its own "abc-drag" breadcrumb
	// strictly AFTER handling the neighborSetC message that would have fired the tap
	// (see neighbor_setc_test.go's identical sync-point comment) — waiting for it gives
	// a happens-before edge for the read of `seen` below instead of racing src's own
	// goroutine.
	var dbg syncBuffer
	md.tr.SetSink(&dbg)

	dstBefore, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst")
	}
	target := dstBefore.Add(vec3{X: 60, Y: 25, Z: -15})
	if !md.RootMove("dst", target) {
		t.Fatal("RootMove(dst) returned false")
	}
	pollDragConverged(t, md, "dst", target)
	waitForAbcDrag(t, &dbg, "src")

	found := false
	mu.Lock()
	for _, e := range seen {
		if e.destID == "src" && e.kind == moveMsgKindNeighborSetC && e.sender == "dst" {
			found = true
			break
		}
	}
	got := len(seen)
	mu.Unlock()

	if !found {
		t.Fatalf("expected the tap to observe a neighborSetC message from dst to src; observed %d entries", got)
	}
}
