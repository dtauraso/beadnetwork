package Wiring

// chain_end_probe_verify_test.go — TEMPORARY (task/log-the-chain-distances): proves the
// three new debug breadcrumbs (chain-drag-stride / chain-end-dist / chain-end-inputs)
// actually fire on a real drag, through the real loader + mover goroutines, following
// the same syncBuffer pattern time_node_abc_drag_breadcrumb_test.go uses for "abc-drag".
// This is instrumentation-verification only, not a behavior assertion; remove alongside
// the breadcrumbs once the gap/overlap symptom is diagnosed.

import (
	"context"
	"testing"
	"time"
)

// waitForBreadcrumb blocks until at least one breadcrumb with the given label naming
// node has appeared in dbg, or fails the test on timeout — same shape as waitForAbcDrag.
func waitForBreadcrumb(t *testing.T, dbg *syncBuffer, label, node string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		for _, b := range parseBreadcrumbLines(t, dbg.String()) {
			if b.Label == label && b.Node == node {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for a %q breadcrumb for node %q; got: %s", label, node, dbg.String())
		}
		time.Sleep(time.Millisecond)
	}
}

// TestChainEndProbeBreadcrumbsFire drags "dst" on the src->dst fixture and asserts each
// of the three new breadcrumbs actually reaches the buffer-decoded stream:
//   - chain-drag-stride, on "dst" itself (the dragged node, item A).
//   - chain-end-dist / chain-end-inputs, on "src" (item B/C) — src OWNS the outgoing
//     edge e0 aimed AT dst, so this is exactly the "incoming chain owned by a neighbour,
//     aimed at the dragged node" case the symptom names: src's own moveMsgKindNeighborSetC
//     handler (neighborSetCRequantize) arms src's chainProbeDirty, and src's own next
//     chainBeads() call (its regular per-cycle emit) logs B/C for edge e0.
func TestChainEndProbeBreadcrumbsFire(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	md.EnableEditPersist(root)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := md.Start(ctx)
	t.Cleanup(func() { cancel(); wg.Wait() })

	var dbg syncBuffer
	md.tr.SetSink(&dbg)

	center, ok := md.centerOfNode("dst")
	if !ok {
		t.Fatal("no center for dst")
	}
	target := center.Add(vec3{X: 60, Y: -40, Z: 30})
	wantDst := quantizedDragTarget(md, "dst", target)
	if !md.RootMove("dst", target) {
		t.Fatal("RootMove(dst) returned false")
	}
	pollDragConverged(t, md, "dst", wantDst)

	// Item A: the dragged node's own drag-stride breadcrumb.
	waitForBreadcrumb(t, &dbg, "chain-drag-stride", "dst")

	// Item B/C setup: wait for src's OWN "abc-drag" breadcrumb — neighborSetCRequantize
	// writes src's requantized LocalPolar AND arms src.chainProbeDirty strictly BEFORE
	// logging that breadcrumb, in the same call on src's own goroutine (mirrors
	// waitForAbcDrag's documented happens-before argument in
	// time_node_abc_drag_breadcrumb_test.go) — so once it is observed here, reading
	// src's state from this (test) goroutine is race-free.
	waitForAbcDrag(t, &dbg, "src")

	// This headless test wires no per-node stream (no WIREFOLD_STREAM_FDS), so
	// writeStreamFrame's streamOut==nil guard means src's own run-loop never actually
	// calls chainBeads() — that only happens in the real editor, which DOES wire
	// SetNodeStreams. Call it directly here (same as production's next writeStreamFrame
	// would) to prove the B/C breadcrumbs themselves fire correctly once chainBeads
	// runs, matching the sync argument above.
	nm, ok := md.mr.nodeMovers["src"]
	if !ok {
		t.Fatal("no nodeMover for src")
	}
	if !nm.chainProbeDirty {
		t.Fatal("expected src.chainProbeDirty to be armed by neighborSetCRequantize after the abc-drag breadcrumb")
	}
	nm.chainBeads()

	waitForBreadcrumb(t, &dbg, "chain-end-dist", "src")
	waitForBreadcrumb(t, &dbg, "chain-end-inputs", "src")
}
