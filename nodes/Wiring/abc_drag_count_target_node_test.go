package Wiring

import (
	"context"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	B "github.com/dtauraso/wirefold/Buffer"
	T "github.com/dtauraso/wirefold/Trace"
)

// abc_drag_count_target_node_test.go — reproduces the live bug report: dragging a node
// whose ONLY edge has that node as the TARGET (e.g. 5->7, drag 7) kept the "drag received
// ×N" VIEW-frame counter at 0, while a node with multiple neighbors (e.g. 6, neighbors
// 2/9/10) worked. This drives the REAL gesture-adjacent path (md.resetAbcDrag then
// md.RootMove, mirroring commitDragStart+the dragging-phase apply) rather than a hand-
// built moveMsgKindNeighborSetC send, and wires the VIEW stream the way
// TestGesturePathPropagatesUIStateToMoverStream does so abcDragCh is actually live.
func writeFT(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mk := func(rel, body string) { writeTreeFile(t, root, rel, body) }
	// f = HoldNewSendOld (has an unwired FromPrevHoldNewSendOldNode input port — an
	// unfed required input is inert, not rejected, per
	// memory/feedback_enforce_required_inputs.md), t = SinkNode, single edge f->t.
	mk("nodes/f/meta.json", `{"id":"f","type":"HoldNewSendOld","data":{"state":{"held":0}},"r":100,"scenePolarR":90,"scenePolarTheta":0.9,"scenePolarPhi":-2.1}`)
	mk("nodes/f/inputs/FromPrevHoldNewSendOldNode.json", `{"name":"FromPrevHoldNewSendOldNode"}`)
	mk("nodes/f/outputs/ToNext.json", `{"name":"ToNext"}`)
	mk("nodes/t/meta.json", `{"id":"t","type":"SinkNode","r":100,"scenePolarR":90,"scenePolarTheta":2.0,"scenePolarPhi":0.4}`)
	mk("nodes/t/inputs/In.json", `{"name":"In"}`)
	mk("edges/eFT.json", `{"label":"eFT","kind":"chain","source":"f","sourceHandle":"ToNext","target":"t","targetHandle":"In"}`)
	return root
}

// TestAbcDragCountAdvancesWhenDraggingTargetOnlyNode drags "t", the TARGET (not source)
// of its only edge f->t, and asserts the VIEW frame's AbcDragCount reaches >= 1 — the
// same assertion shape as TestGesturePathPropagatesUIStateToMoverStream's abc-drag
// section, but through the real drag entry (resetAbcDrag + RootMove) instead of a
// hand-crafted moveMsgKindNeighborSetC send.
func TestAbcDragCountAdvancesWhenDraggingTargetOnlyNode(t *testing.T) {
	root := writeFT(t)
	tr := T.NewWithSinkHook(nil, nil)

	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}

	viewBuf := &uiPubLockedBuf{}
	md.SetViewStream(viewBuf, func(tick uint32,
		camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
		sceneTori, scenePoles, nodePoles, selSpherePoles, handholds, labelsGlobal, overlaysVis, doubleLinks uint8,
		abcDragCount uint32,
		dragNodeRow int32,
		sceneCX, sceneCY, sceneCZ, sceneRadius float32,
		events []wire.RowEvent,
	) []byte {
		return B.BuildViewStreamFrame(tick, camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi,
			B.OverlayRow{
				SceneTori: sceneTori, ScenePoles: scenePoles, NodePoles: nodePoles,
				SelSpherePoles: selSpherePoles, Handholds: handholds, LabelsGlobal: labelsGlobal,
				OverlaysVis: overlaysVis, DoubleLinks: doubleLinks, AbcDragCount: abcDragCount,
				DragNodeRow: dragNodeRow,
			},
			sceneCX, sceneCY, sceneCZ, sceneRadius, nil)
	})

	// Wire "t"'s (the recipient's) own dedicated node stream so its GotDragMsg field is
	// independently observable, mirroring TestGesturePathPropagatesUIStateToMoverStream.
	bufT := wireNodeStream(t, md, "f")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	md.Start(ctx)

	// Mirror commitDragStart: reset the per-drag scope, then drag "t" (the target-only
	// node) a few pointer-move ticks, same as RootMove being called once per move event.
	md.resetAbcDrag()
	target := vec3{X: 500, Y: 100, Z: 200}
	if ok := md.RootMove("t", target); !ok {
		t.Fatalf("RootMove(t) returned false")
	}

	waitForAbcDragTickDrained(t, md)
	if md.ui.abcDragCount < 1 {
		t.Fatalf("abcDragCount after dragging target-only node t = %d, want >= 1", md.ui.abcDragCount)
	}
	if count, ok := lastViewFrameAbcDragCount(viewBuf.Bytes()); !ok || count < 1 {
		t.Fatalf("view frame AbcDragCount = %v (ok=%v), want >= 1", count, ok)
	}
	// f (the neighbor / recipient) should also show gotDragMsg on its OWN stream.
	waitForNodeDragMsg(t, bufT, func(got uint8, dA, dB, dC int32) bool { return got == 1 })
}

// TestAbcDragCountRealGestureTargetOnlyNode drives the SAME target-only-node topology
// through the REAL gesture entry point (md.HandleRawInput pointerdown/pointermove), not
// resetAbcDrag+RootMove directly — this is what actually runs live (HandleRawInput ->
// gestPointerDown's hitClassifiers["node"] -> gestPointerMove's commitEdges/applyAction
// table, gesture_graph.go). If this passes while the live editor still shows count==0,
// the gap is somewhere the raw-input dispatch differs from this in-process call (e.g. a
// raycast hit-kind resolution issue specific to a single-neighbor target node), not in
// resetAbcDrag/RootMove/requantizeLocalPolars/sendAbcDragTick themselves.
func TestAbcDragCountRealGestureTargetOnlyNode(t *testing.T) {
	root := writeFT(t)
	tr := T.NewWithSinkHook(nil, nil)

	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}

	viewBuf := &uiPubLockedBuf{}
	md.SetViewStream(viewBuf, func(tick uint32,
		camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
		sceneTori, scenePoles, nodePoles, selSpherePoles, handholds, labelsGlobal, overlaysVis, doubleLinks uint8,
		abcDragCount uint32,
		dragNodeRow int32,
		sceneCX, sceneCY, sceneCZ, sceneRadius float32,
		events []wire.RowEvent,
	) []byte {
		return B.BuildViewStreamFrame(tick, camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi,
			B.OverlayRow{
				SceneTori: sceneTori, ScenePoles: scenePoles, NodePoles: nodePoles,
				SelSpherePoles: selSpherePoles, Handholds: handholds, LabelsGlobal: labelsGlobal,
				OverlaysVis: overlaysVis, DoubleLinks: doubleLinks, AbcDragCount: abcDragCount,
				DragNodeRow: dragNodeRow,
			},
			sceneCX, sceneCY, sceneCZ, sceneRadius, nil)
	})
	bufF := wireNodeStream(t, md, "f")

	rowT, ok := md.NodeRowFor("t")
	if !ok {
		t.Fatalf("no NODE-ROW for t")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	md.Start(ctx)

	rectFov := func(ev rawInputMsg) rawInputMsg {
		ev.RectLeft, ev.RectTop, ev.RectWidth, ev.RectHeight = 0, 0, 800, 600
		ev.Fov = 50
		return ev
	}

	down := rectFov(rawInputMsg{Kind: "pointerdown", X: 400, Y: 300, Button: 0,
		Hit: rawHit{Kind: "node", NodeRow: int(rowT), PortRow: -1, EdgeRow: -1}})
	md.HandleRawInput(down, nil, tr)
	if md.ui.gest.dragNode != "t" {
		t.Fatalf("gestPointerDown: dragNode = %q, want \"t\" (hitClassifiers[\"node\"] must resolve+set it)", md.ui.gest.dragNode)
	}

	for i, dx := range []float64{20, 40, 60, 80, 100} {
		mv := rectFov(rawInputMsg{Kind: "pointermove", X: 400 + dx, Y: 300 + float64(i)*5})
		md.HandleRawInput(mv, nil, tr)
	}
	if md.ui.gest.phase != gestDragging {
		t.Fatalf("gesture phase after moves = %v, want gestDragging (commit never fired)", md.ui.gest.phase)
	}

	waitForAbcDragTickDrained(t, md)
	if md.ui.abcDragCount < 1 {
		t.Fatalf("REPRO: abcDragCount after real-gesture-drag of target-only node t = %d, want >= 1", md.ui.abcDragCount)
	}
	if count, ok := lastViewFrameAbcDragCount(viewBuf.Bytes()); !ok || count < 1 {
		t.Fatalf("REPRO: view frame AbcDragCount = %v (ok=%v), want >= 1", count, ok)
	}
	waitForNodeDragMsg(t, bufF, func(got uint8, dA, dB, dC int32) bool { return got == 1 })
}
