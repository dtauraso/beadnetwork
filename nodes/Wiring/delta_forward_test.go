package Wiring

// delta_forward_test.go — proves the CASCADE-LINK delta-forward model
// (cascade_links.go's computeCascadeLinks, nodeMover.forwardDelta,
// moveMsgKindDeltaForward): the forwarding graph is every edge MINUS a few
// cycle-closing links (the cascade-link set), so "forward to my cascade-link neighbors,
// excluding the sender, concurrently" is loop-free BY CONSTRUCTION — there is no runtime
// visit-tracking or once-per-drag guard — every node relays on EVERY move it receives,
// never crossing a non-cascade link, and the forwarded log stays in sync with the drag
// as it continues to move (instead of freezing at the first delta, as the old
// forwardedThisDrag guard did).
//
// Real repo topology (topology/) adjacency (edges/*.json):
//
//	1: 2,3   2: 1,5,6   3: 1,9   5: 2,7,8   6: 2,9,10   7: 5   8: 5,10   9: 3,6   10: 6,8
//
// The deterministic spanning-tree walk (BFS from the lexicographically-smallest node id
// "1", visiting each node's neighbors in sorted id order) yields the cascade-link set
// 1-2, 1-3, 2-5, 2-6, 3-9, 5-7, 5-8, 6-10 — leaving exactly two non-cascade (cycle-closing)
// links: 6-9 and 8-10. TestCascadeLinkSetIsDeterministic pins this explicitly.
//
// Dragging leaf node 7 makes 5 the sole direct recipient (gotDragMsg); every other node
// (1,2,3,6,8,9,10) must end up with gotForwardMsg==1 carrying the SAME delta triple 5
// received, via the cascade-link set (never crossing 6-9 or 8-10).
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

// TestCascadeLinkSetIsDeterministic pins the exact cascade-link set the deterministic
// BFS walk yields for the real repo topology: every edge except {6-9, 8-10}. If the
// topology or the walk's tie-break rule ever changes, this test documents and enforces
// the expected result rather than letting it silently drift.
func TestCascadeLinkSetIsDeterministic(t *testing.T) {
	root := filepath.Join(repoRootForDeltaForwardTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)
	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(production topology): %v", err)
	}

	// Every real edge except 6-9 and 8-10 must be a cascade link.
	cascadeEdges := []struct{ a, b string }{
		{"1", "2"}, {"1", "3"}, {"2", "5"}, {"2", "6"}, {"3", "9"}, {"5", "7"}, {"5", "8"}, {"6", "10"},
	}
	for _, e := range cascadeEdges {
		if !md.isCascadeLink(e.a, e.b) {
			t.Errorf("expected %s-%s to be a cascade link, was not", e.a, e.b)
		}
	}
	if got := len(md.cascadeLinks); got != len(cascadeEdges) {
		t.Errorf("cascade-link set size = %d, want %d (set = %v)", got, len(cascadeEdges), md.cascadeLinks)
	}
	// 6-9 and 8-10 must NOT be cascade links.
	wantNonCascade := []struct{ a, b string }{
		{"6", "9"},
		{"8", "10"},
	}
	for _, e := range wantNonCascade {
		if md.isCascadeLink(e.a, e.b) {
			t.Errorf("expected %s-%s to NOT be a cascade link, was marked cascade", e.a, e.b)
		}
	}
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

func TestDeltaForwardPropagatesAcrossWholeGraphAndStaysInSync(t *testing.T) {
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
				ownMover.forwardDelta(md, exceptID, dA, dB, dC)
			}
		}
	}

	// Tap every mover's own outbound sends: record every moveMsgKindDeltaForward
	// (sender -> dest) pair, across the whole test, so we can assert NONE of them cross
	// a dead-end edge (6-9 or 8-10).
	var mu sync.Mutex
	var forwardSends []struct{ from, to string }
	md.SetMsgTap(func(destID string, msg moveMsg) {
		if msg.Kind != moveMsgKindDeltaForward {
			return
		}
		mu.Lock()
		forwardSends = append(forwardSends, struct{ from, to string }{msg.SenderID, destID})
		mu.Unlock()
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	md.Start(ctx)

	sevenBefore, ok := md.centerOfNode("7")
	if !ok {
		t.Fatal("no center for 7")
	}
	firstTarget := sevenBefore.Add(vec3{X: 45, Y: -30, Z: 20})

	md.resetAbcDrag()
	if !md.RootMove("7", firstTarget) {
		t.Fatal("RootMove(7) returned false")
	}
	pollDragConverged(t, md, "7", firstTarget)

	// 5 (7's only neighbor) is the sole direct drag-recipient: gotDragMsg==1, with SOME
	// delta triple — recorded so every forward-recipient below can be checked against
	// the SAME triple (forwardDelta relays it unmodified).
	var firstDA, firstDB, firstDC int32
	waitForNodeDragMsg(t, bufs["5"], func(got uint8, dA, dB, dC, _ int32) bool {
		if got != 1 {
			return false
		}
		firstDA, firstDB, firstDC = dA, dB, dC
		return true
	})

	// Every OTHER node in the connected graph (1,2,3,6,8,9,10 — everyone but the dragged
	// leaf 7) must end up with gotForwardMsg==1 carrying the SAME delta triple, proving
	// full-graph propagation via the tree (reachability survives cutting the two
	// dead-end edges).
	reached := []string{"1", "2", "3", "6", "8", "9", "10"}
	for _, id := range reached {
		waitForNodeForwardMsg(t, bufs[id], func(got uint8, dA, dB, dC, _ int32) bool {
			return got == 1 && dA == firstDA && dB == firstDB && dC == firstDC
		})
	}

	// IN-SYNC REQUIREMENT: move the SAME drag further so 7's delta triple to 5 changes,
	// and assert the forwarded triples across the WHOLE graph UPDATE to the new value —
	// this is exactly what the old forwardedThisDrag guard broke (it froze the forwarded
	// triples at the first pointer-move). No resetAbcDrag/dragStart here: this is a
	// continuation of the SAME drag, mirroring a real further pointer-move mid-drag.
	secondTarget := firstTarget.Add(vec3{X: 60, Y: 25, Z: -35})
	if !md.RootMove("7", secondTarget) {
		t.Fatal("RootMove(7) (second move) returned false")
	}
	pollDragConverged(t, md, "7", secondTarget)

	// Wait for 5's OWN drag-received triple (gotDragMsg side, authoritative for what
	// changed) to reflect the SECOND move — i.e. differ from the first move's triple.
	var secondDA, secondDB, secondDC int32
	deadline := time.Now().Add(2 * time.Second)
	for {
		raw := bufs["5"].Bytes()
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
		if last != nil && len(last) >= nodeOff+B.BufNodeStride {
			dA := int32(binary.LittleEndian.Uint32(last[nodeOff+B.BufNodeColDragDeltaA:]))
			dB := int32(binary.LittleEndian.Uint32(last[nodeOff+B.BufNodeColDragDeltaB:]))
			dC := int32(binary.LittleEndian.Uint32(last[nodeOff+B.BufNodeColDragDeltaC:]))
			if dA != firstDA || dB != firstDB || dC != firstDC {
				secondDA, secondDB, secondDC = dA, dB, dC
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for second-move drag-received triple on 5 to differ from the first")
		}
		time.Sleep(2 * time.Millisecond)
	}

	for _, id := range reached {
		waitForNodeForwardMsg(t, bufs[id], func(got uint8, dA, dB, dC, _ int32) bool {
			return got == 1 && dA == secondDA && dB == secondDB && dC == secondDC
		})
	}

	// Assert the forward wave settles (no more sends drift the state) and that NONE of
	// the recorded forward sends ever crossed a non-cascade link (6-9 or 8-10, in either
	// direction).
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	sendsSnapshot := append([]struct{ from, to string }(nil), forwardSends...)
	mu.Unlock()

	if len(sendsSnapshot) == 0 {
		t.Fatal("no moveMsgKindDeltaForward sends observed at all")
	}
	for _, s := range sendsSnapshot {
		if !md.isCascadeLink(s.from, s.to) {
			t.Errorf("delta-forward crossed a non-cascade link %s -> %s, should never happen", s.from, s.to)
		}
	}

	// dragged node 7 itself never sends a delta-forward (it's the drag origin, driven
	// via moveMsgKindDrag/neighborSetC, not moveMsgKindDeltaForward).
	for _, s := range sendsSnapshot {
		if s.from == "7" {
			t.Fatalf("dragged node 7 sent a delta-forward message to %s, want none", s.to)
		}
	}
}
