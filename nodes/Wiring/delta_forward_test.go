package Wiring

// delta_forward_test.go — proves the delta-forward FULL-GRAPH propagation model
// (moveMsgKindDeltaForward, nodeMover.forwardDeltaOnce): each node forwards the delta
// triple it first picks up EXACTLY ONCE per drag, to every neighbor except the one it
// came from — the drag-recipient's own hop (quantized_move.go neighborSetCRequantize)
// and every subsequent forward-recipient's own hop (node_mover.go handle's
// moveMsgKindDeltaForward case) share the SAME once-per-drag guard
// (nm.forwardedThisDrag), so the triple spreads across the WHOLE reachable graph via
// independent concurrent single hops and terminates on the graph's cycles instead of
// looping forever.
//
// Real repo topology (topology/) adjacency (edges/*.json):
//
//	1: 2,3   2: 1,5,6   3: 1,9   5: 2,7,8   6: 2,9,10   7: 5   8: 5,10   9: 3,6   10: 6,8
//
// Dragging leaf node 7 makes 5 the sole direct recipient (gotDragMsg); every other node
// (1,2,3,6,8,9,10) must end up with gotForwardMsg==1 carrying the SAME delta triple 5
// received. Because each forwarding node relays to (its own neighbor count - 1) peers
// regardless of WHICH neighbor delivered the delta first (the "except" is always
// exactly one real edge), the total number of moveMsgKindDeltaForward sends, per
// SenderID, is deterministic: degree(node)-1 for every node but the dragged leaf 7
// (which never forwards a delta — RootMove drives it via moveMsgKindDrag, not
// moveMsgKindNeighborSetC/DeltaForward). A tap counts these sends and asserts the exact
// expected total per sender — if the once-per-drag guard broke, a node reachable by
// more than one path (3, 9, 10, all with degree 2 sitting between two forward routes)
// would send extra messages the moment a second delta arrived.
import (
	"context"
	"encoding/binary"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	B "github.com/dtauraso/wirefold/Buffer"
	T "github.com/dtauraso/wirefold/Trace"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

func repoRootForDeltaForwardTest(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	// nodes/Wiring/delta_forward_test.go -> repo root is two levels up.
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// lastNodeStreamForwardMsg decodes the LAST complete node-stream frame's
// GotForwardMsg/ForwardDeltaA/B/C/ForwardFromRow fields, mirroring
// lastNodeStreamDragMsg (ui_publish_propagation_test.go).
func lastNodeStreamForwardMsg(raw []byte) (gotForwardMsg uint8, deltaA, deltaB, deltaC, fromRow int32, ok bool) {
	off := 0
	var last []byte
	for off+4 <= len(raw) {
		n := int(binary.LittleEndian.Uint32(raw[off:]))
		off += 4
		if off+n > len(raw) {
			break
		}
		last = raw[off : off+n]
		off += n
	}
	const nodeOff = 20
	if last == nil || len(last) < nodeOff+B.BufNodeStride {
		return 0, 0, 0, 0, 0, false
	}
	return last[nodeOff+B.BufNodeColGotForwardMsg],
		int32(binary.LittleEndian.Uint32(last[nodeOff+B.BufNodeColForwardDeltaA:])),
		int32(binary.LittleEndian.Uint32(last[nodeOff+B.BufNodeColForwardDeltaB:])),
		int32(binary.LittleEndian.Uint32(last[nodeOff+B.BufNodeColForwardDeltaC:])),
		int32(binary.LittleEndian.Uint32(last[nodeOff+B.BufNodeColForwardFromRow:])),
		true
}

// waitForNodeForwardMsg is waitForNodeDragMsg's delta-forward counterpart.
func waitForNodeForwardMsg(t *testing.T, buf *uiPubLockedBuf, check func(gotForwardMsg uint8, dA, dB, dC, fromRow int32) bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if got, dA, dB, dC, fromRow, ok := lastNodeStreamForwardMsg(buf.Bytes()); ok && check(got, dA, dB, dC, fromRow) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("node's dedicated stream frame never reflected the expected delta-forward state within deadline")
		}
		time.Sleep(2 * time.Millisecond)
	}
}

func TestDeltaForwardPropagatesAcrossWholeGraphOncePerNode(t *testing.T) {
	root := filepath.Join(repoRootForDeltaForwardTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)

	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(production topology): %v", err)
	}
	// No EnableEditPersist: this test must not write to the real on-disk topology.

	allButDragged := []string{"1", "2", "3", "5", "6", "8", "9", "10"}
	bufs := map[string]*uiPubLockedBuf{}
	for _, id := range allButDragged {
		bufs[id] = wireNodeStream(t, md, id)
		// wireNodeStream (abc_drag_scope_test.go) sets streamOut/nodeRow/buildFrame
		// directly but does NOT wire nodeRowFor/forwardOnce (SetNodeStreams' job in
		// production, never called by this bare-LoadTopology test harness) — this
		// feature's forward handler needs both, so wire them here for every node under
		// test. forwardOnce mirrors the exact closure newMoveDispatch installs
		// (node_move.go), just resolved after construction instead of at it.
		if nm, ok := md.mr.nodeMovers[id]; ok {
			nm.nodeRowFor = md.NodeRowFor
			ownMover := nm
			nm.forwardOnce = func(exceptID string, dA, dB, dC int32) {
				ownMover.forwardDeltaOnce(md, exceptID, dA, dB, dC)
			}
		}
	}

	// Tap every mover's own outbound sends: count moveMsgKindDeltaForward messages by
	// SenderID, across the whole test, to prove each forwarding node relays exactly
	// once (degree(node)-1 sends), not repeatedly.
	var mu sync.Mutex
	forwardSendCount := map[string]int{}
	md.SetMsgTap(func(_ string, msg moveMsg) {
		if msg.Kind != moveMsgKindDeltaForward {
			return
		}
		mu.Lock()
		forwardSendCount[msg.SenderID]++
		mu.Unlock()
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	md.Start(ctx)

	sevenBefore, ok := md.centerOfNode("7")
	if !ok {
		t.Fatal("no center for 7")
	}
	sevenTarget := sevenBefore.Add(vec3{X: 45, Y: -30, Z: 20})

	md.resetAbcDrag()
	if !md.RootMove("7", sevenTarget) {
		t.Fatal("RootMove(7) returned false")
	}
	pollDragConverged(t, md, "7", sevenTarget)

	// 5 (7's only neighbor, 5To7) is the sole direct drag-recipient: gotDragMsg==1,
	// with SOME delta triple — recorded so every forward-recipient below can be
	// checked against the SAME triple (forwardDeltaOnce relays it unmodified).
	var wantDA, wantDB, wantDC int32
	waitForNodeDragMsg(t, bufs["5"], func(got uint8, dA, dB, dC, _ int32) bool {
		if got != 1 {
			return false
		}
		wantDA, wantDB, wantDC = dA, dB, dC
		return true
	})

	// Every OTHER node in the connected graph (1,2,3,6,8,9,10 — everyone but the
	// dragged leaf 7) must end up with gotForwardMsg==1 carrying the SAME delta
	// triple, proving full-graph propagation (not just one hop past 5).
	for _, id := range []string{"1", "2", "3", "6", "8", "9", "10"} {
		waitForNodeForwardMsg(t, bufs[id], func(got uint8, dA, dB, dC, _ int32) bool {
			return got == 1 && dA == wantDA && dB == wantDB && dC == wantDC
		})
	}

	// ONCE PER NODE: wait for the forward wave to fully settle (no new sends), then
	// assert the EXACT per-sender forward count — degree(node)-1 for every forwarding
	// node, regardless of which neighbor happened to deliver the delta first (the
	// "except" is always exactly one real edge). If the once-per-drag guard broke, a
	// node with more than one route in (3, 9, 10 — each degree 2, each sitting between
	// two forward paths) would send extra messages the moment its second delta
	// arrived.
	want := map[string]int{
		"1":  1, // neighbors {2,3}, forwards to the one other than its source
		"2":  2, // neighbors {1,5,6}
		"3":  1, // neighbors {1,9}
		"5":  2, // neighbors {2,7,8}
		"6":  2, // neighbors {2,9,10}
		"8":  1, // neighbors {5,10}
		"9":  1, // neighbors {3,6}
		"10": 1, // neighbors {6,8}
	}
	settleDeadline := time.Now().Add(2 * time.Second)
	var last map[string]int
	for {
		mu.Lock()
		snap := make(map[string]int, len(forwardSendCount))
		for k, v := range forwardSendCount {
			snap[k] = v
		}
		mu.Unlock()
		if mapsEqualInt(snap, want) {
			last = snap
			break
		}
		if time.Now().After(settleDeadline) {
			last = snap
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !mapsEqualInt(last, want) {
		t.Fatalf("forward send counts by SenderID = %v, want %v (a mismatch means a node forwarded more/less than once)", last, want)
	}
	// Give any (incorrect) extra re-forward more time to land, then confirm the counts
	// held steady — a broken guard would keep incrementing past the expected total.
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	final := make(map[string]int, len(forwardSendCount))
	for k, v := range forwardSendCount {
		final[k] = v
	}
	mu.Unlock()
	if !mapsEqualInt(final, want) {
		t.Fatalf("forward send counts drifted after settling: %v, want %v (a node re-forwarded)", final, want)
	}

	// dragged node 7 itself never sends a delta-forward (it's the drag origin, driven
	// via moveMsgKindDrag/neighborSetC, not moveMsgKindDeltaForward).
	mu.Lock()
	sevenCount := forwardSendCount["7"]
	mu.Unlock()
	if sevenCount != 0 {
		t.Fatalf("dragged node 7 sent %d delta-forward messages, want 0", sevenCount)
	}
}

func mapsEqualInt(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range b {
		if a[k] != v {
			return false
		}
	}
	return true
}
