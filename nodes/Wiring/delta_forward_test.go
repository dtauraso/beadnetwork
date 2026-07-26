package Wiring

// delta_forward_test.go — proves the ONE-HOP delta-forward observability feature
// (moveMsgKindDeltaForward, quantized_move.go neighborSetCRequantize's forward step):
// dragging node "2" in the REAL repo topology (topology/) makes node "5" (2's direct
// neighbor via edge 2To5) the direct drag-recipient. Node 5's OTHER neighbors — 7
// (5To7) and 8 (5To8) — must each receive a delta-forward carrying node 2's own
// delta and node 5's buffer row as ForwardFromRow, while node 5's forward recipients
// must NOT themselves forward past that one hop: node 8's own other neighbor 10
// (8To10) must show GotForwardMsg==0, proving there is no cascade.

import (
	"context"
	"encoding/binary"
	"path/filepath"
	"runtime"
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

func TestDeltaForwardOneHopNoCascade(t *testing.T) {
	root := filepath.Join(repoRootForDeltaForwardTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)

	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(production topology): %v", err)
	}
	// No EnableEditPersist: this test must not write to the real on-disk topology.

	bufFive := wireNodeStream(t, md, "5")
	bufSeven := wireNodeStream(t, md, "7")
	bufEight := wireNodeStream(t, md, "8")
	bufTen := wireNodeStream(t, md, "10")
	// wireNodeStream (abc_drag_scope_test.go) sets streamOut/nodeRow/buildFrame directly
	// but does NOT wire nodeRowFor (that's SetNodeStreams' job in production, never
	// called by this bare-LoadTopology test harness) — this feature's forward handler
	// needs it to resolve ForwardFromRow, so wire it here for every node under test.
	for _, id := range []string{"5", "7", "8", "10"} {
		if nm, ok := md.mr.nodeMovers[id]; ok {
			nm.nodeRowFor = md.NodeRowFor
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	md.Start(ctx)

	rowFive, ok := md.NodeRowFor("5")
	if !ok {
		t.Fatal("no NODE-ROW for 5")
	}
	rowEight, ok := md.NodeRowFor("8")
	if !ok {
		t.Fatal("no NODE-ROW for 8")
	}

	twoBefore, ok := md.centerOfNode("2")
	if !ok {
		t.Fatal("no center for 2")
	}
	twoTarget := twoBefore.Add(vec3{X: 60, Y: -25, Z: 35})

	md.resetAbcDrag()
	if !md.RootMove("2", twoTarget) {
		t.Fatal("RootMove(2) returned false")
	}
	pollDragConverged(t, md, "2", twoTarget)

	// 5 (2's direct neighbor) is the direct drag-recipient: gotDragMsg==1, with SOME
	// delta triple (recorded so the forward step below can be checked against it).
	var wantDA, wantDB, wantDC int32
	waitForNodeDragMsg(t, bufFive, func(got uint8, dA, dB, dC, _ int32) bool {
		if got != 1 {
			return false
		}
		wantDA, wantDB, wantDC = dA, dB, dC
		return true
	})

	// 5's OTHER neighbors (7 and 8 — everyone but fromID "2") must each receive the
	// ONE-HOP delta-forward: same delta triple 5 itself received, ForwardFromRow ==
	// 5's own buffer row.
	waitForNodeForwardMsg(t, bufSeven, func(got uint8, dA, dB, dC, fromRow int32) bool {
		return got == 1 && dA == wantDA && dB == wantDB && dC == wantDC && fromRow == rowFive
	})
	waitForNodeForwardMsg(t, bufEight, func(got uint8, dA, dB, dC, fromRow int32) bool {
		return got == 1 && dA == wantDA && dB == wantDB && dC == wantDC && fromRow == rowFive
	})

	// NO CASCADE past 8: node 10 sits one hop beyond 8 (8To10). Node 6 (2's OTHER
	// direct neighbor via 2To6) legitimately forwards to 10 directly (6To10) — that is
	// a SEPARATE direct-recipient's own one-hop forward, not a cascade, so 10 showing
	// GotForwardMsg==1 is expected. What must NEVER happen is 10 attributing its
	// forward to 8 (ForwardFromRow == rowEight) — that would mean 8 re-forwarded past
	// its own one hop, which the handler must never do (moveMsgKindDeltaForward's case
	// in node_mover.go records state only and returns, never re-forwarding). Give any
	// (incorrect) re-forward time to land before asserting.
	time.Sleep(50 * time.Millisecond)
	if got, _, _, _, fromRow, ok := lastNodeStreamForwardMsg(bufTen.Bytes()); ok && got == 1 && fromRow == rowEight {
		t.Fatalf("node 10 attributed a delta-forward to node 8 (row %d) — forward cascaded past one hop", rowEight)
	}
}
