package Wiring

// delta_forward_test.go — proves the STORED cascade-edges delta-forward model
// (nodes/<id>/cascade-edges.json, specNode.CascadeEdges, nodeMover.cascadeEdges,
// nodeMover.forwardDelta, moveMsgKindDeltaForward): each node's cascade-neighbor list is
// hand-authored/persisted FILE DATA, not derived from the domain-edge/local-polar
// adjacency at load. "Forward to my stored cascade-edge neighbors, excluding the sender,
// concurrently" is loop-free BY CONSTRUCTION because the seeded set already omits the
// two cycle-closing links (5-8, 7-9) — there is no runtime visit-tracking or
// once-per-drag guard — every node relays on EVERY move it receives, never crossing a
// non-cascade link, and the forwarded log stays in sync with the drag as it continues to
// move (instead of freezing at the first delta, as the old forwardedThisDrag guard did).
//
// Real repo topology (topology/) adjacency (edges/*.json):
//
//	1: 2,3   2: 1,4,5   3: 1,8   4: 2,6,7   5: 2,8,9   6: 4   7: 4,9   8: 3,5   9: 5,7
//
// The seeded nodes/<id>/cascade-edges.json files carry every edge except 5-8 and 7-9:
// 1-2, 1-3, 2-4, 2-5, 3-8, 4-6, 4-7, 5-9. TestCascadeEdgesLoadedFromStoredFiles pins this
// explicitly.
//
// Dragging leaf node 6 makes 4 the sole direct recipient (gotDragMsg); every other node
// (1,2,3,5,7,8,9) must end up with gotForwardMsg==1 carrying the SAME delta triple 4
// received, via the stored cascade-edges graph (never crossing 5-8 or 7-9).
import (
	"context"
	"encoding/binary"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
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

// isCascadeLinkForTest reports whether b is in a's STORED cascade-edges list (or vice
// versa) — the test-side replacement for the removed computed isCascadeLink, reading
// directly off the loaded nodeMover field since this file lives in the same package.
func isCascadeLinkForTest(md *MoveDispatch, a, b string) bool {
	if nm, ok := md.mr.nodeMovers[a]; ok {
		for _, to := range nm.cascadeEdges {
			if to == b {
				return true
			}
		}
	}
	return false
}

// TestCascadeEdgesLoadedFromStoredFiles pins the exact per-node cascade-edges the real
// topology's nodes/<id>/cascade-edges.json seed files carry: every edge except {5-8,
// 7-9}. If the topology or the seeded files ever change, this test documents and
// enforces the expected result rather than letting it silently drift.
func TestCascadeEdgesLoadedFromStoredFiles(t *testing.T) {
	root := filepath.Join(repoRootForDeltaForwardTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)
	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(production topology): %v", err)
	}

	want := map[string][]string{
		"1": {"2", "3"},
		"2": {"1", "4", "5"},
		"3": {"1", "8"},
		"4": {"2", "6", "7"},
		"5": {"2", "9"},
		"6": {"4"},
		"7": {"4"},
		"8": {"3"},
		"9": {"5"},
	}
	for id, wantEdges := range want {
		nm, ok := md.mr.nodeMovers[id]
		if !ok {
			t.Fatalf("no nodeMover for %q", id)
		}
		got := append([]string(nil), nm.cascadeEdges...)
		sort.Strings(got)
		wantSorted := append([]string(nil), wantEdges...)
		sort.Strings(wantSorted)
		if !reflect.DeepEqual(got, wantSorted) {
			t.Errorf("node %q cascadeEdges = %v, want %v", id, got, wantSorted)
		}
	}
	// 5-8 and 7-9 must NOT be cascade links in either direction.
	wantNonCascade := []struct{ a, b string }{
		{"5", "8"}, {"8", "5"},
		{"7", "9"}, {"9", "7"},
	}
	for _, e := range wantNonCascade {
		if isCascadeLinkForTest(md, e.a, e.b) {
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

	allButDragged := []string{"1", "2", "3", "4", "5", "7", "8", "9"}
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
	// a non-cascade link (5-8 or 7-9).
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

	sevenBefore, ok := md.centerOfNode("6")
	if !ok {
		t.Fatal("no center for 6")
	}
	firstTarget := sevenBefore.Add(vec3{X: 45, Y: -30, Z: 20})

	md.resetAbcDrag()
	if !md.RootMove("6", firstTarget) {
		t.Fatal("RootMove(6) returned false")
	}
	pollDragConverged(t, md, "6", firstTarget)

	// 4 (6's only neighbor) is the sole direct drag-recipient: gotDragMsg==1, with SOME
	// delta triple — recorded so every forward-recipient below can be checked against
	// the SAME triple (forwardDelta relays it unmodified).
	var firstDA, firstDB, firstDC int32
	waitForNodeDragMsg(t, bufs["4"], func(got uint8, dA, dB, dC, _ int32) bool {
		if got != 1 {
			return false
		}
		firstDA, firstDB, firstDC = dA, dB, dC
		return true
	})

	// Every OTHER node in the connected graph (1,2,3,5,7,8,9 — everyone but the dragged
	// leaf 6) must end up with gotForwardMsg==1 carrying the SAME delta triple, proving
	// full-graph propagation via the stored cascade-edges graph (reachability survives the two
	// omitted non-cascade links).
	reached := []string{"1", "2", "3", "5", "7", "8", "9"}
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
	if !md.RootMove("6", secondTarget) {
		t.Fatal("RootMove(6) (second move) returned false")
	}
	pollDragConverged(t, md, "6", secondTarget)

	// Wait for 4's OWN drag-received triple (gotDragMsg side, authoritative for what
	// changed) to reflect the SECOND move — i.e. differ from the first move's triple.
	var secondDA, secondDB, secondDC int32
	deadline := time.Now().Add(2 * time.Second)
	for {
		raw := bufs["4"].Bytes()
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
			t.Fatal("timed out waiting for second-move drag-received triple on 4 to differ from the first")
		}
		time.Sleep(2 * time.Millisecond)
	}

	for _, id := range reached {
		waitForNodeForwardMsg(t, bufs[id], func(got uint8, dA, dB, dC, _ int32) bool {
			return got == 1 && dA == secondDA && dB == secondDB && dC == secondDC
		})
	}

	// Assert the forward wave settles (no more sends drift the state) and that NONE of
	// the recorded forward sends ever crossed a non-cascade link (5-8 or 7-9, in either
	// direction).
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	sendsSnapshot := append([]struct{ from, to string }(nil), forwardSends...)
	mu.Unlock()

	if len(sendsSnapshot) == 0 {
		t.Fatal("no moveMsgKindDeltaForward sends observed at all")
	}
	for _, s := range sendsSnapshot {
		if !isCascadeLinkForTest(md, s.from, s.to) {
			t.Errorf("delta-forward crossed a non-cascade link %s -> %s, should never happen", s.from, s.to)
		}
	}

	// dragged node 6 itself never sends a delta-forward (it's the drag origin, driven
	// via moveMsgKindDrag/neighborSetC, not moveMsgKindDeltaForward).
	for _, s := range sendsSnapshot {
		if s.from == "6" {
			t.Fatalf("dragged node 6 sent a delta-forward message to %s, want none", s.to)
		}
	}
}

// TestTimeEndIgnoresDeltaForwardCascade proves TimeEnd nodes (the terminal node of a time
// chain) IGNORE a cascaded delta-forward triple: no gotForwardMsg record and no relay
// onward. Node 6 is TimeEnd (see nodes/TimeEnd/node.go, wire.Register("TimeEnd")) and is a
// degree-1 leaf off node 4 (4: 2,6,7 per the adjacency table above), so ignoring changes
// nothing about reachability for anyone else -- 6 never relayed to a third node anyway.
//
// Dragging node "1" floods the cascade graph (1-2, 2-4, 4-6, 4-7, ...) so node 4 receives
// a delta-forward and forwards it to BOTH 6 and 7. Node 7 (a plain leaf, NOT TimeEnd) must
// end up with gotForwardMsg==1; node 6 (TimeEnd) must stay at gotForwardMsg==0 forever --
// proving the gate in node_mover.go's moveMsgKindDeltaForward handler (checked via
// m.geom.Kind == "TimeEnd", not the numeric kindID) actually suppresses both the record
// and the relay.
func TestTimeEndIgnoresDeltaForwardCascade(t *testing.T) {
	root := filepath.Join(repoRootForDeltaForwardTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)

	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(production topology): %v", err)
	}

	if got := md.NodeKind("6"); got != "TimeEnd" {
		t.Fatalf("node 6 kind = %q, want TimeEnd (test assumes 6 is the TimeEnd leaf off 4)", got)
	}

	ids := []string{"1", "2", "4", "6", "7"}
	bufs := map[string]*uiPubLockedBuf{}
	for _, id := range ids {
		bufs[id] = wireNodeStream(t, md, id)
		if nm, ok := md.mr.nodeMovers[id]; ok {
			nm.nodeRowFor = md.NodeRowFor
			ownMover := nm
			nm.forwardOnce = func(exceptID string, dA, dB, dC int32) {
				ownMover.forwardDelta(md, exceptID, dA, dB, dC)
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	md.Start(ctx)

	before, ok := md.centerOfNode("1")
	if !ok {
		t.Fatal("no center for 1")
	}
	target := before.Add(vec3{X: 45, Y: -30, Z: 20})

	md.resetAbcDrag()
	if !md.RootMove("1", target) {
		t.Fatal("RootMove(1) returned false")
	}
	pollDragConverged(t, md, "1", target)

	// Node 7 (plain leaf off 4, NOT TimeEnd) must receive and record the forward.
	waitForNodeForwardMsg(t, bufs["7"], func(got uint8, _, _, _, _ int32) bool {
		return got == 1
	})

	// Node 6 (TimeEnd) must NEVER record the forward, even after giving the cascade
	// plenty of time to reach and be ignored by it.
	time.Sleep(100 * time.Millisecond)
	if got, _, _, _, _, ok := lastNodeStreamForwardMsg(bufs["6"].Bytes()); ok && got != 0 {
		t.Fatalf("TimeEnd node 6 recorded gotForwardMsg=%d, want 0 (cascade must be ignored)", got)
	}
}
